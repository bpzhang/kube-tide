package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapInfo ConfigMap 摘要信息
type ConfigMapInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Labels       map[string]string `json:"labels,omitempty"`
	DataKeys     []string          `json:"dataKeys"`
	CreationTime string            `json:"creationTime"`
}

// ConfigMapDetail ConfigMap 详情
type ConfigMapDetail struct {
	ConfigMapInfo
	Data map[string]string `json:"data"`
}

// ConfigMapService ConfigMap 管理服务
type ConfigMapService struct {
	clientManager *ClientManager
}

// NewConfigMapService 创建 ConfigMap 服务
func NewConfigMapService(clientManager *ClientManager) *ConfigMapService {
	return &ConfigMapService{clientManager: clientManager}
}

func toConfigMapInfo(cm corev1.ConfigMap) ConfigMapInfo {
	keys := make([]string, 0, len(cm.Data))
	for k := range cm.Data {
		keys = append(keys, k)
	}
	return ConfigMapInfo{
		Name:         cm.Name,
		Namespace:    cm.Namespace,
		Labels:       cm.Labels,
		DataKeys:     keys,
		CreationTime: cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}
}

// ListConfigMaps 获取集群所有 ConfigMap
func (s *ConfigMapService) ListConfigMaps(ctx context.Context, clusterName string) ([]ConfigMapInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ConfigMap 列表失败: %w", err)
	}
	result := make([]ConfigMapInfo, 0, len(list.Items))
	for _, cm := range list.Items {
		result = append(result, toConfigMapInfo(cm))
	}
	return result, nil
}

// ListConfigMapsByNamespace 按命名空间获取 ConfigMap
func (s *ConfigMapService) ListConfigMapsByNamespace(ctx context.Context, clusterName, namespace string) ([]ConfigMapInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ConfigMap 列表失败: %w", err)
	}
	result := make([]ConfigMapInfo, 0, len(list.Items))
	for _, cm := range list.Items {
		result = append(result, toConfigMapInfo(cm))
	}
	return result, nil
}

// GetConfigMap 获取 ConfigMap 详情
func (s *ConfigMapService) GetConfigMap(ctx context.Context, clusterName, namespace, name string) (*ConfigMapDetail, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ConfigMap 失败: %w", err)
	}
	info := toConfigMapInfo(*cm)
	return &ConfigMapDetail{ConfigMapInfo: info, Data: cm.Data}, nil
}

// CreateConfigMapRequest 创建 ConfigMap 请求
type CreateConfigMapRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
	Data   map[string]string `json:"data"`
}

// CreateConfigMap 创建 ConfigMap
func (s *ConfigMapService) CreateConfigMap(ctx context.Context, clusterName, namespace string, req CreateConfigMapRequest) (*ConfigMapDetail, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("ConfigMap 名称不能为空")
	}
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: namespace, Labels: req.Labels},
		Data:       req.Data,
	}
	created, err := client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 ConfigMap 失败: %w", err)
	}
	info := toConfigMapInfo(*created)
	return &ConfigMapDetail{ConfigMapInfo: info, Data: created.Data}, nil
}

// UpdateConfigMap 更新 ConfigMap data
func (s *ConfigMapService) UpdateConfigMap(ctx context.Context, clusterName, namespace, name string, data map[string]string, labels map[string]string) (*ConfigMapDetail, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ConfigMap 失败: %w", err)
	}
	if data != nil {
		cm.Data = data
	}
	if labels != nil {
		cm.Labels = labels
	}
	updated, err := client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}
	info := toConfigMapInfo(*updated)
	return &ConfigMapDetail{ConfigMapInfo: info, Data: updated.Data}, nil
}

// DeleteConfigMap 删除 ConfigMap
func (s *ConfigMapService) DeleteConfigMap(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	if err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("删除 ConfigMap 失败: %w", err)
	}
	return nil
}
