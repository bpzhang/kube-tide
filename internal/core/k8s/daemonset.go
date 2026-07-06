package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetService DaemonSet 管理服务
type DaemonSetService struct {
	clientManager *ClientManager
}

// NewDaemonSetService 创建 DaemonSet 服务
func NewDaemonSetService(clientManager *ClientManager) *DaemonSetService {
	return &DaemonSetService{clientManager: clientManager}
}

// DaemonSetInfo DaemonSet 摘要
type DaemonSetInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	DesiredNumber     int32             `json:"desiredNumberScheduled"`
	CurrentNumber     int32             `json:"currentNumberScheduled"`
	ReadyNumber       int32             `json:"numberReady"`
	AvailableNumber   int32             `json:"numberAvailable"`
	UpdateStrategy    string            `json:"updateStrategy"`
	CreationTime      time.Time         `json:"creationTime"`
	Labels            map[string]string `json:"labels,omitempty"`
	Selector          map[string]string `json:"selector,omitempty"`
	ContainerCount    int               `json:"containerCount"`
	Images            []string          `json:"images"`
}

// CreateDaemonSetRequest 创建 DaemonSet 请求
type CreateDaemonSetRequest struct {
	Name         string            `json:"name" binding:"required"`
	Namespace    string            `json:"namespace"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Containers   []ContainerSpec   `json:"containers" binding:"required"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// UpdateDaemonSetRequest 更新 DaemonSet 请求
type UpdateDaemonSetRequest struct {
	Image        map[string]string               `json:"image,omitempty"`
	Resources    map[string]ResourceRequirements `json:"resources,omitempty"`
	Labels       map[string]string               `json:"labels,omitempty"`
	Annotations  map[string]string               `json:"annotations,omitempty"`
	NodeSelector map[string]string               `json:"nodeSelector,omitempty"`
}

// ListDaemonSets 获取 DaemonSet 列表
func (s *DaemonSetService) ListDaemonSets(ctx context.Context, clusterName, namespace string) ([]DaemonSetInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *appsv1.DaemonSetList
	if namespace == "all" || namespace == "" {
		list, err = client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 列表失败: %w", err)
	}
	result := make([]DaemonSetInfo, 0, len(list.Items))
	for _, ds := range list.Items {
		result = append(result, convertDaemonSetInfo(&ds))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetDaemonSet 获取 DaemonSet 详情
func (s *DaemonSetService) GetDaemonSet(ctx context.Context, clusterName, namespace, name string) (*DaemonSetInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}
	info := convertDaemonSetInfo(ds)
	return &info, nil
}

// CreateDaemonSet 创建 DaemonSet
func (s *DaemonSetService) CreateDaemonSet(ctx context.Context, clusterName, namespace string, req CreateDaemonSetRequest) (*DaemonSetInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	labels := req.Labels
	if labels == nil {
		labels = map[string]string{"app": req.Name}
	}
	containers := make([]corev1.Container, 0, len(req.Containers))
	for _, c := range req.Containers {
		containers = append(containers, buildContainerFromSpec(c))
	}
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: req.Annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: containers,
					NodeSelector: req.NodeSelector,
				},
			},
		},
	}
	created, err := client.AppsV1().DaemonSets(namespace).Create(ctx, ds, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 DaemonSet 失败: %w", err)
	}
	info := convertDaemonSetInfo(created)
	return &info, nil
}

// UpdateDaemonSet 更新 DaemonSet
func (s *DaemonSetService) UpdateDaemonSet(ctx context.Context, clusterName, namespace, name string, req UpdateDaemonSetRequest) (*DaemonSetInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}
	if req.Image != nil {
		for i, c := range ds.Spec.Template.Spec.Containers {
			if img, ok := req.Image[c.Name]; ok {
				ds.Spec.Template.Spec.Containers[i].Image = img
			}
		}
	}
	if req.Resources != nil {
		applyResourceUpdates(&ds.Spec.Template.Spec.Containers, req.Resources)
	}
	if req.Labels != nil {
		if ds.Labels == nil {
			ds.Labels = make(map[string]string)
		}
		for k, v := range req.Labels {
			ds.Labels[k] = v
			if ds.Spec.Selector != nil && ds.Spec.Selector.MatchLabels != nil {
				if _, inSelector := ds.Spec.Selector.MatchLabels[k]; inSelector {
					if ds.Spec.Template.Labels == nil {
						ds.Spec.Template.Labels = make(map[string]string)
					}
					ds.Spec.Template.Labels[k] = v
				}
			}
		}
	}
	if req.Annotations != nil {
		ds.Annotations = req.Annotations
	}
	if req.NodeSelector != nil {
		ds.Spec.Template.Spec.NodeSelector = req.NodeSelector
	}
	updated, err := client.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 DaemonSet 失败: %w", err)
	}
	info := convertDaemonSetInfo(updated)
	return &info, nil
}

// DeleteDaemonSet 删除 DaemonSet
func (s *DaemonSetService) DeleteDaemonSet(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetDaemonSetPods 获取 DaemonSet 关联 Pod
func (s *DaemonSetService) GetDaemonSetPods(ctx context.Context, clusterName, namespace, name string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}
	labelSelector := metav1.FormatLabelSelector(ds.Spec.Selector)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}
	return pods.Items, nil
}

func convertDaemonSetInfo(ds *appsv1.DaemonSet) DaemonSetInfo {
	images := make([]string, 0, len(ds.Spec.Template.Spec.Containers))
	for _, c := range ds.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	strategy := string(ds.Spec.UpdateStrategy.Type)
	return DaemonSetInfo{
		Name:            ds.Name,
		Namespace:       ds.Namespace,
		DesiredNumber:   ds.Status.DesiredNumberScheduled,
		CurrentNumber:   ds.Status.CurrentNumberScheduled,
		ReadyNumber:     ds.Status.NumberReady,
		AvailableNumber: ds.Status.NumberAvailable,
		UpdateStrategy:  strategy,
		CreationTime:    ds.CreationTimestamp.Time,
		Labels:          ds.Labels,
		Selector:        ds.Spec.Selector.MatchLabels,
		ContainerCount:  len(ds.Spec.Template.Spec.Containers),
		Images:          images,
	}
}
