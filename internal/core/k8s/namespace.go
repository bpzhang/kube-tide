package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"kube-tide/internal/utils/logger"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// NamespaceInfo 命名空间详细信息
type NamespaceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
}

// NamespaceListResult 命名空间列表结果
type NamespaceListResult struct {
	Namespaces []string        `json:"namespaces"`
	Items      []NamespaceInfo `json:"items"`
}

// CreateNamespaceRequest 创建命名空间请求
type CreateNamespaceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PatchNamespaceLabelsRequest 更新命名空间标签请求
type PatchNamespaceLabelsRequest struct {
	Labels map[string]string `json:"labels" binding:"required"`
}

// ListNamespaces 获取指定集群的所有命名空间
func (s *NamespaceService) ListNamespaces(clusterName string) (*NamespaceListResult, error) {
	logger.Info("获取命名空间列表", "clusterName", clusterName)
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群 %s 的客户端失败: %w", clusterName, err)
	}

	namespaceList, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	items := make([]NamespaceInfo, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
		items = append(items, convertNamespaceInfo(&ns))
	}

	sort.Slice(namespaces, func(i, j int) bool {
		return namespaceSortLess(namespaces[i], namespaces[j])
	})
	sort.Slice(items, func(i, j int) bool {
		return namespaceSortLess(items[i].Name, items[j].Name)
	})

	return &NamespaceListResult{Namespaces: namespaces, Items: items}, nil
}

// GetNamespace 获取单个命名空间
func (s *NamespaceService) GetNamespace(ctx context.Context, clusterName, name string) (*NamespaceInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ns, err := client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间失败: %w", err)
	}
	info := convertNamespaceInfo(ns)
	return &info, nil
}

// CreateNamespace 创建命名空间
func (s *NamespaceService) CreateNamespace(ctx context.Context, clusterName string, req CreateNamespaceRequest) (*NamespaceInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
	}
	created, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建命名空间失败: %w", err)
	}
	info := convertNamespaceInfo(created)
	return &info, nil
}

// DeleteNamespace 删除命名空间
func (s *NamespaceService) DeleteNamespace(ctx context.Context, clusterName, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

// PatchNamespaceLabels 更新命名空间标签
func (s *NamespaceService) PatchNamespaceLabels(ctx context.Context, clusterName, name string, req PatchNamespaceLabelsRequest) (*NamespaceInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	patch := map[string]interface{}{"metadata": map[string]interface{}{"labels": req.Labels}}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("构建 patch 失败: %w", err)
	}
	ns, err := client.CoreV1().Namespaces().Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新命名空间标签失败: %w", err)
	}
	info := convertNamespaceInfo(ns)
	return &info, nil
}

func convertNamespaceInfo(ns *corev1.Namespace) NamespaceInfo {
	status := string(ns.Status.Phase)
	return NamespaceInfo{
		Name:         ns.Name,
		Status:       status,
		Labels:       ns.Labels,
		Annotations:  ns.Annotations,
		CreationTime: ns.CreationTimestamp.Time,
	}
}

func namespaceSortLess(a, b string) bool {
	if a == "default" {
		return true
	}
	if b == "default" {
		return false
	}
	if a == "kube-system" {
		return true
	}
	if b == "kube-system" {
		return false
	}
	return a < b
}
