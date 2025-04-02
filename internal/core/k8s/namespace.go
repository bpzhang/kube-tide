package k8s

import (
	"context"
	"fmt"
	"sort"
	"kube-tide/internal/utils/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceService 提供命名空间相关的服务
type NamespaceService struct {
	clientManager *ClientManager
}

// NewNamespaceService 创建新的命名空间服务
func NewNamespaceService(clientManager *ClientManager) *NamespaceService {
	return &NamespaceService{
		clientManager: clientManager,
	}
}

// ListNamespaces 获取指定集群的所有命名空间
func (s *NamespaceService) ListNamespaces(clusterName string) ([]string, error) {
	logger.Info("获取集群 %s 的命名空间列表", logger.String("clusterName", clusterName))
	// 获取集群的Kubernetes客户端
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群 %s 的客户端失败: %w", clusterName, err)
	}

	// 调用Kubernetes API获取命名空间列表
	namespaceList, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	// 提取命名空间名称
	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	// 对命名空间进行排序，确保默认命名空间在前面
	sort.Slice(namespaces, func(i, j int) bool {
		// 默认命名空间 default 和 kube-system 排在最前面
		if namespaces[i] == "default" {
			return true
		}
		if namespaces[j] == "default" {
			return false
		}
		if namespaces[i] == "kube-system" {
			return true
		}
		if namespaces[j] == "kube-system" {
			return false
		}
		return namespaces[i] < namespaces[j]
	})

	return namespaces, nil
}
