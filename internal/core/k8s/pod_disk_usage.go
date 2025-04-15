package k8s

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// GetPodDiskUsage 获取Pod的实际磁盘使用情况
// 返回一个映射，键为卷的挂载路径，值为使用的字节数
func GetPodDiskUsage(client *kubernetes.Clientset, config *rest.Config, namespace, podName string) (map[string]int64, error) {
	// 存储卷使用情况的映射，键为路径，值为使用的字节数
	diskUsage := make(map[string]int64)

	// 获取Pod详情，检查是否可以执行命令
	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod详情失败: %v", err)
	}

	// 只有Running状态的Pod才能执行命令
	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("Pod未处于Running状态，无法获取磁盘使用情况")
	}

	// 检查Pod中是否有就绪的容器
	var containerName string
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			containerName = containerStatus.Name
			break
		}
	}

	if containerName == "" {
		// 如果没有就绪的容器，使用第一个容器
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			return nil, fmt.Errorf("Pod中没有可用的容器")
		}
	}

	// 检查容器中是否存在df命令
	// 首先尝试检查是否有df命令
	hasDfCmd, err := checkCommandExists(client, config, namespace, podName, containerName, "df")
	if err != nil {
		return nil, fmt.Errorf("检查命令失败: %v", err)
	}

	// 定义变量以记录du命令是否可用
	hasDuCmd := false

	// 如果容器中有df命令，使用它获取磁盘使用情况
	if hasDfCmd {
		output, err := execCommandInContainer(client, config, namespace, podName, containerName, []string{"df", "-B1", "--output=target,used,size"})
		if err != nil {
			return nil, fmt.Errorf("执行df命令失败: %v", err)
		}

		// 解析df命令输出
		diskUsage, err = parseDfOutput(output)
		if err != nil {
			return nil, fmt.Errorf("解析df输出失败: %v", err)
		}
	} else {
		// 如果没有df命令，尝试使用du命令
		var duErr error
		hasDuCmd, duErr = checkCommandExists(client, config, namespace, podName, containerName, "du")
		if duErr != nil {
			return nil, fmt.Errorf("检查命令失败: %v", duErr)
		}

		if hasDuCmd {
			// 对于每个挂载的卷，尝试用du获取大小
			for _, volume := range pod.Spec.Volumes {
				if volume.PersistentVolumeClaim != nil {
					// 查找这个卷在容器中的挂载路径
					mountPath := findVolumeMountPath(pod, containerName, volume.Name)
					if mountPath != "" {
						output, err := execCommandInContainer(client, config, namespace, podName, containerName, []string{"du", "-sb", mountPath})
						if err != nil {
							continue // 跳过这个卷
						}

						// 解析du命令输出
						size, err := parseDuOutput(output)
						if err == nil && size > 0 {
							diskUsage[mountPath] = size
						}
					}
				}
			}
		} else {
			// 如果df和du命令都不可用，返回错误
			return nil, fmt.Errorf("容器中没有可用的磁盘使用查询命令")
		}
	}

	// 如果没有获取到任何磁盘使用情况，尝试查询emptyDir卷
	if len(diskUsage) == 0 {
		for _, volume := range pod.Spec.Volumes {
			if volume.EmptyDir != nil {
				// 查找这个卷在容器中的挂载路径
				mountPath := findVolumeMountPath(pod, containerName, volume.Name)
				if mountPath != "" && hasDuCmd {
					output, err := execCommandInContainer(client, config, namespace, podName, containerName, []string{"du", "-sb", mountPath})
					if err == nil {
						// 解析du命令输出
						size, err := parseDuOutput(output)
						if err == nil && size > 0 {
							diskUsage[mountPath] = size
						}
					}
				}
			}
		}
	}

	return diskUsage, nil
}

// 检查容器中是否存在特定命令
func checkCommandExists(client *kubernetes.Clientset, config *rest.Config, namespace, podName, containerName, command string) (bool, error) {
	// 尝试在容器中执行which命令检查是否存在特定命令
	_, err := execCommandInContainer(client, config, namespace, podName, containerName, []string{"which", command})
	if err != nil {
		// 尝试使用command -v检查
		_, err = execCommandInContainer(client, config, namespace, podName, containerName, []string{"command", "-v", command})
		if err != nil {
			// 最后尝试直接执行命令加--help看是否有响应
			_, err = execCommandInContainer(client, config, namespace, podName, containerName, []string{command, "--help"})
			return err == nil, nil
		}
	}
	return true, nil
}

// 在容器中执行命令
func execCommandInContainer(client *kubernetes.Clientset, config *rest.Config, namespace, podName, containerName string, command []string) (string, error) {
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("创建执行器失败: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// 查找卷在容器中的挂载路径
func findVolumeMountPath(pod *corev1.Pod, containerName, volumeName string) string {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == volumeName {
					return volumeMount.MountPath
				}
			}
		}
	}
	return ""
}

// 解析df命令的输出
func parseDfOutput(output string) (map[string]int64, error) {
	result := make(map[string]int64)
	lines := strings.Split(output, "\n")

	// 跳过标题行
	for i, line := range lines {
		if i == 0 {
			continue // 跳过标题行
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			mountPoint := fields[0]
			usedStr := fields[1]

			used, err := strconv.ParseInt(usedStr, 10, 64)
			if err != nil {
				continue // 跳过解析错误的行
			}

			result[mountPoint] = used
		}
	}

	return result, nil
}

// 解析du命令的输出
func parseDuOutput(output string) (int64, error) {
	// du -sb 输出格式: "大小 路径"
	re := regexp.MustCompile(`^(\d+)\s+.*$`)
	match := re.FindStringSubmatch(strings.TrimSpace(output))
	if len(match) >= 2 {
		return strconv.ParseInt(match[1], 10, 64)
	}
	return 0, fmt.Errorf("无法解析du输出: %s", output)
}
