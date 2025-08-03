package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// AutoScalerService 自动扩缩容服务
type AutoScalerService struct {
	clientManager *ClientManager
}

// NewAutoScalerService 创建自动扩缩容服务
func NewAutoScalerService(clientManager *ClientManager) *AutoScalerService {
	return &AutoScalerService{
		clientManager: clientManager,
	}
}

// AutoScalerConfig Cluster Autoscaler配置
type AutoScalerConfig struct {
	Enabled                bool              `json:"enabled"`
	Image                  string            `json:"image,omitempty"`
	ScaleDownDelay         string            `json:"scaleDownDelay,omitempty"`
	ScaleDownThreshold     string            `json:"scaleDownThreshold,omitempty"`
	ScaleUpThreshold       string            `json:"scaleUpThreshold,omitempty"`
	ScaleDownUnneededTime  string            `json:"scaleDownUnneededTime,omitempty"`
	ScaleDownDelayAfterAdd string            `json:"scaleDownDelayAfterAdd,omitempty"`
	NodeGroups             []NodeGroupConfig `json:"nodeGroups,omitempty"`
}

// NodeGroupConfig 节点组配置
type NodeGroupConfig struct {
	Name     string `json:"name"`
	MinNodes int32  `json:"minNodes"`
	MaxNodes int32  `json:"maxNodes"`
}

// GetAutoScalerConfig 获取集群自动扩缩容配置
func (s *AutoScalerService) GetAutoScalerConfig(ctx context.Context, clusterName string) (*AutoScalerConfig, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 从ConfigMap中获取配置
	configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "cluster-autoscaler-config", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &AutoScalerConfig{Enabled: false}, nil
		}
		return nil, fmt.Errorf("获取自动扩缩容配置失败: %w", err)
	}

	config := &AutoScalerConfig{}
	if data, ok := configMap.Data["config"]; ok {
		if err := json.Unmarshal([]byte(data), config); err != nil {
			return nil, fmt.Errorf("解析自动扩缩容配置失败: %w", err)
		}
	}

	return config, nil
}

// UpdateAutoScalerConfig 更新集群自动扩缩容配置
func (s *AutoScalerService) UpdateAutoScalerConfig(ctx context.Context, clusterName string, config *AutoScalerConfig) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	if config.Enabled {
		return s.enableAutoscaler(ctx, client, config)
	} else {
		return s.disableAutoscaler(ctx, client)
	}
}

// enableAutoscaler 启用自动扩缩容
func (s *AutoScalerService) enableAutoscaler(ctx context.Context, client *kubernetes.Clientset, config *AutoScalerConfig) error {
	// 设置默认值
	if config.Image == "" {
		config.Image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.3"
	}
	if config.ScaleDownDelay == "" {
		config.ScaleDownDelay = "10m"
	}
	if config.ScaleDownThreshold == "" {
		config.ScaleDownThreshold = "0.5"
	}
	if config.ScaleUpThreshold == "" {
		config.ScaleUpThreshold = "0.7"
	}
	if config.ScaleDownUnneededTime == "" {
		config.ScaleDownUnneededTime = "10m"
	}
	if config.ScaleDownDelayAfterAdd == "" {
		config.ScaleDownDelayAfterAdd = "10m"
	}

	// 创建或更新ConfigMap
	if err := s.createOrUpdateConfigMap(ctx, client, config); err != nil {
		return err
	}

	// 创建ServiceAccount
	if err := s.createServiceAccount(ctx, client); err != nil {
		return err
	}

	// 创建ClusterRole
	if err := s.createClusterRole(ctx, client); err != nil {
		return err
	}

	// 创建ClusterRoleBinding
	if err := s.createClusterRoleBinding(ctx, client); err != nil {
		return err
	}

	// 创建Deployment
	if err := s.createDeployment(ctx, client, config); err != nil {
		return err
	}

	return nil
}

// disableAutoscaler 禁用自动扩缩容
func (s *AutoScalerService) disableAutoscaler(ctx context.Context, client *kubernetes.Clientset) error {
	// 删除Deployment
	err := client.AppsV1().Deployments("kube-system").Delete(ctx, "cluster-autoscaler", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除Deployment失败: %w", err)
	}

	// 删除ClusterRoleBinding
	err = client.RbacV1().ClusterRoleBindings().Delete(ctx, "cluster-autoscaler", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除ClusterRoleBinding失败: %w", err)
	}

	// 删除ClusterRole
	err = client.RbacV1().ClusterRoles().Delete(ctx, "cluster-autoscaler", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除ClusterRole失败: %w", err)
	}

	// 删除ServiceAccount
	err = client.CoreV1().ServiceAccounts("kube-system").Delete(ctx, "cluster-autoscaler", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除ServiceAccount失败: %w", err)
	}

	// 更新ConfigMap标记为禁用
	config := &AutoScalerConfig{Enabled: false}
	return s.createOrUpdateConfigMap(ctx, client, config)
}

// createOrUpdateConfigMap 创建或更新ConfigMap
func (s *AutoScalerService) createOrUpdateConfigMap(ctx context.Context, client *kubernetes.Clientset, config *AutoScalerConfig) error {
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-autoscaler-config",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"config": string(configData),
		},
	}

	existing, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "cluster-autoscaler-config", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = client.CoreV1().ConfigMaps("kube-system").Create(ctx, configMap, metav1.CreateOptions{})
			return err
		}
		return err
	}

	existing.Data = configMap.Data
	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

// createServiceAccount 创建ServiceAccount
func (s *AutoScalerService) createServiceAccount(ctx context.Context, client *kubernetes.Clientset) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-autoscaler",
			Namespace: "kube-system",
		},
	}

	_, err := client.CoreV1().ServiceAccounts("kube-system").Create(ctx, sa, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// createClusterRole 创建ClusterRole
func (s *AutoScalerService) createClusterRole(ctx context.Context, client *kubernetes.Clientset) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-autoscaler",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events", "endpoints"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/eviction"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups:     []string{""},
				Resources:     []string{"endpoints"},
				ResourceNames: []string{"cluster-autoscaler"},
				Verbs:         []string{"get", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"watch", "list", "get", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "replicationcontrollers", "persistentvolumeclaims", "persistentvolumes"},
				Verbs:     []string{"watch", "list", "get"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"replicasets", "daemonsets"},
				Verbs:     []string{"watch", "list", "get"},
			},
			{
				APIGroups: []string{"policy"},
				Resources: []string{"poddisruptionbudgets"},
				Verbs:     []string{"watch", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets", "replicasets", "daemonsets"},
				Verbs:     []string{"watch", "list", "get"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses", "csinodes", "csidrivers", "csistoragecapacities"},
				Verbs:     []string{"watch", "list", "get"},
			},
			{
				APIGroups: []string{"batch", "extensions"},
				Resources: []string{"jobs"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups:     []string{"coordination.k8s.io"},
				ResourceNames: []string{"cluster-autoscaler"},
				Resources:     []string{"leases"},
				Verbs:         []string{"get", "update"},
			},
		},
	}

	_, err := client.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// createClusterRoleBinding 创建ClusterRoleBinding
func (s *AutoScalerService) createClusterRoleBinding(ctx context.Context, client *kubernetes.Clientset) error {
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-autoscaler",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-autoscaler",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "cluster-autoscaler",
				Namespace: "kube-system",
			},
		},
	}

	_, err := client.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// createDeployment 创建Deployment
func (s *AutoScalerService) createDeployment(ctx context.Context, client *kubernetes.Clientset, config *AutoScalerConfig) error {
	// 构建命令行参数
	args := []string{
		"--v=4",
		"--stderrthreshold=info",
		"--cloud-provider=generic",
		"--skip-nodes-with-local-storage=false",
		"--expander=least-waste",
		fmt.Sprintf("--scale-down-delay-after-add=%s", config.ScaleDownDelayAfterAdd),
		fmt.Sprintf("--scale-down-unneeded-time=%s", config.ScaleDownUnneededTime),
		"--balance-similar-node-groups",
	}

	// 添加节点组配置
	for _, nodeGroup := range config.NodeGroups {
		args = append(args, fmt.Sprintf("--nodes=%d:%d:%s", nodeGroup.MinNodes, nodeGroup.MaxNodes, nodeGroup.Name))
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-autoscaler",
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "cluster-autoscaler",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "cluster-autoscaler",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "cluster-autoscaler",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "cluster-autoscaler",
					Containers: []corev1.Container{
						{
							Name:  "cluster-autoscaler",
							Image: config.Image,
							Command: []string{
								"./cluster-autoscaler",
							},
							Args: args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("300Mi"),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("300Mi"),
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health-check",
										Port: intstr.FromInt(8085),
									},
								},
								InitialDelaySeconds: 300,
								PeriodSeconds:       60,
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/control-plane",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments("kube-system").Create(ctx, deployment, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		// 如果已存在则更新
		existing, getErr := client.AppsV1().Deployments("kube-system").Get(ctx, "cluster-autoscaler", metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		existing.Spec = deployment.Spec
		_, err = client.AppsV1().Deployments("kube-system").Update(ctx, existing, metav1.UpdateOptions{})
	}
	return err
}

// GetAutoScalerStatus 获取自动扩缩容状态
func (s *AutoScalerService) GetAutoScalerStatus(ctx context.Context, clusterName string) (*AutoScalerStatus, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	status := &AutoScalerStatus{}

	// 检查配置
	config, err := s.GetAutoScalerConfig(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	status.Enabled = config.Enabled

	if !config.Enabled {
		status.Status = "Disabled"
		return status, nil
	}

	// 检查Deployment状态
	deployment, err := client.AppsV1().Deployments("kube-system").Get(ctx, "cluster-autoscaler", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.Status = "NotDeployed"
			return status, nil
		}
		return nil, err
	}

	status.Replicas = deployment.Status.Replicas
	status.ReadyReplicas = deployment.Status.ReadyReplicas
	status.AvailableReplicas = deployment.Status.AvailableReplicas

	if deployment.Status.ReadyReplicas > 0 {
		status.Status = "Running"
	} else {
		status.Status = "Starting"
	}

	return status, nil
}

// AutoScalerStatus 自动扩缩容状态
type AutoScalerStatus struct {
	Enabled           bool   `json:"enabled"`
	Status            string `json:"status"` // Disabled, NotDeployed, Starting, Running
	Replicas          int32  `json:"replicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
}

// int32Ptr returns a pointer to an int32
func int32Ptr(i int32) *int32 {
	return &i
}
