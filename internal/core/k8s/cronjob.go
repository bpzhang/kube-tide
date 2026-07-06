package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobService CronJob 管理服务
type CronJobService struct {
	clientManager *ClientManager
}

// NewCronJobService 创建 CronJob 服务
func NewCronJobService(clientManager *ClientManager) *CronJobService {
	return &CronJobService{clientManager: clientManager}
}

// CronJobInfo CronJob 摘要
type CronJobInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Schedule          string            `json:"schedule"`
	Suspend           bool              `json:"suspend"`
	LastScheduleTime  *time.Time        `json:"lastScheduleTime,omitempty"`
	LastSuccessfulTime *time.Time       `json:"lastSuccessfulTime,omitempty"`
	ActiveJobs        int               `json:"activeJobs"`
	CreationTime      time.Time         `json:"creationTime"`
	Labels            map[string]string `json:"labels,omitempty"`
	ConcurrencyPolicy string            `json:"concurrencyPolicy,omitempty"`
}

// CreateCronJobRequest 创建 CronJob 请求
type CreateCronJobRequest struct {
	Name              string            `json:"name" binding:"required"`
	Namespace         string            `json:"namespace"`
	Schedule          string            `json:"schedule" binding:"required"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Suspend           *bool             `json:"suspend,omitempty"`
	ConcurrencyPolicy string            `json:"concurrencyPolicy,omitempty"`
	Containers        []ContainerSpec   `json:"containers" binding:"required"`
	RestartPolicy     string            `json:"restartPolicy,omitempty"`
}

// UpdateCronJobRequest 更新 CronJob 请求
type UpdateCronJobRequest struct {
	Schedule          *string           `json:"schedule,omitempty"`
	Suspend           *bool             `json:"suspend,omitempty"`
	Image             map[string]string `json:"image,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	ConcurrencyPolicy *string           `json:"concurrencyPolicy,omitempty"`
}

// ListCronJobs 获取 CronJob 列表
func (s *CronJobService) ListCronJobs(ctx context.Context, clusterName, namespace string) ([]CronJobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *batchv1.CronJobList
	if namespace == "all" || namespace == "" {
		list, err = client.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 CronJob 列表失败: %w", err)
	}
	result := make([]CronJobInfo, 0, len(list.Items))
	for _, cj := range list.Items {
		result = append(result, convertCronJobInfo(&cj))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetCronJob 获取 CronJob 详情
func (s *CronJobService) GetCronJob(ctx context.Context, clusterName, namespace, name string) (*CronJobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 CronJob 失败: %w", err)
	}
	info := convertCronJobInfo(cj)
	return &info, nil
}

// CreateCronJob 创建 CronJob
func (s *CronJobService) CreateCronJob(ctx context.Context, clusterName, namespace string, req CreateCronJobRequest) (*CronJobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	labels := req.Labels
	if labels == nil {
		labels = map[string]string{"cronjob": req.Name}
	}
	containers := make([]corev1.Container, 0, len(req.Containers))
	for _, c := range req.Containers {
		containers = append(containers, buildContainerFromSpec(c))
	}
	restartPolicy := corev1.RestartPolicyOnFailure
	if req.RestartPolicy != "" {
		restartPolicy = corev1.RestartPolicy(req.RestartPolicy)
	}
	concurrency := batchv1.AllowConcurrent
	if req.ConcurrencyPolicy != "" {
		concurrency = batchv1.ConcurrencyPolicy(req.ConcurrencyPolicy)
	}
	suspend := false
	if req.Suspend != nil {
		suspend = *req.Suspend
	}
	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: req.Annotations,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          req.Schedule,
			Suspend:           &suspend,
			ConcurrencyPolicy: concurrency,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: restartPolicy,
							Containers:    containers,
						},
					},
				},
			},
		},
	}
	created, err := client.BatchV1().CronJobs(namespace).Create(ctx, cj, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 CronJob 失败: %w", err)
	}
	info := convertCronJobInfo(created)
	return &info, nil
}

// UpdateCronJob 更新 CronJob
func (s *CronJobService) UpdateCronJob(ctx context.Context, clusterName, namespace, name string, req UpdateCronJobRequest) (*CronJobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 CronJob 失败: %w", err)
	}
	if req.Schedule != nil {
		cj.Spec.Schedule = *req.Schedule
	}
	if req.Suspend != nil {
		cj.Spec.Suspend = req.Suspend
	}
	if req.ConcurrencyPolicy != nil {
		cj.Spec.ConcurrencyPolicy = batchv1.ConcurrencyPolicy(*req.ConcurrencyPolicy)
	}
	if req.Image != nil {
		for i, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if img, ok := req.Image[c.Name]; ok {
				cj.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Image = img
			}
		}
	}
	if req.Labels != nil {
		cj.Labels = req.Labels
	}
	if req.Annotations != nil {
		cj.Annotations = req.Annotations
	}
	updated, err := client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 CronJob 失败: %w", err)
	}
	info := convertCronJobInfo(updated)
	return &info, nil
}

// SuspendCronJob 暂停/恢复 CronJob
func (s *CronJobService) SuspendCronJob(ctx context.Context, clusterName, namespace, name string, suspend bool) (*CronJobInfo, error) {
	return s.UpdateCronJob(ctx, clusterName, namespace, name, UpdateCronJobRequest{Suspend: &suspend})
}

// DeleteCronJob 删除 CronJob
func (s *CronJobService) DeleteCronJob(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	return client.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func convertCronJobInfo(cj *batchv1.CronJob) CronJobInfo {
	suspend := false
	if cj.Spec.Suspend != nil {
		suspend = *cj.Spec.Suspend
	}
	info := CronJobInfo{
		Name:              cj.Name,
		Namespace:         cj.Namespace,
		Schedule:          cj.Spec.Schedule,
		Suspend:           suspend,
		ActiveJobs:        len(cj.Status.Active),
		CreationTime:      cj.CreationTimestamp.Time,
		Labels:            cj.Labels,
		ConcurrencyPolicy: string(cj.Spec.ConcurrencyPolicy),
	}
	if cj.Status.LastScheduleTime != nil {
		t := cj.Status.LastScheduleTime.Time
		info.LastScheduleTime = &t
	}
	if cj.Status.LastSuccessfulTime != nil {
		t := cj.Status.LastSuccessfulTime.Time
		info.LastSuccessfulTime = &t
	}
	return info
}
