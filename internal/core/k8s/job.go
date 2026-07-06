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

// JobService Job 管理服务
type JobService struct {
	clientManager *ClientManager
}

// NewJobService 创建 Job 服务
func NewJobService(clientManager *ClientManager) *JobService {
	return &JobService{clientManager: clientManager}
}

// JobInfo Job 摘要
type JobInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Completions    *int32            `json:"completions,omitempty"`
	Parallelism    *int32            `json:"parallelism,omitempty"`
	Succeeded      int32             `json:"succeeded"`
	Failed         int32             `json:"failed"`
	Active         int32             `json:"active"`
	StartTime      *time.Time        `json:"startTime,omitempty"`
	CompletionTime *time.Time        `json:"completionTime,omitempty"`
	CreationTime   time.Time         `json:"creationTime"`
	Labels         map[string]string `json:"labels,omitempty"`
	Images         []string          `json:"images,omitempty"`
}

// CreateJobRequest 创建 Job 请求
type CreateJobRequest struct {
	Name         string            `json:"name" binding:"required"`
	Namespace    string            `json:"namespace"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Completions  *int32            `json:"completions,omitempty"`
	Parallelism  *int32            `json:"parallelism,omitempty"`
	BackoffLimit *int32            `json:"backoffLimit,omitempty"`
	Containers   []ContainerSpec   `json:"containers" binding:"required"`
	RestartPolicy string           `json:"restartPolicy,omitempty"`
}

// ListJobs 获取 Job 列表
func (s *JobService) ListJobs(ctx context.Context, clusterName, namespace string) ([]JobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *batchv1.JobList
	if namespace == "all" || namespace == "" {
		list, err = client.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 Job 列表失败: %w", err)
	}
	result := make([]JobInfo, 0, len(list.Items))
	for _, job := range list.Items {
		result = append(result, convertJobInfo(&job))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetJob 获取 Job 详情
func (s *JobService) GetJob(ctx context.Context, clusterName, namespace, name string) (*JobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	job, err := client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Job 失败: %w", err)
	}
	info := convertJobInfo(job)
	return &info, nil
}

// CreateJob 创建 Job
func (s *JobService) CreateJob(ctx context.Context, clusterName, namespace string, req CreateJobRequest) (*JobInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	labels := req.Labels
	if labels == nil {
		labels = map[string]string{"job-name": req.Name}
	}
	containers := make([]corev1.Container, 0, len(req.Containers))
	for _, c := range req.Containers {
		containers = append(containers, buildContainerFromSpec(c))
	}
	restartPolicy := corev1.RestartPolicyNever
	if req.RestartPolicy != "" {
		restartPolicy = corev1.RestartPolicy(req.RestartPolicy)
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: req.Annotations,
		},
		Spec: batchv1.JobSpec{
			Completions:  req.Completions,
			Parallelism:  req.Parallelism,
			BackoffLimit: req.BackoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					RestartPolicy: restartPolicy,
					Containers:    containers,
				},
			},
		},
	}
	created, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 Job 失败: %w", err)
	}
	info := convertJobInfo(created)
	return &info, nil
}

// DeleteJob 删除 Job
func (s *JobService) DeleteJob(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	return client.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func convertJobInfo(job *batchv1.Job) JobInfo {
	images := make([]string, 0)
	for _, c := range job.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	info := JobInfo{
		Name:          job.Name,
		Namespace:     job.Namespace,
		Completions:   job.Spec.Completions,
		Parallelism:   job.Spec.Parallelism,
		Succeeded:     job.Status.Succeeded,
		Failed:        job.Status.Failed,
		Active:        job.Status.Active,
		CreationTime:  job.CreationTimestamp.Time,
		Labels:        job.Labels,
		Images:        images,
	}
	if job.Status.StartTime != nil {
		t := job.Status.StartTime.Time
		info.StartTime = &t
	}
	if job.Status.CompletionTime != nil {
		t := job.Status.CompletionTime.Time
		info.CompletionTime = &t
	}
	return info
}
