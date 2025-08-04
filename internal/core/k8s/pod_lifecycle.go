package k8s

import (
	"context"
	"fmt"
	"time"

	"kube-tide/internal/utils/logger"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// PodLifecycleService Pod生命周期管理服务
type PodLifecycleService struct {
	clientManager *ClientManager
}

// NewPodLifecycleService 创建Pod生命周期服务
func NewPodLifecycleService(clientManager *ClientManager) *PodLifecycleService {
	return &PodLifecycleService{
		clientManager: clientManager,
	}
}

// PodLifecycleAction Pod生命周期动作类型
type PodLifecycleAction string

const (
	ActionStart   PodLifecycleAction = "start"
	ActionStop    PodLifecycleAction = "stop"
	ActionRestart PodLifecycleAction = "restart"
	ActionPause   PodLifecycleAction = "pause"
	ActionResume  PodLifecycleAction = "resume"
)

// PodLifecycleError Pod生命周期错误类型
type PodLifecycleError struct {
	Type     string                 `json:"type"`
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Original error                  `json:"-"`
}

func (e *PodLifecycleError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Original)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// 错误类型常量
const (
	ErrorTypeValidation = "validation"
	ErrorTypeResource   = "resource"
	ErrorTypeTimeout    = "timeout"
	ErrorTypePermission = "permission"
	ErrorTypeController = "controller"
	ErrorTypeKubernetes = "kubernetes"
	ErrorTypeNetwork    = "network"
	ErrorTypeInternal   = "internal"
)

// 错误代码常量
const (
	ErrorCodePodNotFound        = "POD_NOT_FOUND"
	ErrorCodeInvalidAction      = "INVALID_ACTION"
	ErrorCodeOperationTimeout   = "OPERATION_TIMEOUT"
	ErrorCodeControllerNotFound = "CONTROLLER_NOT_FOUND"
	ErrorCodeKubernetesAPI      = "KUBERNETES_API_ERROR"
	ErrorCodePodNotReady        = "POD_NOT_READY"
	ErrorCodePermissionDenied   = "PERMISSION_DENIED"
	ErrorCodeClusterConnection  = "CLUSTER_CONNECTION"
)

// NewPodLifecycleError 创建Pod生命周期错误
func NewPodLifecycleError(errorType, code, message string, original error, details map[string]interface{}) *PodLifecycleError {
	return &PodLifecycleError{
		Type:     errorType,
		Code:     code,
		Message:  message,
		Details:  details,
		Original: original,
	}
}

// PodLifecycleRequest Pod生命周期操作请求
type PodLifecycleRequest struct {
	Action      PodLifecycleAction `json:"action"`
	GracePeriod *int64             `json:"gracePeriod,omitempty"`
	Force       bool               `json:"force,omitempty"`
	WaitTimeout time.Duration      `json:"waitTimeout,omitempty"`
}

// PodLifecycleResponse Pod生命周期操作响应
type PodLifecycleResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	PodStatus PodLifecycleStatus `json:"podStatus"`
	Duration  time.Duration      `json:"duration"`
}

// PodLifecycleStatus Pod生命周期状态
type PodLifecycleStatus struct {
	Phase             corev1.PodPhase            `json:"phase"`
	Ready             bool                       `json:"ready"`
	ContainerStatuses []ContainerLifecycleStatus `json:"containerStatuses"`
	StartTime         *metav1.Time               `json:"startTime,omitempty"`
	RestartCount      int32                      `json:"restartCount"`
}

// ContainerLifecycleStatus 容器生命周期状态
type ContainerLifecycleStatus struct {
	Name         string         `json:"name"`
	Ready        bool           `json:"ready"`
	RestartCount int32          `json:"restartCount"`
	State        ContainerState `json:"state"`
	LastState    ContainerState `json:"lastState"`
}

// ContainerState 容器状态
type ContainerState struct {
	Running    *ContainerStateRunning    `json:"running,omitempty"`
	Waiting    *ContainerStateWaiting    `json:"waiting,omitempty"`
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

// ContainerStateRunning 容器运行状态
type ContainerStateRunning struct {
	StartedAt metav1.Time `json:"startedAt"`
}

// ContainerStateWaiting 容器等待状态
type ContainerStateWaiting struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// ContainerStateTerminated 容器终止状态
type ContainerStateTerminated struct {
	ExitCode   int32       `json:"exitCode"`
	Signal     int32       `json:"signal"`
	Reason     string      `json:"reason"`
	Message    string      `json:"message"`
	StartedAt  metav1.Time `json:"startedAt"`
	FinishedAt metav1.Time `json:"finishedAt"`
}

// ManagePodLifecycle 管理Pod生命周期
func (s *PodLifecycleService) ManagePodLifecycle(ctx context.Context, clusterName, namespace, podName string, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	startTime := time.Now()

	// 记录开始操作的日志
	logger.Infof("开始Pod生命周期操作 - 集群: %s, 命名空间: %s, Pod: %s, 操作: %s",
		clusterName, namespace, podName, request.Action)

	// 验证操作类型
	validActions := map[PodLifecycleAction]bool{
		ActionStart: true, ActionStop: true, ActionRestart: true,
		ActionPause: true, ActionResume: true,
	}
	if !validActions[request.Action] {
		return nil, NewPodLifecycleError(
			ErrorTypeValidation,
			ErrorCodeInvalidAction,
			fmt.Sprintf("不支持的生命周期动作: %s", request.Action),
			nil,
			map[string]interface{}{
				"action":       request.Action,
				"validActions": []string{"start", "stop", "restart", "pause", "resume"},
			},
		)
	}

	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, NewPodLifecycleError(
			ErrorTypeNetwork,
			ErrorCodeClusterConnection,
			"获取集群客户端失败",
			err,
			map[string]interface{}{
				"cluster": clusterName,
			},
		)
	}

	// 获取当前Pod状态
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, NewPodLifecycleError(
				ErrorTypeResource,
				ErrorCodePodNotFound,
				fmt.Sprintf("Pod '%s' 在命名空间 '%s' 中不存在", podName, namespace),
				err,
				map[string]interface{}{
					"cluster":   clusterName,
					"namespace": namespace,
					"pod":       podName,
				},
			)
		}
		return nil, NewPodLifecycleError(
			ErrorTypeKubernetes,
			"pod_get_failed",
			"获取Pod详情失败",
			err,
			map[string]interface{}{
				"cluster":   clusterName,
				"namespace": namespace,
				"pod":       podName,
			},
		)
	}

	var response *PodLifecycleResponse

	switch request.Action {
	case ActionStart:
		response, err = s.startPod(ctx, client, namespace, podName, pod, request)
	case ActionStop:
		response, err = s.stopPod(ctx, client, namespace, podName, pod, request)
	case ActionRestart:
		response, err = s.restartPod(ctx, client, namespace, podName, pod, request)
	case ActionPause:
		response, err = s.pausePod(ctx, client, namespace, podName, pod, request)
	case ActionResume:
		response, err = s.resumePod(ctx, client, namespace, podName, pod, request)
	default:
		return nil, NewPodLifecycleError(
			ErrorTypeValidation,
			ErrorCodeInvalidAction,
			fmt.Sprintf("不支持的生命周期动作: %s", request.Action),
			nil,
			map[string]interface{}{
				"action": request.Action,
			},
		)
	}

	if err != nil {
		// 如果已经是PodLifecycleError，直接返回
		if lifecycleErr, ok := err.(*PodLifecycleError); ok {
			return nil, lifecycleErr
		}
		// 否则包装为内部错误
		return nil, NewPodLifecycleError(
			ErrorTypeInternal,
			"operation_failed",
			fmt.Sprintf("执行生命周期操作失败: %v", err),
			err,
			map[string]interface{}{
				"action": request.Action,
			},
		)
	}

	response.Duration = time.Since(startTime)

	// 记录操作完成的日志
	logger.Infof("Pod生命周期操作完成 - 集群: %s, 命名空间: %s, Pod: %s, 操作: %s, 耗时: %v, 成功: %t",
		clusterName, namespace, podName, request.Action, response.Duration, response.Success)

	return response, nil
}

// startPod 启动Pod（适用于已停止的Pod）
func (s *PodLifecycleService) startPod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	// 如果Pod已经在运行，返回当前状态
	if pod.Status.Phase == corev1.PodRunning {
		status := s.getPodLifecycleStatus(pod)
		return &PodLifecycleResponse{
			Success:   true,
			Message:   "Pod已经在运行中",
			PodStatus: status,
		}, nil
	}

	// 如果Pod处于Failed或Succeeded状态，需要重新创建
	if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
		return s.recreatePod(ctx, client, namespace, podName, pod, request)
	}

	// 等待Pod启动
	return s.waitForPodReady(ctx, client, namespace, podName, request.WaitTimeout)
}

// stopPod 停止Pod
func (s *PodLifecycleService) stopPod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
		status := s.getPodLifecycleStatus(pod)
		return &PodLifecycleResponse{
			Success:   true,
			Message:   "Pod已经停止",
			PodStatus: status,
		}, nil
	}

	// 设置删除选项
	deleteOptions := metav1.DeleteOptions{}
	if request.GracePeriod != nil {
		deleteOptions.GracePeriodSeconds = request.GracePeriod
	}
	if request.Force {
		gracePeriod := int64(0)
		deleteOptions.GracePeriodSeconds = &gracePeriod
	}

	// 删除Pod
	err := client.CoreV1().Pods(namespace).Delete(ctx, podName, deleteOptions)
	if err != nil {
		return nil, fmt.Errorf("停止Pod失败: %w", err)
	}

	// 等待Pod终止
	return s.waitForPodTerminated(ctx, client, namespace, podName, request.WaitTimeout)
}

// restartPod 重启Pod
func (s *PodLifecycleService) restartPod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	// 先停止Pod
	stopRequest := &PodLifecycleRequest{
		Action:      ActionStop,
		GracePeriod: request.GracePeriod,
		Force:       request.Force,
		WaitTimeout: request.WaitTimeout / 2,
	}

	_, err := s.stopPod(ctx, client, namespace, podName, pod, stopRequest)
	if err != nil {
		return nil, fmt.Errorf("重启Pod时停止操作失败: %w", err)
	}

	// 重新创建Pod
	return s.recreatePod(ctx, client, namespace, podName, pod, request)
}

// pausePod 暂停Pod（通过设置副本数为0，适用于有控制器的Pod）
func (s *PodLifecycleService) pausePod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	// 检查Pod是否由控制器管理
	if len(pod.OwnerReferences) == 0 {
		return nil, fmt.Errorf("Pod没有控制器，无法执行暂停操作")
	}

	// 根据控制器类型执行暂停操作
	for _, owner := range pod.OwnerReferences {
		switch owner.Kind {
		case "Deployment":
			return s.pauseDeployment(ctx, client, namespace, owner.Name)
		case "StatefulSet":
			return s.pauseStatefulSet(ctx, client, namespace, owner.Name)
		case "ReplicaSet":
			return s.pauseReplicaSet(ctx, client, namespace, owner.Name)
		}
	}

	return nil, fmt.Errorf("不支持的控制器类型，无法执行暂停操作")
}

// resumePod 恢复Pod
func (s *PodLifecycleService) resumePod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	// 实现恢复逻辑（与暂停相反）
	if len(pod.OwnerReferences) == 0 {
		return nil, fmt.Errorf("Pod没有控制器，无法执行恢复操作")
	}

	for _, owner := range pod.OwnerReferences {
		switch owner.Kind {
		case "Deployment":
			return s.resumeDeployment(ctx, client, namespace, owner.Name)
		case "StatefulSet":
			return s.resumeStatefulSet(ctx, client, namespace, owner.Name)
		case "ReplicaSet":
			return s.resumeReplicaSet(ctx, client, namespace, owner.Name)
		}
	}

	return nil, fmt.Errorf("不支持的控制器类型，无法执行恢复操作")
}

// recreatePod 重新创建Pod
func (s *PodLifecycleService) recreatePod(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod, request *PodLifecycleRequest) (*PodLifecycleResponse, error) {
	// 创建新的Pod规格
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   namespace,
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
		},
		Spec: pod.Spec,
	}

	// 清理不应该被复制的字段
	newPod.ObjectMeta.ResourceVersion = ""
	newPod.ObjectMeta.UID = ""
	newPod.ObjectMeta.CreationTimestamp = metav1.Time{}
	newPod.Spec.NodeName = ""

	// 创建新Pod
	createdPod, err := client.CoreV1().Pods(namespace).Create(ctx, newPod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("重新创建Pod失败: %w", err)
	}

	// 等待Pod就绪
	return s.waitForPodReady(ctx, client, namespace, createdPod.Name, request.WaitTimeout)
}

// waitForPodReady 等待Pod就绪
func (s *PodLifecycleService) waitForPodReady(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, timeout time.Duration) (*PodLifecycleResponse, error) {
	if timeout == 0 {
		timeout = 5 * time.Minute // 默认5分钟超时
	}

	var finalPod *corev1.Pod

	waitErr := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		finalPod = pod

		// 检查Pod是否就绪
		if pod.Status.Phase == corev1.PodRunning {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}

		// 检查是否失败
		if pod.Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("Pod进入失败状态: %s", pod.Status.Message)
		}

		return false, nil
	})

	if waitErr != nil {
		if finalPod != nil {
			status := s.getPodLifecycleStatus(finalPod)
			return &PodLifecycleResponse{
				Success:   false,
				Message:   fmt.Sprintf("等待Pod就绪超时: %v", waitErr),
				PodStatus: status,
			}, waitErr
		}
		return nil, waitErr
	}

	status := s.getPodLifecycleStatus(finalPod)
	return &PodLifecycleResponse{
		Success:   true,
		Message:   "Pod已成功启动并就绪",
		PodStatus: status,
	}, nil
}

// waitForPodTerminated 等待Pod终止
func (s *PodLifecycleService) waitForPodTerminated(ctx context.Context, client *kubernetes.Clientset, namespace, podName string, timeout time.Duration) (*PodLifecycleResponse, error) {
	if timeout == 0 {
		timeout = 2 * time.Minute // 默认2分钟超时
	}

	waitErr := wait.PollImmediate(1*time.Second, timeout, func() (bool, error) {
		_, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			// Pod不存在，说明已经被删除
			return true, nil
		}
		return false, nil
	})

	if waitErr != nil {
		return &PodLifecycleResponse{
			Success: false,
			Message: fmt.Sprintf("等待Pod终止超时: %v", waitErr),
		}, waitErr
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "Pod已成功停止",
	}, nil
}

// getPodLifecycleStatus 获取Pod生命周期状态
func (s *PodLifecycleService) getPodLifecycleStatus(pod *corev1.Pod) PodLifecycleStatus {
	status := PodLifecycleStatus{
		Phase:     pod.Status.Phase,
		Ready:     false,
		StartTime: pod.Status.StartTime,
	}

	// 检查Pod是否就绪
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			status.Ready = condition.Status == corev1.ConditionTrue
			break
		}
	}

	// 获取容器状态
	status.ContainerStatuses = make([]ContainerLifecycleStatus, len(pod.Status.ContainerStatuses))
	for i, cs := range pod.Status.ContainerStatuses {
		containerStatus := ContainerLifecycleStatus{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
		}

		// 转换容器状态
		if cs.State.Running != nil {
			containerStatus.State.Running = &ContainerStateRunning{
				StartedAt: cs.State.Running.StartedAt,
			}
		}
		if cs.State.Waiting != nil {
			containerStatus.State.Waiting = &ContainerStateWaiting{
				Reason:  cs.State.Waiting.Reason,
				Message: cs.State.Waiting.Message,
			}
		}
		if cs.State.Terminated != nil {
			containerStatus.State.Terminated = &ContainerStateTerminated{
				ExitCode:   cs.State.Terminated.ExitCode,
				Signal:     cs.State.Terminated.Signal,
				Reason:     cs.State.Terminated.Reason,
				Message:    cs.State.Terminated.Message,
				StartedAt:  cs.State.Terminated.StartedAt,
				FinishedAt: cs.State.Terminated.FinishedAt,
			}
		}

		// 转换上一个状态
		if cs.LastTerminationState.Terminated != nil {
			containerStatus.LastState.Terminated = &ContainerStateTerminated{
				ExitCode:   cs.LastTerminationState.Terminated.ExitCode,
				Signal:     cs.LastTerminationState.Terminated.Signal,
				Reason:     cs.LastTerminationState.Terminated.Reason,
				Message:    cs.LastTerminationState.Terminated.Message,
				StartedAt:  cs.LastTerminationState.Terminated.StartedAt,
				FinishedAt: cs.LastTerminationState.Terminated.FinishedAt,
			}
		}

		status.ContainerStatuses[i] = containerStatus
		status.RestartCount += cs.RestartCount
	}

	return status
}

// GetPodLifecycleStatus 获取Pod生命周期状态（公共方法）
func (s *PodLifecycleService) GetPodLifecycleStatus(pod *corev1.Pod) PodLifecycleStatus {
	return s.getPodLifecycleStatus(pod)
}

// 控制器暂停/恢复方法
func (s *PodLifecycleService) pauseDeployment(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment失败: %w", err)
	}

	// 暂停Deployment（设置副本数为0）
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
		return &PodLifecycleResponse{
			Success: true,
			Message: "Deployment已经暂停",
		}, nil
	}

	// 保存原始副本数
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	if deployment.Spec.Replicas != nil {
		deployment.Annotations["kube-tide/original-replicas"] = fmt.Sprintf("%d", *deployment.Spec.Replicas)
	}

	// 设置副本数为0
	replicas := int32(0)
	deployment.Spec.Replicas = &replicas

	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("暂停Deployment失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "Deployment已成功暂停",
	}, nil
}

func (s *PodLifecycleService) resumeDeployment(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment失败: %w", err)
	}

	// 检查是否有保存的原始副本数
	originalReplicas := "1" // 默认值
	if deployment.Annotations != nil {
		if orig, exists := deployment.Annotations["kube-tide/original-replicas"]; exists {
			originalReplicas = orig
		}
	}

	// 解析副本数
	var replicas int32 = 1
	if _, err := fmt.Sscanf(originalReplicas, "%d", &replicas); err != nil {
		replicas = 1
	}

	deployment.Spec.Replicas = &replicas

	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("恢复Deployment失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "Deployment已成功恢复",
	}, nil
}

func (s *PodLifecycleService) pauseStatefulSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取StatefulSet失败: %w", err)
	}

	// 保存原始副本数并设置为0
	if statefulSet.Annotations == nil {
		statefulSet.Annotations = make(map[string]string)
	}
	if statefulSet.Spec.Replicas != nil {
		statefulSet.Annotations["kube-tide/original-replicas"] = fmt.Sprintf("%d", *statefulSet.Spec.Replicas)
	}

	replicas := int32(0)
	statefulSet.Spec.Replicas = &replicas

	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("暂停StatefulSet失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "StatefulSet已成功暂停",
	}, nil
}

func (s *PodLifecycleService) resumeStatefulSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取StatefulSet失败: %w", err)
	}

	// 恢复原始副本数
	originalReplicas := "1"
	if statefulSet.Annotations != nil {
		if orig, exists := statefulSet.Annotations["kube-tide/original-replicas"]; exists {
			originalReplicas = orig
		}
	}

	var replicas int32 = 1
	if _, err := fmt.Sscanf(originalReplicas, "%d", &replicas); err != nil {
		replicas = 1
	}

	statefulSet.Spec.Replicas = &replicas

	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("恢复StatefulSet失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "StatefulSet已成功恢复",
	}, nil
}

func (s *PodLifecycleService) pauseReplicaSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	replicaSet, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取ReplicaSet失败: %w", err)
	}

	// 保存原始副本数并设置为0
	if replicaSet.Annotations == nil {
		replicaSet.Annotations = make(map[string]string)
	}
	if replicaSet.Spec.Replicas != nil {
		replicaSet.Annotations["kube-tide/original-replicas"] = fmt.Sprintf("%d", *replicaSet.Spec.Replicas)
	}

	replicas := int32(0)
	replicaSet.Spec.Replicas = &replicas

	_, err = client.AppsV1().ReplicaSets(namespace).Update(ctx, replicaSet, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("暂停ReplicaSet失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "ReplicaSet已成功暂停",
	}, nil
}

func (s *PodLifecycleService) resumeReplicaSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*PodLifecycleResponse, error) {
	replicaSet, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取ReplicaSet失败: %w", err)
	}

	// 恢复原始副本数
	originalReplicas := "1"
	if replicaSet.Annotations != nil {
		if orig, exists := replicaSet.Annotations["kube-tide/original-replicas"]; exists {
			originalReplicas = orig
		}
	}

	var replicas int32 = 1
	if _, err := fmt.Sscanf(originalReplicas, "%d", &replicas); err != nil {
		replicas = 1
	}

	replicaSet.Spec.Replicas = &replicas

	_, err = client.AppsV1().ReplicaSets(namespace).Update(ctx, replicaSet, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("恢复ReplicaSet失败: %w", err)
	}

	return &PodLifecycleResponse{
		Success: true,
		Message: "ReplicaSet已成功恢复",
	}, nil
}

// GetPodLifecycleHistory 获取Pod生命周期历史
func (s *PodLifecycleService) GetPodLifecycleHistory(ctx context.Context, clusterName, namespace, podName string) ([]PodLifecycleEvent, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %w", err)
	}

	// 获取Pod相关事件
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", podName, namespace)
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Pod事件失败: %w", err)
	}

	// 转换为生命周期事件
	lifecycleEvents := make([]PodLifecycleEvent, 0, len(events.Items))
	for _, event := range events.Items {
		lifecycleEvent := PodLifecycleEvent{
			Timestamp: event.LastTimestamp.Time,
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Source:    fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host),
		}
		lifecycleEvents = append(lifecycleEvents, lifecycleEvent)
	}

	return lifecycleEvents, nil
}

// PodLifecycleEvent Pod生命周期事件
type PodLifecycleEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}
