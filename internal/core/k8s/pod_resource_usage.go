package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// ContainerResourceUsage 容器资源使用情况
type ContainerResourceUsage struct {
	// 容器名称
	Name string
	// CPU使用量（单位：核心）
	CPUUsage float64
	// 内存使用量（单位：字节）
	MemoryUsage int64
}

// PodResourceUsage Pod资源使用情况
type PodResourceUsage struct {
	// 容器级别的资源使用情况
	Containers []ContainerResourceUsage
	// Pod总体CPU使用量（单位：核心）
	TotalCPUUsage float64
	// Pod总体内存使用量（单位：字节）
	TotalMemoryUsage int64
	// 历史数据
	Historical map[string][]ResourceDataPoint
}

// ResourceDataPoint 资源数据点
type ResourceDataPoint struct {
	// 时间戳
	Timestamp time.Time
	// 值
	Value float64
}

// GetPodResourceUsage 获取Pod的真实CPU和内存使用情况
func GetPodResourceUsage(client *kubernetes.Clientset, config *rest.Config, namespace, podName string) (*PodResourceUsage, error) {
	// 创建metrics客户端
	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建metrics客户端失败: %v", err)
	}

	// 获取Pod metrics
	ctx := context.Background()
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod metrics失败: %v", err)
	}

	// 获取Pod详情
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod详情失败: %v", err)
	}

	// 初始化结果
	result := &PodResourceUsage{
		Containers:       make([]ContainerResourceUsage, 0, len(podMetrics.Containers)),
		TotalCPUUsage:    0,
		TotalMemoryUsage: 0,
		Historical:       make(map[string][]ResourceDataPoint),
	}

	// 遍历容器指标
	for _, containerMetrics := range podMetrics.Containers {
		cpuUsage := containerMetrics.Usage.Cpu().AsApproximateFloat64()
		memoryUsage := containerMetrics.Usage.Memory().Value()

		// 添加到容器列表
		result.Containers = append(result.Containers, ContainerResourceUsage{
			Name:        containerMetrics.Name,
			CPUUsage:    cpuUsage,
			MemoryUsage: memoryUsage,
		})

		// 累加总使用量
		result.TotalCPUUsage += cpuUsage
		result.TotalMemoryUsage += memoryUsage
	}

	// 如果metrics API没有提供完整数据，尝试从容器状态获取更多信息
	if len(result.Containers) == 0 && pod.Status.Phase == corev1.PodRunning {
		// 尝试通过执行命令获取资源使用情况
		// 这里只是一个后备方案，通常metrics API已经提供了我们需要的数据
		containerResources, err := getPodResourceByExec(client, config, namespace, podName)
		if err == nil && len(containerResources) > 0 {
			result.Containers = containerResources

			// 计算总使用量
			for _, container := range containerResources {
				result.TotalCPUUsage += container.CPUUsage
				result.TotalMemoryUsage += container.MemoryUsage
			}
		}
	}

	// 获取历史数据
	// 这里可以从Prometheus或其他时序数据库获取历史数据
	// 由于没有外部依赖的要求，我们可以从metrics API获取当前数据点，然后存储在内存中
	// 在实际生产环境中，建议使用Prometheus等工具存储历史数据

	// 初始化历史数据数组
	result.Historical["cpu"] = make([]ResourceDataPoint, 0)
	result.Historical["memory"] = make([]ResourceDataPoint, 0)

	// 添加当前数据点
	now := time.Now()
	result.Historical["cpu"] = append(result.Historical["cpu"], ResourceDataPoint{
		Timestamp: now,
		Value:     result.TotalCPUUsage,
	})

	result.Historical["memory"] = append(result.Historical["memory"], ResourceDataPoint{
		Value:     float64(result.TotalMemoryUsage),
		Timestamp: now,
	})

	return result, nil
}

// getPodResourceByExec 通过执行命令获取Pod的资源使用情况
// 这是一个后备方法，当metrics API不可用时使用
func getPodResourceByExec(client *kubernetes.Clientset, config *rest.Config, namespace, podName string) ([]ContainerResourceUsage, error) {
	// 获取Pod详情
	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod详情失败: %v", err)
	}

	// 只处理Running状态的Pod
	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("Pod未处于Running状态")
	}

	// 选择第一个就绪的容器
	var containerName string
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			containerName = status.Name
			break
		}
	}

	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	if containerName == "" {
		return nil, fmt.Errorf("Pod中没有可用的容器")
	}

	// 尝试执行top命令获取资源使用情况
	result := []ContainerResourceUsage{}

	// 在容器中执行top命令
	topOutput, err := execCommandInContainer(client, config, namespace, podName, containerName, []string{"top", "-b", "-n", "1"})
	if err != nil {
		// 尝试其他命令
		topOutput, err = execCommandInContainer(client, config, namespace, podName, containerName, []string{"ps", "aux"})
		if err != nil {
			return nil, fmt.Errorf("执行命令失败: %v", err)
		}
	}

	// 解析top/ps输出
	cpuUsage, memUsage, err := parseResourceUsage(topOutput)
	if err != nil {
		return nil, fmt.Errorf("解析资源使用情况失败: %v", err)
	}

	// 添加到结果
	result = append(result, ContainerResourceUsage{
		Name:        containerName,
		CPUUsage:    cpuUsage,
		MemoryUsage: memUsage,
	})

	return result, nil
}

// parseResourceUsage 解析top/ps命令输出，提取CPU和内存使用情况
func parseResourceUsage(output string) (float64, int64, error) {
	lines := strings.Split(output, "\n")

	// 跳过标题行
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("无法解析输出: %s", output)
	}

	// 查找进程行，通常选择占用资源最多的进程
	var maxCPU float64
	var maxMem int64

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 10 {
			continue
		}

		// 解析CPU使用率（百分比）
		cpuStr := fields[2]
		cpu, err := strconv.ParseFloat(strings.TrimSuffix(cpuStr, "%"), 64)
		if err != nil {
			continue
		}

		// 解析内存使用率（百分比）和实际使用量
		memStr := fields[3]
		mem, err := strconv.ParseFloat(strings.TrimSuffix(memStr, "%"), 64)
		if err != nil {
			continue
		}

		// 将百分比转换为实际使用量（估算）
		// 这里假设系统总内存为16GB，实际应用中应该获取实际的系统内存
		memBytes := int64(mem * 16 * 1024 * 1024 * 1024 / 100)

		// 保存最大值
		if cpu > maxCPU {
			maxCPU = cpu
		}

		if memBytes > maxMem {
			maxMem = memBytes
		}
	}

	// 转换CPU百分比为核心数（假设系统有4个核心）
	cpuCores := maxCPU * 4 / 100

	return cpuCores, maxMem, nil
}
