package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPAService 提供 HPA 管理服务
type HPAService struct {
	clientManager *ClientManager
}

// NewHPAService 创建 HPA 服务
func NewHPAService(clientManager *ClientManager) *HPAService {
	return &HPAService{clientManager: clientManager}
}

// HPAInfo HPA 摘要信息
type HPAInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	MinReplicas     *int32            `json:"minReplicas,omitempty"`
	MaxReplicas     int32             `json:"maxReplicas"`
	CurrentReplicas int32             `json:"currentReplicas"`
	DesiredReplicas int32             `json:"desiredReplicas"`
	TargetRef       HPATargetRef      `json:"targetRef"`
	Metrics         []HPAMetricInfo   `json:"metrics,omitempty"`
	CreationTime    time.Time         `json:"creationTime"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// HPATargetRef 扩缩容目标引用
type HPATargetRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// HPAMetricInfo 指标摘要
type HPAMetricInfo struct {
	Type     string  `json:"type"`
	Name     string  `json:"name,omitempty"`
	Average  string  `json:"average,omitempty"`
	Utilization *int32 `json:"utilization,omitempty"`
}

// HPADetails HPA 详情
type HPADetails struct {
	HPAInfo
	Annotations map[string]string `json:"annotations,omitempty"`
	Behavior    *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// CreateHPARequest 创建 HPA 请求
type CreateHPARequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	MinReplicas *int32            `json:"minReplicas,omitempty"`
	MaxReplicas int32             `json:"maxReplicas" binding:"required"`
	TargetRef   HPATargetRef      `json:"targetRef" binding:"required"`
	Metrics     []HPAMetricSpec   `json:"metrics" binding:"required"`
}

// HPAMetricSpec 指标规格
type HPAMetricSpec struct {
	Type               string  `json:"type"`
	ResourceName       string  `json:"resourceName,omitempty"`
	TargetAverageValue string  `json:"targetAverageValue,omitempty"`
	TargetUtilization  *int32  `json:"targetUtilization,omitempty"`
}

// UpdateHPARequest 更新 HPA 请求
type UpdateHPARequest struct {
	MinReplicas *int32          `json:"minReplicas,omitempty"`
	MaxReplicas *int32          `json:"maxReplicas,omitempty"`
	Metrics     []HPAMetricSpec `json:"metrics,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (s *HPAService) listOptions(namespace string) (string, error) {
	if namespace == "all" || namespace == "" {
		return "", nil
	}
	return namespace, nil
}

// ListHPAs 获取 HPA 列表
func (s *HPAService) ListHPAs(ctx context.Context, clusterName, namespace string) ([]HPAInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ns, _ := s.listOptions(namespace)
	var list *autoscalingv2.HorizontalPodAutoscalerList
	if ns == "" {
		list, err = client.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 HPA 列表失败: %w", err)
	}

	result := make([]HPAInfo, 0, len(list.Items))
	for _, hpa := range list.Items {
		result = append(result, convertHPAInfo(&hpa))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetHPA 获取 HPA 详情
func (s *HPAService) GetHPA(ctx context.Context, clusterName, namespace, name string) (*HPADetails, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 HPA 失败: %w", err)
	}
	info := convertHPAInfo(hpa)
	return &HPADetails{
		HPAInfo:     info,
		Annotations: hpa.Annotations,
		Behavior:    hpa.Spec.Behavior,
	}, nil
}

// CreateHPA 创建 HPA
func (s *HPAService) CreateHPA(ctx context.Context, clusterName, namespace string, req CreateHPARequest) (*HPAInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: req.MinReplicas,
			MaxReplicas: req.MaxReplicas,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       req.TargetRef.Kind,
				Name:       req.TargetRef.Name,
				APIVersion: "apps/v1",
			},
			Metrics: buildHPAMetrics(req.Metrics),
		},
	}
	created, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 HPA 失败: %w", err)
	}
	info := convertHPAInfo(created)
	return &info, nil
}

// UpdateHPA 更新 HPA
func (s *HPAService) UpdateHPA(ctx context.Context, clusterName, namespace, name string, req UpdateHPARequest) (*HPAInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 HPA 失败: %w", err)
	}
	if req.MinReplicas != nil {
		hpa.Spec.MinReplicas = req.MinReplicas
	}
	if req.MaxReplicas != nil {
		hpa.Spec.MaxReplicas = *req.MaxReplicas
	}
	if req.Metrics != nil {
		hpa.Spec.Metrics = buildHPAMetrics(req.Metrics)
	}
	if req.Labels != nil {
		hpa.Labels = req.Labels
	}
	if req.Annotations != nil {
		hpa.Annotations = req.Annotations
	}
	updated, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, hpa, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 HPA 失败: %w", err)
	}
	info := convertHPAInfo(updated)
	return &info, nil
}

// DeleteHPA 删除 HPA
func (s *HPAService) DeleteHPA(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertHPAInfo(hpa *autoscalingv2.HorizontalPodAutoscaler) HPAInfo {
	metrics := make([]HPAMetricInfo, 0, len(hpa.Spec.Metrics))
	for _, m := range hpa.Spec.Metrics {
		info := HPAMetricInfo{Type: string(m.Type)}
		if m.Resource != nil {
			info.Name = string(m.Resource.Name)
			if m.Resource.Target.AverageUtilization != nil {
				info.Utilization = m.Resource.Target.AverageUtilization
			}
			if m.Resource.Target.AverageValue != nil {
				info.Average = m.Resource.Target.AverageValue.String()
			}
		}
		metrics = append(metrics, info)
	}
	return HPAInfo{
		Name:            hpa.Name,
		Namespace:       hpa.Namespace,
		MinReplicas:     hpa.Spec.MinReplicas,
		MaxReplicas:     hpa.Spec.MaxReplicas,
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		TargetRef: HPATargetRef{
			Kind: hpa.Spec.ScaleTargetRef.Kind,
			Name: hpa.Spec.ScaleTargetRef.Name,
		},
		Metrics:      metrics,
		CreationTime: hpa.CreationTimestamp.Time,
		Labels:       hpa.Labels,
	}
}

func buildHPAMetrics(specs []HPAMetricSpec) []autoscalingv2.MetricSpec {
	metrics := make([]autoscalingv2.MetricSpec, 0, len(specs))
	for _, spec := range specs {
		m := autoscalingv2.MetricSpec{Type: autoscalingv2.MetricSourceType(spec.Type)}
		if spec.Type == "Resource" {
			target := autoscalingv2.MetricTarget{Type: autoscalingv2.UtilizationMetricType}
			if spec.TargetUtilization != nil {
				target.AverageUtilization = spec.TargetUtilization
			}
			if spec.TargetAverageValue != "" {
				target.Type = autoscalingv2.AverageValueMetricType
				qty := resource.MustParse(spec.TargetAverageValue)
				target.AverageValue = &qty
			}
			resourceName := corev1.ResourceCPU
			if spec.ResourceName != "" {
				resourceName = corev1.ResourceName(spec.ResourceName)
			}
			m.Resource = &autoscalingv2.ResourceMetricSource{
				Name:   resourceName,
				Target: target,
			}
		}
		metrics = append(metrics, m)
	}
	return metrics
}
