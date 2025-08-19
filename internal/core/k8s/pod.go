package k8s

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// IsNotFoundError 判断错误是否为"资源未找到"错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误字符串中是否包含典型的"not found"信息
	errMsg := err.Error()
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "NotFound") ||
		strings.Contains(errMsg, "no such") ||
		errors.IsNotFound(err) // k8s.io/apimachinery/pkg/api/errors 提供的函数
}

// PodService Pod服务
type PodService struct {
	clientManager    *ClientManager
	metricsService   *PodMetricsService
	lifecycleService *PodLifecycleService
}

// NewPodService 创建Pod服务
func NewPodService(clientManager *ClientManager) *PodService {
	metricsService := NewPodMetricsService(clientManager)
	lifecycleService := NewPodLifecycleService(clientManager)
	return &PodService{
		clientManager:    clientManager,
		metricsService:   metricsService,
		lifecycleService: lifecycleService,
	}
}

// GetPods 获取所有命名空间的Pod列表
func (s *PodService) GetPods(ctx context.Context, clusterName string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	podList, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	return podList.Items, nil
}

// GetPodsByNamespace 获取指定命名空间的Pod列表
func (s *PodService) GetPodsByNamespace(ctx context.Context, clusterName, namespace string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间 %s 的Pod列表失败: %w", namespace, err)
	}

	return podList.Items, nil
}

// GetPodsBySelector 根据标签选择器获取Pod列表
func (s *PodService) GetPodsBySelector(ctx context.Context, clusterName, namespace string, selector map[string]string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 构建标签选择器
	labelSelector := ""
	for k, v := range selector {
		if labelSelector != "" {
			labelSelector += ","
		}
		labelSelector += fmt.Sprintf("%s=%s", k, v)
	}

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	return podList.Items, nil
}

// GetPodDetails 获取Pod详情
func (s *PodService) GetPodDetails(ctx context.Context, clusterName, namespace, podName string) (*corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod详情失败: %w", err)
	}

	return pod, nil
}

// GetPodLogs 获取Pod日志
func (s *PodService) GetPodLogs(ctx context.Context, clusterName, namespace, podName, containerName string, tailLines int64) (string, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return "", err
	}

	podLogOptions := &corev1.PodLogOptions{
		Container: containerName,
	}

	if tailLines > 0 {
		podLogOptions.TailLines = &tailLines
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("获取Pod日志流失败: %w", err)
	}
	defer podLogs.Close()

	buf := new([]byte)
	*buf, err = io.ReadAll(podLogs)
	if err != nil {
		return "", fmt.Errorf("读取Pod日志失败: %w", err)
	}

	return string(*buf), nil
}

// StreamPodLogs 获取Pod日志流，适用于实时日志
func (s *PodService) StreamPodLogs(ctx context.Context, clusterName, namespace, podName, containerName string, tailLines int64, follow bool) (io.ReadCloser, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	podLogOptions := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    follow, // 是否持续跟踪日志
	}

	if tailLines > 0 {
		podLogOptions.TailLines = &tailLines
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
	return req.Stream(ctx)
}

// GetPodStatus 获取Pod状态
func (s *PodService) GetPodStatus(pod *corev1.Pod) string {
	// 检查Pod是否处于删除状态（存在deletion timestamp）
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	// 如果容器状态不为空，可能需要更详细的状态判断
	if len(pod.Status.ContainerStatuses) > 0 {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			// 检查容器是否处于特殊状态
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason != "" {
				return containerStatus.State.Waiting.Reason
			}
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.Reason != "" {
				return containerStatus.State.Terminated.Reason
			}
		}
	}

	// 默认返回Phase状态
	return string(pod.Status.Phase)
}

// DeletePod 删除Pod
func (s *PodService) DeletePod(ctx context.Context, clusterName, namespace, podName string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	deleteOptions := metav1.DeleteOptions{}
	err = client.CoreV1().Pods(namespace).Delete(ctx, podName, deleteOptions)
	if err != nil {
		return fmt.Errorf("删除Pod失败: %w", err)
	}

	return nil
}

// GetPodExecExecutor 获取Pod终端服务
func (s *PodService) GetPodExecExecutor(
	ctx context.Context,
	clusterName string,
	namespace string,
	podName string,
	containerName string,
	command []string,
	stdin bool,
	stdout bool,
	stderr bool,
	tty bool,
) (remotecommand.Executor, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取客户端失败: %w", err)
	}

	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		TTY:       tty,
	}, scheme.ParameterCodec)

	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("创建SPDY执行器失败: %w", err)
	}

	return executor, nil
}

// ExecToPod 在Pod中执行命令
func (s *PodService) ExecToPod(clusterName, namespace, podName, containerName string, terminal remotecommand.TerminalSizeQueue) error {
	ctx := context.Background()

	// 默认终端命令
	command := []string{"/bin/sh", "-c", "if [ -x /bin/bash ]; then /bin/bash; elif [ -x /bin/sh ]; then /bin/sh; else echo 'No shell available'; exit 1; fi"}

	// 获取执行器
	executor, err := s.GetPodExecExecutor(
		ctx,
		clusterName,
		namespace,
		podName,
		containerName,
		command,
		true, // stdin
		true, // stdout
		true, // stderr
		true, // tty
	)
	if err != nil {
		return fmt.Errorf("创建终端执行器失败: %w", err)
	}

	// 确保terminal实现了必要的接口
	if _, ok := terminal.(remotecommand.TerminalSizeQueue); !ok {
		return fmt.Errorf("终端未实现必要的接口")
	}

	// 确保terminal同时实现了io.Reader和io.Writer
	stdinReader, ok := terminal.(io.Reader)
	if !ok {
		return fmt.Errorf("终端未实现io.Reader接口")
	}

	stdoutWriter, ok := terminal.(io.Writer)
	if !ok {
		return fmt.Errorf("终端未实现io.Writer接口")
	}

	// 启动SPDY流
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdinReader,
		Stdout:            stdoutWriter,
		Stderr:            stdoutWriter,
		Tty:               true,
		TerminalSizeQueue: terminal,
	})
	if err != nil {
		return fmt.Errorf("执行终端命令失败: %w", err)
	}

	return nil
}

// CheckPodExists 检查Pod是否存在
func (s *PodService) CheckPodExists(ctx context.Context, clusterName, namespace, podName string) (*corev1.Pod, bool, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, false, err
	}

	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("检查Pod是否存在失败: %w", err)
	}

	return pod, true, nil
}

// GetPodEvents 获取Pod相关的事件
func (s *PodService) GetPodEvents(ctx context.Context, clusterName, namespace, podName string) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// pod, err := s.GetPodDetails(ctx, clusterName, namespace, podName)
	// if (err != nil) {
	// 	return nil, fmt.Errorf("获取Pod详情失败: %w", err)
	// }

	// 设置字段选择器，筛选与指定Pod相关的事件
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", podName, namespace)

	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Pod事件列表失败: %w", err)
	}

	// 按照最后时间戳降序排序，确保最新事件在前
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})

	return events.Items, nil
}

// GetPodMetrics 获取Pod的CPU和内存监控指标（使用缓存服务）
func (s *PodService) GetPodMetrics(ctx context.Context, clusterName, namespace, podName string) (*PodMetrics, error) {
	// 直接使用指标服务获取Pod指标（会优先从缓存获取，缓存中没有再从API获取）
	return s.metricsService.GetPodMetrics(ctx, clusterName, namespace, podName)
}

// RestartPolicyConfig 重启策略配置结构
type RestartPolicyConfig struct {
	RestartPolicy string `json:"restartPolicy"`
}

// UpdatePodRestartPolicy 更新Pod的重启策略
// 实际实现：由于Kubernetes不允许直接修改Pod的重启策略，此方法提供Pod重新创建的功能
func (s *PodService) UpdatePodRestartPolicy(ctx context.Context, clusterName, namespace, podName, restartPolicy string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 验证重启策略的有效性
	if !isValidRestartPolicy(restartPolicy) {
		return fmt.Errorf("invalid restart policy: %s. Valid values are: Always, OnFailure, Never", restartPolicy)
	}

	// 获取原Pod
	originalPod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Pod: %w", err)
	}

	// 检查Pod状态
	if originalPod.DeletionTimestamp != nil {
		return fmt.Errorf("cannot update restart policy for pod in deletion state")
	}

	// 检查当前重启策略是否已经相同
	if string(originalPod.Spec.RestartPolicy) == restartPolicy {
		return fmt.Errorf("restart policy is already set to %s", restartPolicy)
	}

	// 创建新的Pod定义
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        originalPod.Name + "-new",
			Namespace:   originalPod.Namespace,
			Labels:      originalPod.Labels,
			Annotations: originalPod.Annotations,
		},
		Spec: originalPod.Spec,
	}

	// 更新重启策略
	newPod.Spec.RestartPolicy = corev1.RestartPolicy(restartPolicy)

	// 清理不应该被复制的字段
	newPod.ObjectMeta.ResourceVersion = ""
	newPod.ObjectMeta.UID = ""
	newPod.ObjectMeta.CreationTimestamp = metav1.Time{}
	newPod.Spec.NodeName = ""

	// 删除状态相关的annotation
	if newPod.ObjectMeta.Annotations != nil {
		delete(newPod.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	}

	// 创建新Pod
	_, err = client.CoreV1().Pods(namespace).Create(ctx, newPod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create new Pod: %w", err)
	}

	// 删除原Pod（可选，用户确认后执行）
	// 这里我们返回成功，让用户手动决定是否删除原Pod
	return nil
}

// RecreatePodWithRestartPolicy 重新创建Pod并设置新的重启策略
func (s *PodService) RecreatePodWithRestartPolicy(ctx context.Context, clusterName, namespace, podName, restartPolicy string, deleteOriginal bool) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %w", err)
	}

	// 验证重启策略的有效性
	if !isValidRestartPolicy(restartPolicy) {
		return fmt.Errorf("无效的重启策略: %s. 有效值为: Always, OnFailure, Never", restartPolicy)
	}

	// 获取原Pod
	originalPod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取Pod失败: %w", err)
	}

	// 检查Pod是否由控制器管理
	if len(originalPod.OwnerReferences) > 0 {
		return fmt.Errorf("无法修改由控制器管理的Pod的重启策略。Pod由 %s/%s 控制器管理",
			originalPod.OwnerReferences[0].Kind, originalPod.OwnerReferences[0].Name)
	}

	// 检查当前重启策略是否已经相同
	if string(originalPod.Spec.RestartPolicy) == restartPolicy {
		return fmt.Errorf("重启策略已经设置为 %s，无需修改", restartPolicy)
	}

	// 创建新的Pod定义，使用相同的名称（如果要删除原Pod）
	newPodName := podName
	if !deleteOriginal {
		// Kubernetes resource names must be lowercase and <= 63 characters
		safePolicy := strings.ToLower(restartPolicy)
		maxNameLen := 63
		baseName := podName + "-" + safePolicy
		if len(baseName) > maxNameLen {
			baseName = baseName[:maxNameLen]
		}
		newPodName = baseName
	}

	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        newPodName,
			Namespace:   originalPod.Namespace,
			Labels:      originalPod.Labels,
			Annotations: originalPod.Annotations,
		},
		Spec: originalPod.Spec,
	}

	// 更新重启策略
	newPod.Spec.RestartPolicy = corev1.RestartPolicy(restartPolicy)

	// 清理不应该被复制的字段
	newPod.ObjectMeta.ResourceVersion = ""
	newPod.ObjectMeta.UID = ""
	newPod.ObjectMeta.CreationTimestamp = metav1.Time{}
	newPod.Spec.NodeName = ""

	// 删除状态相关的annotation
	if newPod.ObjectMeta.Annotations != nil {
		delete(newPod.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	}

	// 如果要删除原Pod，先删除再创建
	if deleteOriginal {
		err = client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("删除原始Pod失败: %w", err)
		}

		// 等待Pod被完全删除
		for i := 0; i < 30; i++ {
			_, err = client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				break
			}
			if i == 29 {
				return fmt.Errorf("等待原始Pod删除超时")
			}
			time.Sleep(1 * time.Second)
		}
	}

	// 创建新Pod
	_, err = client.CoreV1().Pods(namespace).Create(ctx, newPod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建新Pod失败: %w", err)
	}

	return nil
}

// GetPodRestartPolicy 获取Pod的重启策略
func (s *PodService) GetPodRestartPolicy(ctx context.Context, clusterName, namespace, podName string) (string, error) {
	pod, err := s.GetPodDetails(ctx, clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	return string(pod.Spec.RestartPolicy), nil
}

// isValidRestartPolicy 检查重启策略是否有效
func isValidRestartPolicy(policy string) bool {
	validPolicies := []string{"Always", "OnFailure", "Never"}
	for _, validPolicy := range validPolicies {
		if policy == validPolicy {
			return true
		}
	}
	return false
}

// ManagePodLifecycle 管理Pod生命周期
func (s *PodService) ManagePodLifecycle(ctx context.Context, clusterName, namespace, podName string, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	return s.lifecycleService.ManagePodLifecycle(ctx, clusterName, namespace, podName, request)
}

// GetPodLifecycleHistory 获取Pod生命周期历史
func (s *PodService) GetPodLifecycleHistory(ctx context.Context, clusterName, namespace, podName string) ([]PodLifecycleEvent, error) {
	return s.lifecycleService.GetPodLifecycleHistory(ctx, clusterName, namespace, podName)
}

// GetClientManager 获取客户端管理器（供生命周期服务使用）
func (s *PodService) GetClientManager() *ClientManager {
	return s.clientManager
}
