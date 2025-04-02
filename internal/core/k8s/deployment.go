package k8s

import (
	"context"
	"fmt"
	"kube-tide/internal/utils/logger"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeploymentService 提供与Kubernetes Deployments交互的服务
type DeploymentService struct {
	clientManager *ClientManager
}

// NewDeploymentService 创建一个新的DeploymentService实例
func NewDeploymentService(clientManager *ClientManager) *DeploymentService {
	return &DeploymentService{
		clientManager: clientManager,
	}
}

// DeploymentInfo 包含Deployment的基本信息
type DeploymentInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Replicas       int32             `json:"replicas"`
	ReadyReplicas  int32             `json:"readyReplicas"`
	Strategy       string            `json:"strategy"`
	CreationTime   time.Time         `json:"creationTime"`
	Labels         map[string]string `json:"labels"`
	Selector       map[string]string `json:"selector"`
	ContainerCount int               `json:"containerCount"`
	Images         []string          `json:"images"`
}

// DeploymentDetails 包含Deployment的详细信息
type DeploymentDetails struct {
	DeploymentInfo
	Annotations             map[string]string     `json:"annotations"`
	Containers              []ContainerInfo       `json:"containers"`
	Conditions              []DeploymentCondition `json:"conditions"`
	RevisionHistoryLimit    *int32                `json:"revisionHistoryLimit,omitempty"`
	ProgressDeadlineSeconds *int32                `json:"progressDeadlineSeconds,omitempty"`
	MinReadySeconds         int32                 `json:"minReadySeconds"`
	Paused                  bool                  `json:"paused"`
}

// DeploymentCondition 包含Deployment的状态条件
type DeploymentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastUpdateTime     time.Time `json:"lastUpdateTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// 环境变量引用的配置映射或密钥
type ConfigMapKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// 环境变量引用的密钥
type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// 环境变量值来源
type EnvVarSource struct {
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	SecretKeyRef    *SecretKeySelector    `json:"secretKeyRef,omitempty"`
}

// EnvVar 定义环境变量
type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// ResourceRequirements 定义资源需求和限制
type ResourceRequirements struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}

// ListDeployments 获取所有Deployment列表
func (ds *DeploymentService) ListDeployments(clusterName string) ([]DeploymentInfo, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %v", err)
	}

	deployments, err := client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployments列表失败: %v", err)
	}

	return ds.convertDeploymentList(deployments.Items), nil
}

// ListDeploymentsByNamespace 获取指定命名空间的Deployment列表
func (ds *DeploymentService) ListDeploymentsByNamespace(clusterName, namespace string) ([]DeploymentInfo, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %v", err)
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间 %s 的Deployments列表失败: %v", namespace, err)
	}

	return ds.convertDeploymentList(deployments.Items), nil
}

// GetDeploymentDetails 获取单个Deployment的详细信息
func (ds *DeploymentService) GetDeploymentDetails(clusterName, namespace, name string) (*DeploymentDetails, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %v", err)
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment详情失败: %v", err)
	}

	details := &DeploymentDetails{
		DeploymentInfo:          ds.convertDeployment(deployment),
		Annotations:             deployment.Annotations,
		RevisionHistoryLimit:    deployment.Spec.RevisionHistoryLimit,
		ProgressDeadlineSeconds: deployment.Spec.ProgressDeadlineSeconds,
		MinReadySeconds:         deployment.Spec.MinReadySeconds,
		Paused:                  deployment.Spec.Paused,
	}

	// 添加容器信息
	for _, container := range deployment.Spec.Template.Spec.Containers {
		details.Containers = append(details.Containers, ContainerInfo{
			Name:      container.Name,
			Image:     container.Image,
			Resources: container.Resources,
			Ports:     container.Ports,
			Env:       container.Env,
		})
	}

	// 添加状态条件
	for _, condition := range deployment.Status.Conditions {
		details.Conditions = append(details.Conditions, DeploymentCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastUpdateTime:     condition.LastUpdateTime.Time,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	return details, nil
}

// GetDeploymentEvents 获取Deployment相关的事件
func (s *DeploymentService) GetDeploymentEvents(ctx context.Context, clusterName, namespace, deploymentName string) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	deployment, err := s.GetDeploymentDetails(clusterName, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取Deployment详情失败: %w", err)
	}

	if deployment == nil {
		return nil, fmt.Errorf("deployment %s/%s不存在", namespace, deploymentName)
	}

	// 设置字段选择器，筛选与指定Deployment相关的事件
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Deployment",
		deploymentName, namespace)

	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment事件列表失败: %w", err)
	}

	// 按照最后时间戳降序排序，确保最新事件在前
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})

	return events.Items, nil
}

// GetDeploymentPodEvents 获取Deployment关联的所有Pod的事件
func (s *DeploymentService) GetDeploymentPodEvents(ctx context.Context, clusterName, namespace, deploymentName string) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 查询与Deployment关联的所有Pod
	// 使用Deployment的selector来过滤Pod
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment失败: %w", err)
	}

	// 获取Deployment的选择器
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析标签选择器失败: %w", err)
	}

	// 查询符合选择器的Pod
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("查询Pod列表失败: %w", err)
	}

	var allEvents []corev1.Event

	// 获取每个Pod的事件
	for _, pod := range pods.Items {
		fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod",
			pod.Name, namespace)

		events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("获取Pod %s 的事件列表失败: %w", pod.Name, err)
		}

		allEvents = append(allEvents, events.Items...)
	}

	// 按照最后时间戳降序排序，确保最新事件在前
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].LastTimestamp.After(allEvents[j].LastTimestamp.Time)
	})

	return allEvents, nil
}

// GetAllDeploymentEvents 获取Deployment及其关联的ReplicaSet和Pod的所有事件
func (s *DeploymentService) GetAllDeploymentEvents(ctx context.Context, clusterName, namespace, deploymentName string) (map[string][]corev1.Event, error) {
	// 获取Deployment自身事件
	deploymentEvents, err := s.GetDeploymentEvents(ctx, clusterName, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取Deployment事件失败: %w", err)
	}

	// 获取相关ReplicaSet事件
	replicaSetEvents, err := s.GetReplicaSetEvents(ctx, clusterName, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取ReplicaSet事件失败: %w", err)
	}

	// 获取相关Pod事件
	podEvents, err := s.GetDeploymentPodEvents(ctx, clusterName, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("获取Pod事件失败: %w", err)
	}

	// 将所有事件分类返回
	result := map[string][]corev1.Event{
		"deployment": deploymentEvents,
		"replicaSet": replicaSetEvents,
		"pod":        podEvents,
	}

	return result, nil
}

// GetReplicaSetEvents 获取ReplicaSet相关的事件
func (s *DeploymentService) GetReplicaSetEvents(ctx context.Context, clusterName, namespace, deploymentName string) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 查询与Deployment关联的所有ReplicaSets
	labelSelector := fmt.Sprintf("app=%s", deploymentName)
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取ReplicaSet列表失败: %w", err)
	}

	var allEvents []corev1.Event

	// 获取每个ReplicaSet的事件
	for _, rs := range rsList.Items {
		// 检查ReplicaSet是否属于该Deployment（通过所有者引用）
		if isOwnedByDeployment(rs.OwnerReferences, deploymentName) {
			fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=ReplicaSet",
				rs.Name, namespace)

			events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
				FieldSelector: fieldSelector,
			})
			if err != nil {
				return nil, fmt.Errorf("获取ReplicaSet %s 的事件列表失败: %w", rs.Name, err)
			}

			allEvents = append(allEvents, events.Items...)
		}
	}

	// 按照最后时间戳降序排序，确保最新事件在前
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].LastTimestamp.After(allEvents[j].LastTimestamp.Time)
	})

	return allEvents, nil
}

// 检查资源是否被指定的Deployment拥有
func isOwnedByDeployment(ownerRefs []metav1.OwnerReference, deploymentName string) bool {
	for _, owner := range ownerRefs {
		if owner.Kind == "Deployment" && owner.Name == deploymentName {
			return true
		}
	}
	return false
}

// convertDeploymentList 将K8s API的Deployment列表转换为自定义格式
func (ds *DeploymentService) convertDeploymentList(deployments []appsv1.Deployment) []DeploymentInfo {
	result := make([]DeploymentInfo, 0, len(deployments))
	for _, deployment := range deployments {
		result = append(result, ds.convertDeployment(&deployment))
	}
	return result
}

// convertDeployment 将K8s API的Deployment转换为自定义格式
func (ds *DeploymentService) convertDeployment(deployment *appsv1.Deployment) DeploymentInfo {
	images := make([]string, 0)
	containerCount := len(deployment.Spec.Template.Spec.Containers)

	for _, container := range deployment.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	var strategy string
	switch {
	case deployment.Spec.Strategy.Type == appsv1.RollingUpdateDeploymentStrategyType:
		strategy = "RollingUpdate"
	case deployment.Spec.Strategy.Type == appsv1.RecreateDeploymentStrategyType:
		strategy = "Recreate"
	default:
		strategy = string(deployment.Spec.Strategy.Type)
	}

	return DeploymentInfo{
		Name:           deployment.Name,
		Namespace:      deployment.Namespace,
		Replicas:       *deployment.Spec.Replicas,
		ReadyReplicas:  deployment.Status.ReadyReplicas,
		Strategy:       strategy,
		CreationTime:   deployment.CreationTimestamp.Time,
		Labels:         deployment.Labels,
		Selector:       deployment.Spec.Selector.MatchLabels,
		ContainerCount: containerCount,
		Images:         images,
	}
}

// ScaleDeployment 调整Deployment的副本数
func (ds *DeploymentService) ScaleDeployment(clusterName, namespace, name string, replicas int32) error {
	logger.Info("调整Deployment副本数:", clusterName, namespace, name, replicas)
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %v", err)
	}

	// 获取当前Deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取Deployment失败: %v", err)
	}

	// 修改副本数
	deployment.Spec.Replicas = &replicas

	// 更新Deployment
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新Deployment副本数失败: %v", err)
	}

	logger.Info("成功调整Deployment副本数:", clusterName, namespace, name, replicas)
	return nil
}

// RestartDeployment 重启Deployment（通过添加重启注解实现）
func (ds *DeploymentService) RestartDeployment(clusterName, namespace, name string) error {
	logger.Info("重启Deployment:", clusterName, namespace, name)
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %v", err)
	}

	// 获取当前Deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.Err("获取Deployment失败", err)
		return fmt.Errorf("获取Deployment失败: %v", err)
	}

	// 添加或更新重启注解
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新Deployment
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		logger.Err("更新Deployment失败", err)
		return fmt.Errorf("重启Deployment失败: %v", err)
	}

	logger.Info("成功重启Deployment:", clusterName, namespace, name)
	return nil
}

// UpdateDeploymentRequest 定义更新Deployment的请求结构
type UpdateDeploymentRequest struct {
	Replicas             *int32                          `json:"replicas,omitempty"`
	Image                map[string]string               `json:"image,omitempty"`     // 容器名称到镜像的映射
	Env                  map[string][]EnvVar             `json:"env,omitempty"`       // 容器名称到环境变量的映射
	Resources            map[string]ResourceRequirements `json:"resources,omitempty"` // 容器名称到资源需求的映射
	Labels               map[string]string               `json:"labels,omitempty"`
	Annotations          map[string]string               `json:"annotations,omitempty"`
	Strategy             *DeploymentStrategy             `json:"strategy,omitempty"`
	MinReadySeconds      *int32                          `json:"minReadySeconds,omitempty"`
	RevisionHistoryLimit *int32                          `json:"revisionHistoryLimit,omitempty"`
	Paused               *bool                           `json:"paused,omitempty"`
	// 添加健康检查配置
	LivenessProbe  map[string]*Probe        `json:"livenessProbe,omitempty"`  // 容器名称到存活探针的映射
	ReadinessProbe map[string]*Probe        `json:"readinessProbe,omitempty"` // 容器名称到就绪探针的映射
	StartupProbe   map[string]*Probe        `json:"startupProbe,omitempty"`   // 容器名称到启动探针的映射
	Volumes        []Volume                 `json:"volumes,omitempty"`
	VolumeMounts   map[string][]VolumeMount `json:"volumeMounts,omitempty"` // 容器名称到卷挂载点的映射
}

// DeploymentStrategy 定义Deployment的更新策略
type DeploymentStrategy struct {
	Type          string                   `json:"type"` // RollingUpdate 或 Recreate
	RollingUpdate *RollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

// RollingUpdateDeployment 定义滚动更新参数
type RollingUpdateDeployment struct {
	MaxSurge       *string `json:"maxSurge,omitempty"`
	MaxUnavailable *string `json:"maxUnavailable,omitempty"`
}

// Probe 定义健康检查探针的配置
type Probe struct {
	HTTPGet             *HTTPGetAction `json:"httpGet,omitempty"`
	TCPSocket           *TCPSocket     `json:"tcpSocket,omitempty"`
	Exec                *ExecAction    `json:"exec,omitempty"`
	InitialDelaySeconds int32          `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds      int32          `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       int32          `json:"periodSeconds,omitempty"`
	SuccessThreshold    int32          `json:"successThreshold,omitempty"`
	FailureThreshold    int32          `json:"failureThreshold,omitempty"`
}

// HTTPGetAction 定义HTTP GET检查的配置
type HTTPGetAction struct {
	Path        string       `json:"path"`
	Port        int32        `json:"port"`
	Host        string       `json:"host,omitempty"`
	Scheme      string       `json:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

// TCPSocket 定义TCP Socket检查的配置
type TCPSocket struct {
	Port int32  `json:"port"`
	Host string `json:"host,omitempty"`
}

// ExecAction 定义命令行检查的配置
type ExecAction struct {
	Command []string `json:"command"`
}

// HTTPHeader 定义HTTP头部
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// VolumeMount 定义容器卷挂载点
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// Volume 定义卷的配置
type Volume struct {
	Name                  string                 `json:"name"`
	ConfigMap             *ConfigMapVolumeSource `json:"configMap,omitempty"`
	Secret                *SecretVolumeSource    `json:"secret,omitempty"`
	PersistentVolumeClaim *PVCVolumeSource       `json:"persistentVolumeClaim,omitempty"`
	EmptyDir              *EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	HostPath              *HostPathVolumeSource  `json:"hostPath,omitempty"`
}

// ConfigMapVolumeSource 定义ConfigMap卷源
type ConfigMapVolumeSource struct {
	Name        string      `json:"name"`
	Items       []KeyToPath `json:"items,omitempty"`
	DefaultMode *int32      `json:"defaultMode,omitempty"`
}

// SecretVolumeSource 定义Secret卷源
type SecretVolumeSource struct {
	SecretName  string      `json:"secretName"`
	Items       []KeyToPath `json:"items,omitempty"`
	DefaultMode *int32      `json:"defaultMode,omitempty"`
}

// PVCVolumeSource 定义PVC卷源
type PVCVolumeSource struct {
	ClaimName string `json:"claimName"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// EmptyDirVolumeSource 定义EmptyDir卷源
type EmptyDirVolumeSource struct {
	Medium    string `json:"medium,omitempty"`
	SizeLimit string `json:"sizeLimit,omitempty"`
}

// HostPathVolumeSource 定义主机路径卷源
type HostPathVolumeSource struct {
	Path string `json:"path"`
	Type string `json:"type,omitempty"`
}

// KeyToPath 定义键到路径的映射
type KeyToPath struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Mode *int32 `json:"mode,omitempty"`
}

// CreateDeploymentRequest 定义创建Deployment的请求结构
type CreateDeploymentRequest struct {
	// 基本信息
	Name        string            `json:"name" binding:"required"`
	Replicas    *int32            `json:"replicas,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`

	// 部署选项
	Strategy                *DeploymentStrategy `json:"strategy,omitempty"`
	MinReadySeconds         *int32              `json:"minReadySeconds,omitempty"`
	RevisionHistoryLimit    *int32              `json:"revisionHistoryLimit,omitempty"`
	ProgressDeadlineSeconds *int32              `json:"progressDeadlineSeconds,omitempty"`
	Paused                  *bool               `json:"paused,omitempty"`

	// Pod相关配置
	Containers         []ContainerSpec   `json:"containers" binding:"required"`
	InitContainers     []ContainerSpec   `json:"initContainers,omitempty"`
	Volumes            []Volume          `json:"volumes,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	Tolerations        []Toleration      `json:"tolerations,omitempty"`
	Affinity           *Affinity         `json:"affinity,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	HostNetwork        *bool             `json:"hostNetwork,omitempty"`
	DNSPolicy          string            `json:"dnsPolicy,omitempty"`
}

// ContainerSpec 定义容器规格
type ContainerSpec struct {
	Name            string               `json:"name" binding:"required"`
	Image           string               `json:"image" binding:"required"`
	Command         []string             `json:"command,omitempty"`
	Args            []string             `json:"args,omitempty"`
	WorkingDir      string               `json:"workingDir,omitempty"`
	Ports           []ContainerPort      `json:"ports,omitempty"`
	Env             []EnvVar             `json:"env,omitempty"`
	Resources       ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts    []VolumeMount        `json:"volumeMounts,omitempty"`
	LivenessProbe   *Probe               `json:"livenessProbe,omitempty"`
	ReadinessProbe  *Probe               `json:"readinessProbe,omitempty"`
	StartupProbe    *Probe               `json:"startupProbe,omitempty"`
	ImagePullPolicy string               `json:"imagePullPolicy,omitempty"`
	SecurityContext *SecurityContext     `json:"securityContext,omitempty"`
}

// ContainerPort 定义容器端口
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort" binding:"required"`
	HostPort      int32  `json:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

// Toleration 定义容忍
type Toleration struct {
	Key               string `json:"key,omitempty"`
	Operator          string `json:"operator,omitempty"`
	Value             string `json:"value,omitempty"`
	Effect            string `json:"effect,omitempty"`
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty"`
}

// Affinity 定义亲和性
type Affinity struct {
	NodeAffinity    *NodeAffinity    `json:"nodeAffinity,omitempty"`
	PodAffinity     *PodAffinity     `json:"podAffinity,omitempty"`
	PodAntiAffinity *PodAntiAffinity `json:"podAntiAffinity,omitempty"`
}

// NodeAffinity 定义节点亲和性
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  *NodeSelector             `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// NodeSelector 定义节点选择器
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms,omitempty"`
}

// NodeSelectorTerm 定义节点选择器条件
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty"`
	MatchFields      []NodeSelectorRequirement `json:"matchFields,omitempty"`
}

// NodeSelectorRequirement 定义节点选择器要求
type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

// PreferredSchedulingTerm 定义优选调度条件
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

// PodAffinity 定义Pod亲和性
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAntiAffinity 定义Pod反亲和性
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAffinityTerm 定义Pod亲和性条件
type PodAffinityTerm struct {
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Namespaces    []string          `json:"namespaces,omitempty"`
	TopologyKey   string            `json:"topologyKey"`
}

// WeightedPodAffinityTerm 定义带权重的Pod亲和性条件
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm"`
}

// SecurityContext 定义容器安全上下文
type SecurityContext struct {
	Capabilities             *Capabilities `json:"capabilities,omitempty"`
	Privileged               *bool         `json:"privileged,omitempty"`
	RunAsUser                *int64        `json:"runAsUser,omitempty"`
	RunAsGroup               *int64        `json:"runAsGroup,omitempty"`
	RunAsNonRoot             *bool         `json:"runAsNonRoot,omitempty"`
	ReadOnlyRootFilesystem   *bool         `json:"readOnlyRootFilesystem,omitempty"`
	AllowPrivilegeEscalation *bool         `json:"allowPrivilegeEscalation,omitempty"`
}

// Capabilities 定义容器权能
type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// UpdateDeployment 更新Deployment配置
func (ds *DeploymentService) UpdateDeployment(clusterName, namespace, name string, update UpdateDeploymentRequest) error {
	logger.Info("更新Deployment:", clusterName, namespace, name)
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %v", err)
	}

	// 获取当前Deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.Err("获取Deployment失败", err)
		return fmt.Errorf("获取Deployment失败: %v", err)
	}

	// 更新副本数
	if update.Replicas != nil {
		deployment.Spec.Replicas = update.Replicas
	}

	// 更新容器镜像
	if update.Image != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if newImage, exists := update.Image[container.Name]; exists {
				deployment.Spec.Template.Spec.Containers[i].Image = newImage
			}
		}
	}

	// 更新容器环境变量
	if update.Env != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if envs, exists := update.Env[container.Name]; exists {
				// 转换为k8s环境变量格式
				k8sEnvs := make([]corev1.EnvVar, 0, len(envs))
				for _, env := range envs {
					k8sEnv := corev1.EnvVar{
						Name:  env.Name,
						Value: env.Value,
					}
					if env.ValueFrom != nil {
						// 处理环境变量引用
						k8sEnv.ValueFrom = &corev1.EnvVarSource{}
						if env.ValueFrom.ConfigMapKeyRef != nil {
							k8sEnv.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: env.ValueFrom.ConfigMapKeyRef.Name,
								},
								Key: env.ValueFrom.ConfigMapKeyRef.Key,
							}
						}
						if env.ValueFrom.SecretKeyRef != nil {
							k8sEnv.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: env.ValueFrom.SecretKeyRef.Name,
								},
								Key: env.ValueFrom.SecretKeyRef.Key,
							}
						}
					}
					k8sEnvs = append(k8sEnvs, k8sEnv)
				}
				deployment.Spec.Template.Spec.Containers[i].Env = k8sEnvs
			}
		}
	}

	// 更新容器资源需求
	if update.Resources != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if res, exists := update.Resources[container.Name]; exists {
				container := &deployment.Spec.Template.Spec.Containers[i]

				// 创建资源限制
				if len(res.Limits) > 0 {
					limits := corev1.ResourceList{}
					for k, v := range res.Limits {
						quantity, err := resource.ParseQuantity(v)
						if err != nil {
							// 解析资源限制值失败
							return fmt.Errorf("解析资源限制值失败 %s=%s: %v", k, v, err)
						}
						limits[corev1.ResourceName(k)] = quantity
					}
					container.Resources.Limits = limits
				}

				// 创建资源请求
				if len(res.Requests) > 0 {
					requests := corev1.ResourceList{}
					for k, v := range res.Requests {
						quantity, err := resource.ParseQuantity(v)
						if err != nil {
							return fmt.Errorf("解析资源请求值失败 %s=%s: %v", k, v, err)
						}
						requests[corev1.ResourceName(k)] = quantity
					}
					container.Resources.Requests = requests
				}
			}
		}
	}

	// 更新标签
	if update.Labels != nil {
		if deployment.Labels == nil {
			deployment.Labels = make(map[string]string)
		}
		for k, v := range update.Labels {
			deployment.Labels[k] = v
		}
		// 同时更新Pod模板的标签
		if deployment.Spec.Template.Labels == nil {
			deployment.Spec.Template.Labels = make(map[string]string)
		}
		for k, v := range update.Labels {
			deployment.Spec.Template.Labels[k] = v
		}
	}

	// 更新注解
	if update.Annotations != nil {
		if deployment.Annotations == nil {
			deployment.Annotations = make(map[string]string)
		}
		for k, v := range update.Annotations {
			deployment.Annotations[k] = v
		}
	}

	// 更新部署策略
	if update.Strategy != nil {
		switch update.Strategy.Type {
		case "RollingUpdate":
			deployment.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
			if update.Strategy.RollingUpdate != nil {
				rollingUpdate := &appsv1.RollingUpdateDeployment{}
				if update.Strategy.RollingUpdate.MaxSurge != nil {
					maxSurge := intstr.FromString(*update.Strategy.RollingUpdate.MaxSurge)
					rollingUpdate.MaxSurge = &maxSurge
				}
				if update.Strategy.RollingUpdate.MaxUnavailable != nil {
					maxUnavailable := intstr.FromString(*update.Strategy.RollingUpdate.MaxUnavailable)
					rollingUpdate.MaxUnavailable = &maxUnavailable
				}
				deployment.Spec.Strategy.RollingUpdate = rollingUpdate
			}
		case "Recreate":
			deployment.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
			deployment.Spec.Strategy.RollingUpdate = nil
		}
	}

	// 更新健康检查探针
	if update.LivenessProbe != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if probe, exists := update.LivenessProbe[container.Name]; exists {
				deployment.Spec.Template.Spec.Containers[i].LivenessProbe = convertProbe(probe)
			}
		}
	}
	if update.ReadinessProbe != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if probe, exists := update.ReadinessProbe[container.Name]; exists {
				deployment.Spec.Template.Spec.Containers[i].ReadinessProbe = convertProbe(probe)
			}
		}
	}
	if update.StartupProbe != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if probe, exists := update.StartupProbe[container.Name]; exists {
				deployment.Spec.Template.Spec.Containers[i].StartupProbe = convertProbe(probe)
			}
		}
	}

	// 更新卷
	if update.Volumes != nil {
		k8sVolumes := make([]corev1.Volume, 0, len(update.Volumes))
		for _, volume := range update.Volumes {
			k8sVolume := corev1.Volume{
				Name: volume.Name,
			}
			if volume.ConfigMap != nil {
				k8sVolume.ConfigMap = &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: volume.ConfigMap.Name,
					},
					Items:       convertKeyToPath(volume.ConfigMap.Items),
					DefaultMode: volume.ConfigMap.DefaultMode,
				}
			}
			if volume.Secret != nil {
				k8sVolume.Secret = &corev1.SecretVolumeSource{
					SecretName:  volume.Secret.SecretName,
					Items:       convertKeyToPath(volume.Secret.Items),
					DefaultMode: volume.Secret.DefaultMode,
				}
			}
			if volume.PersistentVolumeClaim != nil {
				k8sVolume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: volume.PersistentVolumeClaim.ClaimName,
					ReadOnly:  volume.PersistentVolumeClaim.ReadOnly,
				}
			}
			if volume.EmptyDir != nil {
				sizeLimit := resource.MustParse(volume.EmptyDir.SizeLimit)
				k8sVolume.EmptyDir = &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMedium(volume.EmptyDir.Medium),
					SizeLimit: &sizeLimit,
				}
			}
			if volume.HostPath != nil {
				k8sVolume.HostPath = &corev1.HostPathVolumeSource{
					Path: volume.HostPath.Path,
					Type: (*corev1.HostPathType)(&volume.HostPath.Type),
				}
			}
			k8sVolumes = append(k8sVolumes, k8sVolume)
		}
		deployment.Spec.Template.Spec.Volumes = k8sVolumes
	}

	// 更新卷挂载
	if update.VolumeMounts != nil {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if mounts, exists := update.VolumeMounts[container.Name]; exists {
				k8sMounts := make([]corev1.VolumeMount, 0, len(mounts))
				for _, mount := range mounts {
					k8sMount := corev1.VolumeMount{
						Name:      mount.Name,
						MountPath: mount.MountPath,
						SubPath:   mount.SubPath,
						ReadOnly:  mount.ReadOnly,
					}
					k8sMounts = append(k8sMounts, k8sMount)
				}
				deployment.Spec.Template.Spec.Containers[i].VolumeMounts = k8sMounts
			}
		}
	}

	// 更新其他字段
	if update.MinReadySeconds != nil {
		deployment.Spec.MinReadySeconds = *update.MinReadySeconds
	}
	if update.RevisionHistoryLimit != nil {
		deployment.Spec.RevisionHistoryLimit = update.RevisionHistoryLimit
	}
	if update.Paused != nil {
		deployment.Spec.Paused = *update.Paused
	}

	// 更新Deployment
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新Deployment失败: %v", err)
	}

	logger.Info("成功更新Deployment:", clusterName, namespace, name)
	return nil
}

// CreateDeployment 创建新的Deployment
func (ds *DeploymentService) CreateDeployment(clusterName, namespace string, create CreateDeploymentRequest) (*DeploymentInfo, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %v", err)
	}

	// 设置默认副本数
	replicas := int32(1)
	if create.Replicas != nil {
		replicas = *create.Replicas
	}

	// 构建Deployment对象
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        create.Name,
			Namespace:   namespace,
			Labels:      create.Labels,
			Annotations: create.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": create.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": create.Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: create.ServiceAccountName,
					NodeSelector:       create.NodeSelector,
				},
			},
		},
	}

	// 如果提供了其他标签，合并到Pod模板中
	if create.Labels != nil {
		for k, v := range create.Labels {
			deployment.Spec.Template.ObjectMeta.Labels[k] = v
		}
	}

	// 设置部署策略
	if create.Strategy != nil {
		switch create.Strategy.Type {
		case "RollingUpdate":
			deployment.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
			if create.Strategy.RollingUpdate != nil {
				rollingUpdate := &appsv1.RollingUpdateDeployment{}
				if create.Strategy.RollingUpdate.MaxSurge != nil {
					maxSurge := intstr.FromString(*create.Strategy.RollingUpdate.MaxSurge)
					rollingUpdate.MaxSurge = &maxSurge
				}
				if create.Strategy.RollingUpdate.MaxUnavailable != nil {
					maxUnavailable := intstr.FromString(*create.Strategy.RollingUpdate.MaxUnavailable)
					rollingUpdate.MaxUnavailable = &maxUnavailable
				}
				deployment.Spec.Strategy.RollingUpdate = rollingUpdate
			}
		case "Recreate":
			deployment.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
		}
	} else {
		// 默认使用RollingUpdate策略
		deployment.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
	}

	// 设置其他部署参数
	if create.MinReadySeconds != nil {
		deployment.Spec.MinReadySeconds = *create.MinReadySeconds
	}
	if create.RevisionHistoryLimit != nil {
		deployment.Spec.RevisionHistoryLimit = create.RevisionHistoryLimit
	}
	if create.ProgressDeadlineSeconds != nil {
		deployment.Spec.ProgressDeadlineSeconds = create.ProgressDeadlineSeconds
	}
	if create.Paused != nil {
		deployment.Spec.Paused = *create.Paused
	}

	// 设置主容器
	for _, container := range create.Containers {
		k8sContainer := corev1.Container{
			Name:            container.Name,
			Image:           container.Image,
			Command:         container.Command,
			Args:            container.Args,
			WorkingDir:      container.WorkingDir,
			ImagePullPolicy: corev1.PullPolicy(container.ImagePullPolicy),
		}

		// 设置容器端口
		if len(container.Ports) > 0 {
			k8sPorts := make([]corev1.ContainerPort, 0, len(container.Ports))
			for _, port := range container.Ports {
				protocol := corev1.ProtocolTCP
				if port.Protocol != "" {
					protocol = corev1.Protocol(port.Protocol)
				}
				k8sPort := corev1.ContainerPort{
					Name:          port.Name,
					ContainerPort: port.ContainerPort,
					HostPort:      port.HostPort,
					Protocol:      protocol,
				}
				k8sPorts = append(k8sPorts, k8sPort)
			}
			k8sContainer.Ports = k8sPorts
		}

		// 设置环境变量
		if len(container.Env) > 0 {
			k8sEnvs := make([]corev1.EnvVar, 0, len(container.Env))
			for _, env := range container.Env {
				k8sEnv := corev1.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				}
				if env.ValueFrom != nil {
					k8sEnv.ValueFrom = &corev1.EnvVarSource{}
					if env.ValueFrom.ConfigMapKeyRef != nil {
						k8sEnv.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: env.ValueFrom.ConfigMapKeyRef.Name,
							},
							Key: env.ValueFrom.ConfigMapKeyRef.Key,
						}
					}
					if env.ValueFrom.SecretKeyRef != nil {
						k8sEnv.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: env.ValueFrom.SecretKeyRef.Name,
							},
							Key: env.ValueFrom.SecretKeyRef.Key,
						}
					}
				}
				k8sEnvs = append(k8sEnvs, k8sEnv)
			}
			k8sContainer.Env = k8sEnvs
		}

		// 设置资源需求
		if container.Resources.Limits != nil || container.Resources.Requests != nil {
			k8sContainer.Resources = corev1.ResourceRequirements{}

			if len(container.Resources.Limits) > 0 {
				limits := corev1.ResourceList{}
				for k, v := range container.Resources.Limits {
					quantity, err := resource.ParseQuantity(v)
					if err != nil {
						return nil, fmt.Errorf("解析资源限制值失败 %s=%s: %v", k, v, err)
					}
					limits[corev1.ResourceName(k)] = quantity
				}
				k8sContainer.Resources.Limits = limits
			}

			if len(container.Resources.Requests) > 0 {
				requests := corev1.ResourceList{}
				for k, v := range container.Resources.Requests {
					quantity, err := resource.ParseQuantity(v)
					if err != nil {
						return nil, fmt.Errorf("解析资源请求值失败 %s=%s: %v", k, v, err)
					}
					requests[corev1.ResourceName(k)] = quantity
				}
				k8sContainer.Resources.Requests = requests
			}
		}

		// 设置卷挂载
		if len(container.VolumeMounts) > 0 {
			k8sMounts := make([]corev1.VolumeMount, 0, len(container.VolumeMounts))
			for _, mount := range container.VolumeMounts {
				k8sMount := corev1.VolumeMount{
					Name:      mount.Name,
					MountPath: mount.MountPath,
					SubPath:   mount.SubPath,
					ReadOnly:  mount.ReadOnly,
				}
				k8sMounts = append(k8sMounts, k8sMount)
			}
			k8sContainer.VolumeMounts = k8sMounts
		}

		// 设置健康检查
		if container.LivenessProbe != nil {
			k8sContainer.LivenessProbe = convertProbe(container.LivenessProbe)
		}
		if container.ReadinessProbe != nil {
			k8sContainer.ReadinessProbe = convertProbe(container.ReadinessProbe)
		}
		if container.StartupProbe != nil {
			k8sContainer.StartupProbe = convertProbe(container.StartupProbe)
		}

		// 设置安全上下文
		if container.SecurityContext != nil {
			k8sSC := &corev1.SecurityContext{}
			if container.SecurityContext.Privileged != nil {
				k8sSC.Privileged = container.SecurityContext.Privileged
			}
			if container.SecurityContext.RunAsUser != nil {
				k8sSC.RunAsUser = container.SecurityContext.RunAsUser
			}
			if container.SecurityContext.RunAsGroup != nil {
				k8sSC.RunAsGroup = container.SecurityContext.RunAsGroup
			}
			if container.SecurityContext.RunAsNonRoot != nil {
				k8sSC.RunAsNonRoot = container.SecurityContext.RunAsNonRoot
			}
			if container.SecurityContext.ReadOnlyRootFilesystem != nil {
				k8sSC.ReadOnlyRootFilesystem = container.SecurityContext.ReadOnlyRootFilesystem
			}
			if container.SecurityContext.AllowPrivilegeEscalation != nil {
				k8sSC.AllowPrivilegeEscalation = container.SecurityContext.AllowPrivilegeEscalation
			}
			if container.SecurityContext.Capabilities != nil {
				capAdd := make([]corev1.Capability, 0, len(container.SecurityContext.Capabilities.Add))
				for _, cap := range container.SecurityContext.Capabilities.Add {
					capAdd = append(capAdd, corev1.Capability(cap))
				}

				capDrop := make([]corev1.Capability, 0, len(container.SecurityContext.Capabilities.Drop))
				for _, cap := range container.SecurityContext.Capabilities.Drop {
					capDrop = append(capDrop, corev1.Capability(cap))
				}

				k8sSC.Capabilities = &corev1.Capabilities{
					Add:  capAdd,
					Drop: capDrop,
				}
			}
			k8sContainer.SecurityContext = k8sSC
		}

		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, k8sContainer)
	}

	// 设置初始化容器
	if len(create.InitContainers) > 0 {
		k8sInitContainers := make([]corev1.Container, 0, len(create.InitContainers))
		for _, initContainer := range create.InitContainers {
			k8sContainer := corev1.Container{
				Name:            initContainer.Name,
				Image:           initContainer.Image,
				Command:         initContainer.Command,
				Args:            initContainer.Args,
				WorkingDir:      initContainer.WorkingDir,
				ImagePullPolicy: corev1.PullPolicy(initContainer.ImagePullPolicy),
			}

			// 设置环境变量
			if len(initContainer.Env) > 0 {
				k8sEnvs := make([]corev1.EnvVar, 0, len(initContainer.Env))
				for _, env := range initContainer.Env {
					k8sEnv := corev1.EnvVar{
						Name:  env.Name,
						Value: env.Value,
					}
					if env.ValueFrom != nil {
						k8sEnv.ValueFrom = &corev1.EnvVarSource{}
						if env.ValueFrom.ConfigMapKeyRef != nil {
							k8sEnv.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: env.ValueFrom.ConfigMapKeyRef.Name,
								},
								Key: env.ValueFrom.ConfigMapKeyRef.Key,
							}
						}
						if env.ValueFrom.SecretKeyRef != nil {
							k8sEnv.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: env.ValueFrom.SecretKeyRef.Name,
								},
								Key: env.ValueFrom.SecretKeyRef.Key,
							}
						}
					}
					k8sEnvs = append(k8sEnvs, k8sEnv)
				}
				k8sContainer.Env = k8sEnvs
			}

			// 设置卷挂载
			if len(initContainer.VolumeMounts) > 0 {
				k8sMounts := make([]corev1.VolumeMount, 0, len(initContainer.VolumeMounts))
				for _, mount := range initContainer.VolumeMounts {
					k8sMount := corev1.VolumeMount{
						Name:      mount.Name,
						MountPath: mount.MountPath,
						SubPath:   mount.SubPath,
						ReadOnly:  mount.ReadOnly,
					}
					k8sMounts = append(k8sMounts, k8sMount)
				}
				k8sContainer.VolumeMounts = k8sMounts
			}

			k8sInitContainers = append(k8sInitContainers, k8sContainer)
		}
		deployment.Spec.Template.Spec.InitContainers = k8sInitContainers
	}

	// 设置卷
	if len(create.Volumes) > 0 {
		k8sVolumes := make([]corev1.Volume, 0, len(create.Volumes))
		for _, volume := range create.Volumes {
			k8sVolume := corev1.Volume{
				Name: volume.Name,
			}
			if volume.ConfigMap != nil {
				k8sVolume.ConfigMap = &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: volume.ConfigMap.Name,
					},
					Items:       convertKeyToPath(volume.ConfigMap.Items),
					DefaultMode: volume.ConfigMap.DefaultMode,
				}
			}
			if volume.Secret != nil {
				k8sVolume.Secret = &corev1.SecretVolumeSource{
					SecretName:  volume.Secret.SecretName,
					Items:       convertKeyToPath(volume.Secret.Items),
					DefaultMode: volume.Secret.DefaultMode,
				}
			}
			if volume.PersistentVolumeClaim != nil {
				k8sVolume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: volume.PersistentVolumeClaim.ClaimName,
					ReadOnly:  volume.PersistentVolumeClaim.ReadOnly,
				}
			}
			if volume.EmptyDir != nil {
				sizeLimit := resource.MustParse(volume.EmptyDir.SizeLimit)
				k8sVolume.EmptyDir = &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMedium(volume.EmptyDir.Medium),
					SizeLimit: &sizeLimit,
				}
			}
			if volume.HostPath != nil {
				k8sVolume.HostPath = &corev1.HostPathVolumeSource{
					Path: volume.HostPath.Path,
					Type: (*corev1.HostPathType)(&volume.HostPath.Type),
				}
			}
			k8sVolumes = append(k8sVolumes, k8sVolume)
		}
		deployment.Spec.Template.Spec.Volumes = k8sVolumes
	}

	// 设置容忍度
	if len(create.Tolerations) > 0 {
		k8sTolerations := make([]corev1.Toleration, 0, len(create.Tolerations))
		for _, toleration := range create.Tolerations {
			k8sToleration := corev1.Toleration{
				Key:               toleration.Key,
				Operator:          corev1.TolerationOperator(toleration.Operator),
				Value:             toleration.Value,
				Effect:            corev1.TaintEffect(toleration.Effect),
				TolerationSeconds: toleration.TolerationSeconds,
			}
			k8sTolerations = append(k8sTolerations, k8sToleration)
		}
		deployment.Spec.Template.Spec.Tolerations = k8sTolerations
	}

	// 设置亲和性
	if create.Affinity != nil {
		k8sAffinity := &corev1.Affinity{}

		// 设置节点亲和性
		if create.Affinity.NodeAffinity != nil {
			k8sNodeAffinity := &corev1.NodeAffinity{}

			if create.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				nodeSelectorTerms := make([]corev1.NodeSelectorTerm, 0, len(create.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms))
				for _, term := range create.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					nodeSelectorTerm := corev1.NodeSelectorTerm{}

					if len(term.MatchExpressions) > 0 {
						exprs := make([]corev1.NodeSelectorRequirement, 0, len(term.MatchExpressions))
						for _, expr := range term.MatchExpressions {
							exprs = append(exprs, corev1.NodeSelectorRequirement{
								Key:      expr.Key,
								Operator: corev1.NodeSelectorOperator(expr.Operator),
								Values:   expr.Values,
							})
						}
						nodeSelectorTerm.MatchExpressions = exprs
					}

					if len(term.MatchFields) > 0 {
						fields := make([]corev1.NodeSelectorRequirement, 0, len(term.MatchFields))
						for _, field := range term.MatchFields {
							fields = append(fields, corev1.NodeSelectorRequirement{
								Key:      field.Key,
								Operator: corev1.NodeSelectorOperator(field.Operator),
								Values:   field.Values,
							})
						}
						nodeSelectorTerm.MatchFields = fields
					}

					nodeSelectorTerms = append(nodeSelectorTerms, nodeSelectorTerm)
				}

				k8sNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
					NodeSelectorTerms: nodeSelectorTerms,
				}
			}

			k8sAffinity.NodeAffinity = k8sNodeAffinity
		}

		deployment.Spec.Template.Spec.Affinity = k8sAffinity
	}

	// 设置主机网络
	if create.HostNetwork != nil {
		deployment.Spec.Template.Spec.HostNetwork = *create.HostNetwork
	}

	// 设置DNS策略
	if create.DNSPolicy != "" {
		deployment.Spec.Template.Spec.DNSPolicy = corev1.DNSPolicy(create.DNSPolicy)
	}

	// 创建Deployment
	result, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建Deployment失败: %v", err)
	}

	logger.Info("成功创建Deployment:", clusterName, namespace)

	// 转换并返回创建结果
	deploymentInfo := ds.convertDeployment(result)
	return &deploymentInfo, nil
}

// convertProbe 将自定义Probe转换为K8s的Probe
func convertProbe(probe *Probe) *corev1.Probe {
	if probe == nil {
		return nil
	}

	k8sProbe := &corev1.Probe{
		InitialDelaySeconds: probe.InitialDelaySeconds,
		TimeoutSeconds:      probe.TimeoutSeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		SuccessThreshold:    probe.SuccessThreshold,
		FailureThreshold:    probe.FailureThreshold,
	}

	if probe.HTTPGet != nil {
		k8sProbe.HTTPGet = &corev1.HTTPGetAction{
			Path:   probe.HTTPGet.Path,
			Port:   intstr.FromInt(int(probe.HTTPGet.Port)),
			Host:   probe.HTTPGet.Host,
			Scheme: corev1.URIScheme(probe.HTTPGet.Scheme),
		}
		for _, header := range probe.HTTPGet.HTTPHeaders {
			k8sProbe.HTTPGet.HTTPHeaders = append(k8sProbe.HTTPGet.HTTPHeaders, corev1.HTTPHeader{
				Name:  header.Name,
				Value: header.Value,
			})
		}
	}

	if probe.TCPSocket != nil {
		k8sProbe.TCPSocket = &corev1.TCPSocketAction{
			Port: intstr.FromInt(int(probe.TCPSocket.Port)),
			Host: probe.TCPSocket.Host,
		}
	}

	if probe.Exec != nil {
		k8sProbe.Exec = &corev1.ExecAction{
			Command: probe.Exec.Command,
		}
	}

	return k8sProbe
}

// convertKeyToPath 将自定义KeyToPath转换为K8s的KeyToPath
func convertKeyToPath(items []KeyToPath) []corev1.KeyToPath {
	if items == nil {
		return nil
	}

	k8sItems := make([]corev1.KeyToPath, 0, len(items))
	for _, item := range items {
		k8sItem := corev1.KeyToPath{
			Key:  item.Key,
			Path: item.Path,
			Mode: item.Mode,
		}
		k8sItems = append(k8sItems, k8sItem)
	}

	return k8sItems
}
