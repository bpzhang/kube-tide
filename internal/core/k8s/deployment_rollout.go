package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RolloutStatus Deployment 滚动更新状态
type RolloutStatus struct {
	UpdatedReplicas     int32                 `json:"updatedReplicas"`
	ReadyReplicas       int32                 `json:"readyReplicas"`
	AvailableReplicas   int32                 `json:"availableReplicas"`
	UnavailableReplicas int32                 `json:"unavailableReplicas"`
	Replicas            int32                 `json:"replicas"`
	ObservedGeneration  int64                 `json:"observedGeneration"`
	Paused              bool                  `json:"paused"`
	Conditions          []DeploymentCondition `json:"conditions,omitempty"`
}

// CreateCanaryDeploymentRequest 创建金丝雀 Deployment 请求
type CreateCanaryDeploymentRequest struct {
	Name           string            `json:"name" binding:"required"`
	Replicas       *int32            `json:"replicas,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	CanaryLabelKey string            `json:"canaryLabelKey,omitempty"`
	CanaryLabelVal string            `json:"canaryLabelValue,omitempty"`
}

// PauseRollout 暂停 Deployment 滚动更新
func (ds *DeploymentService) PauseRollout(ctx context.Context, clusterName, namespace, name string) error {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}
	deployment.Spec.Paused = true
	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

// ResumeRollout 恢复 Deployment 滚动更新
func (ds *DeploymentService) ResumeRollout(ctx context.Context, clusterName, namespace, name string) error {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}
	deployment.Spec.Paused = false
	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

// GetRolloutStatus 获取 Deployment 滚动更新状态
func (ds *DeploymentService) GetRolloutStatus(ctx context.Context, clusterName, namespace, name string) (*RolloutStatus, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}
	status := &RolloutStatus{
		UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		ReadyReplicas:       deployment.Status.ReadyReplicas,
		AvailableReplicas:   deployment.Status.AvailableReplicas,
		UnavailableReplicas: deployment.Status.UnavailableReplicas,
		Replicas:            deployment.Status.Replicas,
		ObservedGeneration:  deployment.Status.ObservedGeneration,
		Paused:              deployment.Spec.Paused,
	}
	for _, c := range deployment.Status.Conditions {
		status.Conditions = append(status.Conditions, DeploymentCondition{
			Type:               string(c.Type),
			Status:             string(c.Status),
			LastUpdateTime:     c.LastUpdateTime.Time,
			LastTransitionTime: c.LastTransitionTime.Time,
			Reason:             c.Reason,
			Message:            c.Message,
		})
	}
	return status, nil
}

// CreateCanaryDeployment 基于现有 Deployment 创建金丝雀版本
func (ds *DeploymentService) CreateCanaryDeployment(ctx context.Context, clusterName, namespace, baseName string, req CreateCanaryDeploymentRequest) (*DeploymentInfo, error) {
	client, err := ds.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	base, err := client.AppsV1().Deployments(namespace).Get(ctx, baseName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取基础 Deployment 失败: %w", err)
	}
	if base.Spec.Selector == nil || len(base.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("基础 Deployment 缺少 selector.matchLabels，无法创建金丝雀版本")
	}

	canaryName := req.Name
	labelKey := req.CanaryLabelKey
	if labelKey == "" {
		labelKey = "track"
	}
	labelVal := req.CanaryLabelVal
	if labelVal == "" {
		labelVal = "canary"
	}

	labels := copyStringMap(base.Spec.Template.Labels)
	for k, v := range req.Labels {
		labels[k] = v
	}
	labels[labelKey] = labelVal

	selectorLabels := copyStringMap(base.Spec.Selector.MatchLabels)
	selectorLabels[labelKey] = labelVal

	replicas := int32(1)
	if req.Replicas != nil {
		replicas = *req.Replicas
	}

	canary := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        canaryName,
			Namespace:   namespace,
			Labels:      mergeStringMaps(base.Labels, req.Labels),
			Annotations: copyStringMap(base.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: copyStringMap(base.Spec.Template.Annotations),
				},
				Spec: *base.Spec.Template.Spec.DeepCopy(),
			},
			Strategy:                base.Spec.Strategy,
			MinReadySeconds:         base.Spec.MinReadySeconds,
			RevisionHistoryLimit:    base.Spec.RevisionHistoryLimit,
			ProgressDeadlineSeconds: base.Spec.ProgressDeadlineSeconds,
		},
	}

	created, err := client.AppsV1().Deployments(namespace).Create(ctx, canary, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建金丝雀 Deployment 失败: %w", err)
	}
	info := ds.convertDeployment(created)
	return &info, nil
}

func mergeStringMaps(base, extra map[string]string) map[string]string {
	out := copyStringMap(base)
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
