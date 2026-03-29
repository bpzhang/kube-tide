package k8s

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressManager Ingress 资源管理器。
type IngressManager struct {
	clientManager *ClientManager
}

// NewIngressManager 创建 Ingress 管理器。
func NewIngressManager(clientManager *ClientManager) *IngressManager {
	return &IngressManager{clientManager: clientManager}
}

// GetIngressesByNamespace 获取指定命名空间中的 Ingress 列表。
func (m *IngressManager) GetIngressesByNamespace(ctx context.Context, clusterName, namespace string) ([]networkingv1.Ingress, error) {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间 %s 的 Ingress 列表失败: %w", namespace, err)
	}

	return ingressList.Items, nil
}
