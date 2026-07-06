package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceQuotaService ResourceQuota 管理服务
type ResourceQuotaService struct {
	clientManager *ClientManager
}

// NewResourceQuotaService 创建 ResourceQuota 服务
func NewResourceQuotaService(clientManager *ClientManager) *ResourceQuotaService {
	return &ResourceQuotaService{clientManager: clientManager}
}

// ResourceQuotaInfo ResourceQuota 摘要
type ResourceQuotaInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Hard         map[string]string `json:"hard,omitempty"`
	Used         map[string]string `json:"used,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// CreateResourceQuotaRequest 创建 ResourceQuota 请求
type CreateResourceQuotaRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	Hard      map[string]string `json:"hard" binding:"required"`
}

// UpdateResourceQuotaRequest 更新 ResourceQuota 请求
type UpdateResourceQuotaRequest struct {
	Hard   map[string]string `json:"hard,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

// ListResourceQuotas 获取 ResourceQuota 列表
func (s *ResourceQuotaService) ListResourceQuotas(ctx context.Context, clusterName, namespace string) ([]ResourceQuotaInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *corev1.ResourceQuotaList
	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().ResourceQuotas("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 ResourceQuota 列表失败: %w", err)
	}
	result := make([]ResourceQuotaInfo, 0, len(list.Items))
	for _, rq := range list.Items {
		result = append(result, convertResourceQuotaInfo(&rq))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetResourceQuota 获取 ResourceQuota 详情
func (s *ResourceQuotaService) GetResourceQuota(ctx context.Context, clusterName, namespace, name string) (*ResourceQuotaInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	rq, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ResourceQuota 失败: %w", err)
	}
	info := convertResourceQuotaInfo(rq)
	return &info, nil
}

// CreateResourceQuota 创建 ResourceQuota
func (s *ResourceQuotaService) CreateResourceQuota(ctx context.Context, clusterName, namespace string, req CreateResourceQuotaRequest) (*ResourceQuotaInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: parseResourceList(req.Hard),
		},
	}
	created, err := client.CoreV1().ResourceQuotas(namespace).Create(ctx, rq, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 ResourceQuota 失败: %w", err)
	}
	info := convertResourceQuotaInfo(created)
	return &info, nil
}

// UpdateResourceQuota 更新 ResourceQuota
func (s *ResourceQuotaService) UpdateResourceQuota(ctx context.Context, clusterName, namespace, name string, req UpdateResourceQuotaRequest) (*ResourceQuotaInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	rq, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ResourceQuota 失败: %w", err)
	}
	if req.Hard != nil {
		rq.Spec.Hard = parseResourceList(req.Hard)
	}
	if req.Labels != nil {
		rq.Labels = req.Labels
	}
	updated, err := client.CoreV1().ResourceQuotas(namespace).Update(ctx, rq, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 ResourceQuota 失败: %w", err)
	}
	info := convertResourceQuotaInfo(updated)
	return &info, nil
}

// DeleteResourceQuota 删除 ResourceQuota
func (s *ResourceQuotaService) DeleteResourceQuota(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertResourceQuotaInfo(rq *corev1.ResourceQuota) ResourceQuotaInfo {
	return ResourceQuotaInfo{
		Name:         rq.Name,
		Namespace:    rq.Namespace,
		Hard:         resourceListToMap(rq.Spec.Hard),
		Used:         resourceListToMap(rq.Status.Used),
		CreationTime: rq.CreationTimestamp.Time,
		Labels:       rq.Labels,
	}
}

func parseResourceList(m map[string]string) corev1.ResourceList {
	list := corev1.ResourceList{}
	for k, v := range m {
		list[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	return list
}

func resourceListToMap(list corev1.ResourceList) map[string]string {
	if len(list) == 0 {
		return nil
	}
	m := make(map[string]string, len(list))
	for k, v := range list {
		m[string(k)] = v.String()
	}
	return m
}
