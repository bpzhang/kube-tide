package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodePool 节点池配置
type NodePool struct {
	Name        string             `json:"name"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Taints      []corev1.Taint     `json:"taints,omitempty"`
	AutoScaling *AutoScalingConfig `json:"autoScaling,omitempty"`
}

// AutoScalingConfig 自动扩缩容配置
type AutoScalingConfig struct {
	Enabled                bool   `json:"enabled"`
	MinNodes               int32  `json:"minNodes"`
	MaxNodes               int32  `json:"maxNodes"`
	ScaleDownDelay         string `json:"scaleDownDelay,omitempty"`         // 缩容延迟时间，默认 10m
	ScaleDownThreshold     string `json:"scaleDownThreshold,omitempty"`     // 缩容阈值，默认 0.5 (50%)
	ScaleUpThreshold       string `json:"scaleUpThreshold,omitempty"`       // 扩容阈值，默认 0.7 (70%)
	ScaleDownUnneededTime  string `json:"scaleDownUnneededTime,omitempty"`  // 节点空闲多长时间后可以被缩容，默认 10m
	ScaleDownDelayAfterAdd string `json:"scaleDownDelayAfterAdd,omitempty"` // 添加节点后多长时间内不进行缩容，默认 10m
}

// NodePoolService 节点池服务
type NodePoolService struct {
	clientManager *ClientManager
}

// NewNodePoolService 创建节点池服务
func NewNodePoolService(clientManager *ClientManager) *NodePoolService {
	return &NodePoolService{
		clientManager: clientManager,
	}
}

// ListNodePools 获取集群的所有节点池
func (s *NodePoolService) ListNodePools(ctx context.Context, clusterName string) ([]NodePool, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 从ConfigMap中获取节点池配置
	configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "node-pools", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 如果ConfigMap不存在，返回空列表
			return []NodePool{}, nil
		}
		return nil, fmt.Errorf("获取节点池配置失败: %w", err)
	}

	// 解析节点池配置
	pools := []NodePool{}
	if data, ok := configMap.Data["pools"]; ok {
		if err := json.Unmarshal([]byte(data), &pools); err != nil {
			return nil, fmt.Errorf("解析节点池配置失败: %w", err)
		}
	}

	return pools, nil
}

// CreateNodePool 创建新的节点池
func (s *NodePoolService) CreateNodePool(ctx context.Context, clusterName string, pool NodePool) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 获取现有的节点池配置
	configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "node-pools", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 如果ConfigMap不存在，创建新的
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-pools",
					Namespace: "kube-system",
				},
				Data: map[string]string{},
			}
		} else {
			return fmt.Errorf("获取节点池配置失败: %w", err)
		}
	}

	// 获取现有的节点池列表
	pools := []NodePool{}
	if data, ok := configMap.Data["pools"]; ok {
		if err := json.Unmarshal([]byte(data), &pools); err != nil {
			return fmt.Errorf("解析节点池配置失败: %w", err)
		}
	}

	// 检查节点池名称是否已存在
	for _, p := range pools {
		if p.Name == pool.Name {
			return fmt.Errorf("节点池 %s 已存在", pool.Name)
		}
	}

	// 添加新的节点池
	pools = append(pools, pool)

	// 更新ConfigMap
	poolsData, err := json.Marshal(pools)
	if err != nil {
		return fmt.Errorf("序列化节点池配置失败: %w", err)
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	configMap.Data["pools"] = string(poolsData)

	// 保存ConfigMap
	if _, err := client.CoreV1().ConfigMaps("kube-system").Update(ctx, configMap, metav1.UpdateOptions{}); err != nil {
		if errors.IsNotFound(err) {
			// 如果不存在则创建
			_, err = client.CoreV1().ConfigMaps("kube-system").Create(ctx, configMap, metav1.CreateOptions{})
		}
		if err != nil {
			return fmt.Errorf("保存节点池配置失败: %w", err)
		}
	}

	return nil
}

// UpdateNodePool 更新节点池配置
func (s *NodePoolService) UpdateNodePool(ctx context.Context, clusterName string, pool NodePool) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 获取现有的节点池配置
	configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "node-pools", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点池配置失败: %w", err)
	}

	// 解析现有的节点池列表
	pools := []NodePool{}
	if data, ok := configMap.Data["pools"]; ok {
		if err := json.Unmarshal([]byte(data), &pools); err != nil {
			return fmt.Errorf("解析节点池配置失败: %w", err)
		}
	}

	// 查找并更新节点池
	found := false
	for i, p := range pools {
		if p.Name == pool.Name {
			pools[i] = pool
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("节点池 %s 不存在", pool.Name)
	}

	// 更新ConfigMap
	poolsData, err := json.Marshal(pools)
	if err != nil {
		return fmt.Errorf("序列化节点池配置失败: %w", err)
	}

	configMap.Data["pools"] = string(poolsData)

	// 保存ConfigMap
	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("保存节点池配置失败: %w", err)
	}

	return nil
}

// DeleteNodePool 删除节点池
func (s *NodePoolService) DeleteNodePool(ctx context.Context, clusterName, poolName string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 获取现有的节点池配置
	configMap, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "node-pools", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点池配置失败: %w", err)
	}

	// 解析现有的节点池列表
	pools := []NodePool{}
	if data, ok := configMap.Data["pools"]; ok {
		if err := json.Unmarshal([]byte(data), &pools); err != nil {
			return fmt.Errorf("解析节点池配置失败: %w", err)
		}
	}

	// 查找并删除节点池
	found := false
	newPools := []NodePool{}
	for _, p := range pools {
		if p.Name == poolName {
			found = true
			continue
		}
		newPools = append(newPools, p)
	}

	if !found {
		return fmt.Errorf("节点池 %s 不存在", poolName)
	}

	// 更新ConfigMap
	poolsData, err := json.Marshal(newPools)
	if err != nil {
		return fmt.Errorf("序列化节点池配置失败: %w", err)
	}

	configMap.Data["pools"] = string(poolsData)

	// 保存ConfigMap
	_, err = client.CoreV1().ConfigMaps("kube-system").Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("保存节点池配置失败: %w", err)
	}

	return nil
}

// GetNodePool 获取指定节点池的配置
func (s *NodePoolService) GetNodePool(ctx context.Context, clusterName, poolName string) (*NodePool, error) {
	pools, err := s.ListNodePools(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		if pool.Name == poolName {
			return &pool, nil
		}
	}

	return nil, fmt.Errorf("节点池 %s 不存在", poolName)
}
