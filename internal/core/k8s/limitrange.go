package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LimitRangeService LimitRange 管理服务
type LimitRangeService struct {
	clientManager *ClientManager
}

// NewLimitRangeService 创建 LimitRange 服务
func NewLimitRangeService(clientManager *ClientManager) *LimitRangeService {
	return &LimitRangeService{clientManager: clientManager}
}

// LimitRangeInfo LimitRange 摘要
type LimitRangeInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	LimitsCount  int               `json:"limitsCount"`
	CreationTime time.Time         `json:"creationTime"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// LimitRangeLimitSpec 限制规格
type LimitRangeLimitSpec struct {
	Type           string            `json:"type"`
	Default        map[string]string `json:"default,omitempty"`
	DefaultRequest map[string]string `json:"defaultRequest,omitempty"`
	Max            map[string]string `json:"max,omitempty"`
	Min            map[string]string `json:"min,omitempty"`
}

// CreateLimitRangeRequest 创建 LimitRange 请求
type CreateLimitRangeRequest struct {
	Name      string                `json:"name" binding:"required"`
	Namespace string                `json:"namespace"`
	Labels    map[string]string     `json:"labels,omitempty"`
	Limits    []LimitRangeLimitSpec `json:"limits" binding:"required"`
}

// UpdateLimitRangeRequest 更新 LimitRange 请求
type UpdateLimitRangeRequest struct {
	Labels map[string]string     `json:"labels,omitempty"`
	Limits []LimitRangeLimitSpec `json:"limits,omitempty"`
}

// ListLimitRanges 获取 LimitRange 列表
func (s *LimitRangeService) ListLimitRanges(ctx context.Context, clusterName, namespace string) ([]LimitRangeInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *corev1.LimitRangeList
	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().LimitRanges("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 LimitRange 列表失败: %w", err)
	}
	result := make([]LimitRangeInfo, 0, len(list.Items))
	for _, lr := range list.Items {
		result = append(result, convertLimitRangeInfo(&lr))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetLimitRange 获取 LimitRange 详情
func (s *LimitRangeService) GetLimitRange(ctx context.Context, clusterName, namespace, name string) (*corev1.LimitRange, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateLimitRange 创建 LimitRange
func (s *LimitRangeService) CreateLimitRange(ctx context.Context, clusterName, namespace string, req CreateLimitRangeRequest) (*LimitRangeInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: buildLimitRangeItems(req.Limits),
		},
	}
	created, err := client.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 LimitRange 失败: %w", err)
	}
	info := convertLimitRangeInfo(created)
	return &info, nil
}

// UpdateLimitRange 更新 LimitRange
func (s *LimitRangeService) UpdateLimitRange(ctx context.Context, clusterName, namespace, name string, req UpdateLimitRangeRequest) (*LimitRangeInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	lr, err := client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 LimitRange 失败: %w", err)
	}
	if req.Labels != nil {
		lr.Labels = req.Labels
	}
	if req.Limits != nil {
		lr.Spec.Limits = buildLimitRangeItems(req.Limits)
	}
	updated, err := client.CoreV1().LimitRanges(namespace).Update(ctx, lr, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 LimitRange 失败: %w", err)
	}
	info := convertLimitRangeInfo(updated)
	return &info, nil
}

// DeleteLimitRange 删除 LimitRange
func (s *LimitRangeService) DeleteLimitRange(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertLimitRangeInfo(lr *corev1.LimitRange) LimitRangeInfo {
	return LimitRangeInfo{
		Name:         lr.Name,
		Namespace:    lr.Namespace,
		LimitsCount:  len(lr.Spec.Limits),
		CreationTime: lr.CreationTimestamp.Time,
		Labels:       lr.Labels,
	}
}

func buildLimitRangeItems(specs []LimitRangeLimitSpec) []corev1.LimitRangeItem {
	items := make([]corev1.LimitRangeItem, 0, len(specs))
	for _, spec := range specs {
		item := corev1.LimitRangeItem{Type: corev1.LimitType(spec.Type)}
		if spec.Default != nil {
			item.Default = parseResourceList(spec.Default)
		}
		if spec.DefaultRequest != nil {
			item.DefaultRequest = parseResourceList(spec.DefaultRequest)
		}
		if spec.Max != nil {
			item.Max = parseResourceList(spec.Max)
		}
		if spec.Min != nil {
			item.Min = parseResourceList(spec.Min)
		}
		items = append(items, item)
	}
	return items
}
