package k8s

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"kube-tide/internal/utils/logger"

	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodeService 节点服务
type NodeService struct {
	clientManager   *ClientManager
	nodePoolService *NodePoolService
}

// NewNodeService 创建节点服务
func NewNodeService(clientManager *ClientManager, nodePoolService *NodePoolService) *NodeService {
	return &NodeService{
		clientManager:   clientManager,
		nodePoolService: nodePoolService,
	}
}

// GetNodes 获取节点列表
func (s *NodeService) GetNodes(ctx context.Context, clusterName string, limit int, page int) ([]corev1.Node, int, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, 0, err
	}

	// 获取所有节点
	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取节点列表失败: %w", err)
	}

	totalCount := len(nodeList.Items)

	// 如果没有指定分页参数，则返回所有节点
	if limit <= 0 || page <= 0 {
		return nodeList.Items, totalCount, nil
	}

	// 计算分页
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	// 边界检查
	if startIndex >= totalCount {
		return []corev1.Node{}, totalCount, nil
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	return nodeList.Items[startIndex:endIndex], totalCount, nil
}

// GetNodeDetails 获取节点详情
func (s *NodeService) GetNodeDetails(ctx context.Context, clusterName, nodeName string) (*corev1.Node, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点详情失败: %w", err)
	}

	return node, nil
}

// GetNodeStatus 获取节点状态信息
func (s *NodeService) GetNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			} else {
				return "NotReady"
			}
		}
	}
	return "Unknown"
}

// GetNodeMetrics 获取节点资源使用指标
func (s *NodeService) GetNodeMetrics(ctx context.Context, clusterName, nodeName string) (map[string]string, error) {
	node, err := s.GetNodeDetails(ctx, clusterName, nodeName)
	if err != nil {
		return nil, err
	}

	// 获取节点上所有的Pod
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取节点Pod列表失败: %w", err)
	}

	// 计算资源请求和限制
	var cpuRequests, cpuLimits int64
	var memoryRequests, memoryLimits int64

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				cpuReq := container.Resources.Requests.Cpu()
				memReq := container.Resources.Requests.Memory()
				if !cpuReq.IsZero() {
					cpuRequests += cpuReq.MilliValue()
				}
				if !memReq.IsZero() {
					memoryRequests += memReq.Value()
				}
			}
			if container.Resources.Limits != nil {
				cpuLim := container.Resources.Limits.Cpu()
				memLim := container.Resources.Limits.Memory()
				if !cpuLim.IsZero() {
					cpuLimits += cpuLim.MilliValue()
				}
				if !memLim.IsZero() {
					memoryLimits += memLim.Value()
				}
			}
		}
	}

	// 获取资源使用量（通过metrics-server）
	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return nil, err
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建metrics客户端失败: %w", err)
	}

	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})

	// 返回所有指标
	metrics := map[string]string{
		"cpu_capacity":    node.Status.Capacity.Cpu().String(),
		"memory_capacity": node.Status.Capacity.Memory().String(),
	}

	// 如果成功获取到使用量指标，添加使用量数据
	if err == nil && nodeMetrics != nil {
		metrics["cpu_usage"] = fmt.Sprintf("%dm", nodeMetrics.Usage.Cpu().MilliValue())
		metrics["memory_usage"] = fmt.Sprintf("%dMi", nodeMetrics.Usage.Memory().Value()/(1024*1024))
	} else {
		metrics["cpu_usage"] = "0"
		metrics["memory_usage"] = "0"
	}

	// 添加请求量和限制量
	if cpuRequests > 0 {
		metrics["cpu_requests"] = fmt.Sprintf("%dm", cpuRequests)
	} else {
		metrics["cpu_requests"] = "0"
	}
	if cpuLimits > 0 {
		metrics["cpu_limits"] = fmt.Sprintf("%dm", cpuLimits)
	} else {
		metrics["cpu_limits"] = "0"
	}
	if memoryRequests > 0 {
		metrics["memory_requests"] = fmt.Sprintf("%dMi", memoryRequests/(1024*1024))
	} else {
		metrics["memory_requests"] = "0"
	}
	if memoryLimits > 0 {
		metrics["memory_limits"] = fmt.Sprintf("%dMi", memoryLimits/(1024*1024))
	} else {
		metrics["memory_limits"] = "0"
	}

	return metrics, nil
}

// DrainNode 对节点进行排水操作
func (s *NodeService) DrainNode(ctx context.Context, clusterName, nodeName string, gracePeriodSeconds int, deleteLocalData bool, ignoreDaemonSets bool) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 首先将节点设置为不可调度
	if err = s.CordonNode(ctx, clusterName, nodeName); err != nil {
		return fmt.Errorf("设置节点为不可调度状态失败: %w", err)
	}

	// 创建排水辅助对象
	drainHelper := &drain.Helper{
		Client:              client,
		Force:               true,
		GracePeriodSeconds:  int(gracePeriodSeconds),
		IgnoreAllDaemonSets: ignoreDaemonSets,
		DeleteEmptyDirData:  deleteLocalData,
		Timeout:             time.Minute * 5,
		Out:                 io.Discard, // 使用标准库的io.Discard替代
		ErrOut:              io.Discard, // 使用标准库的io.Discard替代
	}

	// 执行节点排水
	if err = drain.RunNodeDrain(drainHelper, nodeName); err != nil {
		return fmt.Errorf("节点排水操作失败: %w", err)
	}

	return nil
}

// CordonNode 将节点设置为不可调度
func (s *NodeService) CordonNode(ctx context.Context, clusterName, nodeName string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	return s.updateNodeScheduling(ctx, client, nodeName, true)
}

// UncordonNode 将节点设置为可调度
func (s *NodeService) UncordonNode(ctx context.Context, clusterName, nodeName string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	return s.updateNodeScheduling(ctx, client, nodeName, false)
}

// updateNodeScheduling 更新节点的调度状态
func (s *NodeService) updateNodeScheduling(ctx context.Context, client kubernetes.Interface, nodeName string, unschedulable bool) error {
	// 获取节点
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 如果节点当前状态已经是目标状态，直接返回
	if node.Spec.Unschedulable == unschedulable {
		return nil
	}

	// 设置节点的调度状态
	node.Spec.Unschedulable = unschedulable

	// 更新节点
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		action := "不可调度"
		if !unschedulable {
			action = "可调度"
		}
		return fmt.Errorf("将节点设置为%s状态失败: %w", action, err)
	}

	return nil
}

// 污点管理相关方法

// GetNodeTaints 获取节点的污点
func (s *NodeService) GetNodeTaints(ctx context.Context, clusterName, nodeName string) ([]corev1.Taint, error) {
	node, err := s.GetNodeDetails(ctx, clusterName, nodeName)
	if err != nil {
		return nil, err
	}
	return node.Spec.Taints, nil
}

// AddNodeTaint 添加节点污点
func (s *NodeService) AddNodeTaint(ctx context.Context, clusterName, nodeName string, taint corev1.Taint) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 获取节点
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 检查污点是否已存在
	for _, existingTaint := range node.Spec.Taints {
		if existingTaint.Key == taint.Key && existingTaint.Effect == taint.Effect {
			return fmt.Errorf("污点已存在")
		}
	}

	// 添加新污点
	node.Spec.Taints = append(node.Spec.Taints, taint)

	// 更新节点
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("添加污点失败: %w", err)
	}

	return nil
}

// RemoveNodeTaint 删除节点污点
func (s *NodeService) RemoveNodeTaint(ctx context.Context, clusterName, nodeName string, taintKey string, effect corev1.TaintEffect) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 获取节点
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 查找并删除匹配的污点
	var newTaints []corev1.Taint
	taintFound := false
	for _, existingTaint := range node.Spec.Taints {
		if existingTaint.Key == taintKey && existingTaint.Effect == effect {
			taintFound = true
			continue
		}
		newTaints = append(newTaints, existingTaint)
	}

	if !taintFound {
		return fmt.Errorf("未找到指定的污点")
	}

	// 更新节点的污点
	node.Spec.Taints = newTaints

	// 更新节点
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("删除污点失败: %w", err)
	}

	return nil
}

// 标签管理相关方法

// GetNodeLabels 获取节点的标签
func (s *NodeService) GetNodeLabels(ctx context.Context, clusterName, nodeName string) (map[string]string, error) {
	node, err := s.GetNodeDetails(ctx, clusterName, nodeName)
	if err != nil {
		return nil, err
	}
	return node.Labels, nil
}

// AddNodeLabel 添加或更新节点标签
func (s *NodeService) AddNodeLabel(ctx context.Context, clusterName, nodeName, key, value string) error {

	logger.Info("开始添加节点标签")

	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		logger.Error("获取客户端失败", err)
		return err
	}

	// 获取节点
	logger.Debug("正在获取节点")
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Error("获取节点失败", err)
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 如果节点没有标签，初始化标签map
	if node.Labels == nil {
		logger.Debug("节点没有标签，初始化标签map")
		node.Labels = make(map[string]string)
	}

	oldValue, exists := node.Labels[key]
	if exists {
		logger.Info("更新已有标签", oldValue)
	} else {
		logger.Info("添加新标签")
	}

	// 添加或更新标签
	node.Labels[key] = value

	// 更新节点
	logger.Debug("正在更新节点")
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		logger.Error("添加标签失败", err)
		return fmt.Errorf("添加标签失败: %w", err)
	}

	logger.Info("成功添加节点标签")
	return nil
}

// RemoveNodeLabel 删除节点标签
func (s *NodeService) RemoveNodeLabel(ctx context.Context, clusterName, nodeName, key string) error {

	logger.Info("开始删除节点标签")

	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		logger.Error("获取客户端失败", err)
		return err
	}

	// 获取节点
	logger.Debug("正在获取节点")
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Error("获取节点失败", err)
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 检查标签是否存在
	value, exists := node.Labels[key]
	if !exists {
		logger.Warn("标签不存在", key)
		return fmt.Errorf("标签不存在")
	}

	logger.Info("删除标签", value)
	// 删除标签
	delete(node.Labels, key)

	// 更新节点
	logger.Debug("正在更新节点")
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		logger.Error("删除标签失败", err)
		return fmt.Errorf("删除标签失败: %w", err)
	}

	logger.Info("成功删除节点标签")
	return nil
}

// NodeConfig 节点配置
type NodeConfig struct {
	Name        string            `json:"name"`
	IP          string            `json:"ip"`
	Role        string            `json:"role"`
	Labels      map[string]string `json:"labels"`
	Taints      []corev1.Taint    `json:"taints"`
	NodePool    string            `json:"nodePool"`
	SSHPort     int               `json:"sshPort"`
	SSHUser     string            `json:"sshUser"`
	AuthType    string            `json:"authType"` // "key" 或 "password"
	SSHKeyFile  string            `json:"sshKeyFile"`
	SSHPassword string            `json:"sshPassword"`
}

// AddNode 添加新节点
func (s *NodeService) AddNode(ctx context.Context, clusterName string, nodeConfig NodeConfig) error {
	// 如果指定了节点池，获取节点池配置
	if nodeConfig.NodePool != "" {
		nodePool, err := s.nodePoolService.GetNodePool(ctx, clusterName, nodeConfig.NodePool)
		if err != nil {
			return fmt.Errorf("获取节点池配置失败: %w", err)
		}

		// 初始化标签map
		if nodeConfig.Labels == nil {
			nodeConfig.Labels = make(map[string]string)
		}

		// 添加节点池标签
		nodeConfig.Labels["k8s.io/pool-name"] = nodeConfig.NodePool

		// 合并节点池的标签
		for k, v := range nodePool.Labels {
			nodeConfig.Labels[k] = v
		}

		// 合并节点池的污点
		nodeConfig.Taints = append(nodeConfig.Taints, nodePool.Taints...)
	}

	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	// 获取集群配置和token
	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return fmt.Errorf("获取集群配置失败: %w", err)
	}

	// 从配置中提取API服务器地址
	apiServerURL := config.Host

	// 获取集群token
	// 通过创建service account并获取其token
	tokenName := fmt.Sprintf("node-joiner-%s", nodeConfig.Name)

	// 创建一个service account
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenName,
			Namespace: "kube-system",
		},
	}

	// 检查service account是否已存在，如果不存在则创建
	_, err = client.CoreV1().ServiceAccounts("kube-system").Get(ctx, tokenName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = client.CoreV1().ServiceAccounts("kube-system").Create(ctx, sa, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建service account失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("检查service account失败: %w", err)
	}

	// 创建ClusterRoleBinding以赋予该SA必要的权限
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: tokenName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      tokenName,
				Namespace: "kube-system",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:node-bootstrapper",
		},
	}

	// 检查ClusterRoleBinding是否已存在，如果不存在则创建
	_, err = client.RbacV1().ClusterRoleBindings().Get(ctx, tokenName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建ClusterRoleBinding失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("检查ClusterRoleBinding失败: %w", err)
	}

	// 创建token
	tokenRequest := &authenticationv1.TokenRequest{}
	tr, err := client.CoreV1().ServiceAccounts("kube-system").CreateToken(ctx, tokenName, tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建token失败: %w", err)
	}

	token := tr.Status.Token

	// 准备shell命令，根据认证类型来确定SSH连接方式
	var cmd string

	// 构建kubeadm join命令
	joinCmd := fmt.Sprintf("kubeadm join %s --token %s --discovery-token-unsafe-skip-ca-verification",
		strings.TrimPrefix(apiServerURL, "https://"), token)

	// 根据认证类型执行不同的SSH命令
	if nodeConfig.AuthType == "password" {
		// 使用sshpass通过密码连接
		cmd = fmt.Sprintf("sshpass -p '%s' ssh -o StrictHostKeyChecking=no -p %d %s@%s '%s'",
			nodeConfig.SSHPassword, nodeConfig.SSHPort, nodeConfig.SSHUser, nodeConfig.IP, joinCmd)
	} else {
		// 默认使用密钥文件连接
		cmd = fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -p %d %s@%s '%s'",
			nodeConfig.SSHKeyFile, nodeConfig.SSHPort, nodeConfig.SSHUser, nodeConfig.IP, joinCmd)
	}

	// 执行shell命令
	// 注意: 这里简化处理，实际应用中需要更健壮的SSH客户端实现
	command := exec.Command("sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行命令失败: %s, %w", string(output), err)
	}

	return nil
}

// RemoveNode 移除节点
func (s *NodeService) RemoveNode(ctx context.Context, clusterName, nodeName string, force bool) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return fmt.Errorf("获取客户端失败: %w", err)
	}

	// 如果不是强制删除，先检查节点上是否还有Pod
	if !force {
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + nodeName,
		})
		if err != nil {
			return fmt.Errorf("检查节点Pod失败: %w", err)
		}

		if len(pods.Items) > 0 {
			return fmt.Errorf("节点上仍有 %d 个Pod运行，请先驱逐或使用强制删除", len(pods.Items))
		}
	}

	// 如果配置了force，先将节点设置为不可调度
	if force {
		if err := s.CordonNode(ctx, clusterName, nodeName); err != nil {
			return fmt.Errorf("设置节点为不可调度失败: %w", err)
		}

		// 驱逐节点上的Pod
		if err := s.DrainNode(ctx, clusterName, nodeName, 300, true, true); err != nil {
			return fmt.Errorf("驱逐节点Pod失败: %w", err)
		}
	}

	// 删除节点
	err = client.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除节点失败: %w", err)
	}

	// TODO: 清理节点上的Kubernetes组件
	// 这部分可能包括：
	// 1. 重置kubeadm（如果使用kubeadm）
	// 2. 停止并删除kubelet服务
	// 3. 清理相关配置文件和证书

	return nil
}

// GetNodePods 获取节点上运行的Pod列表
func (s *NodeService) GetNodePods(ctx context.Context, clusterName, nodeName string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 使用字段选择器查询在特定节点上运行的Pod
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取节点Pod列表失败: %w", err)
	}

	return pods.Items, nil
}
