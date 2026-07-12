package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PrometheusInfo Prometheus 连接信息
type PrometheusInfo struct {
	Configured bool   `json:"configured"`
	URL        string `json:"url,omitempty"`
	Source     string `json:"source"` // manual, discovered, unconfigured
	Healthy    bool   `json:"healthy"`
	Message    string `json:"message,omitempty"`
}

// PodHealthSummary Pod 健康汇总
type PodHealthSummary struct {
	Total     int `json:"total"`
	Running   int `json:"running"`
	Pending   int `json:"pending"`
	Failed    int `json:"failed"`
	CrashLoop int `json:"crashLoop"`
	NotReady  int `json:"notReady"`
}

// EventSummary 事件汇总
type EventSummary struct {
	WarningRecent int `json:"warningRecent"`
}

// ObservabilitySummary 可观测性总览
type ObservabilitySummary struct {
	Prometheus PrometheusInfo        `json:"prometheus"`
	Health     *ClusterHealthReport  `json:"health,omitempty"`
	Pods       PodHealthSummary      `json:"pods"`
	Events     EventSummary          `json:"events"`
}

// ObservabilityAlert 可观测性告警项
type ObservabilityAlert struct {
	Level     string `json:"level"` // warning, error, info
	Category  string `json:"category"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Message   string `json:"message"`
}

// ObservabilityService 可观测性服务
type ObservabilityService struct {
	clientManager *ClientManager
	prometheus    *PrometheusService
}

// NewObservabilityService 创建可观测性服务
func NewObservabilityService(clientManager *ClientManager, prometheus *PrometheusService) *ObservabilityService {
	return &ObservabilityService{clientManager: clientManager, prometheus: prometheus}
}

// GetPrometheusInfo 获取 Prometheus 配置与连通性
func (s *ObservabilityService) GetPrometheusInfo(ctx context.Context, clusterName string) (*PrometheusInfo, error) {
	info := &PrometheusInfo{Source: "unconfigured"}
	manual := s.clientManager.GetPrometheusURL(clusterName)
	if manual != "" {
		info.Configured = true
		info.URL = manual
		info.Source = "manual"
	} else {
		client, err := s.clientManager.GetClient(clusterName)
		if err != nil {
			return nil, err
		}
		if discovered := DiscoverPrometheusURL(ctx, client); discovered != "" {
			info.Configured = true
			info.URL = discovered
			info.Source = "discovered"
		}
	}
	if !info.Configured {
		info.Message = "prometheus_not_configured"
		return info, nil
	}
	if err := ValidatePrometheusURL(info.URL); err != nil {
		info.Message = err.Error()
		return info, nil
	}
	_, err := s.prometheus.QueryInstant(ctx, clusterName, "up", 5*time.Second)
	if err != nil {
		info.Message = err.Error()
		return info, nil
	}
	info.Healthy = true
	info.Message = "ok"
	return info, nil
}

// GetSummary 获取可观测性总览
func (s *ObservabilityService) GetSummary(ctx context.Context, clusterName string) (*ObservabilitySummary, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	promInfo, err := s.GetPrometheusInfo(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	health, _ := CheckClusterHealth(ctx, client)
	pods, err := summarizePods(ctx, client)
	if err != nil {
		return nil, err
	}
	events, err := summarizeWarningEvents(ctx, client)
	if err != nil {
		return nil, err
	}

	return &ObservabilitySummary{
		Prometheus: *promInfo,
		Health:     health,
		Pods:       pods,
		Events:     events,
	}, nil
}

// ListAlerts 列出集群可观测性告警
func (s *ObservabilityService) ListAlerts(ctx context.Context, clusterName string, limit int) ([]ObservabilityAlert, error) {
	if limit <= 0 {
		limit = 50
	}
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	alerts := make([]ObservabilityAlert, 0)

	if health, err := CheckClusterHealth(ctx, client); err == nil {
		for _, c := range health.Components {
			if c.Status == "healthy" {
				continue
			}
			level := "warning"
			if c.Status == "unhealthy" || c.Status == "missing" {
				level = "error"
			}
			alerts = append(alerts, ObservabilityAlert{
				Level: level, Category: "component", Name: c.Name, Message: c.Message,
			})
		}
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		for i := range nodes.Items {
			node := &nodes.Items[i]
			if !nodeReady(node) {
				alerts = append(alerts, ObservabilityAlert{
					Level: "error", Category: "node", Name: node.Name,
					Message: "节点 NotReady",
				})
			}
		}
	}

	podList, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}
	for i := range podList.Items {
		pod := &podList.Items[i]
		if msg := podProblemMessage(pod); msg != "" {
			alerts = append(alerts, ObservabilityAlert{
				Level: podAlertLevel(pod), Category: "pod",
				Namespace: pod.Namespace, Name: pod.Name, Message: msg,
			})
		}
	}

	deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for i := range deployments.Items {
			dep := &deployments.Items[i]
			if dep.Status.Replicas > 0 && dep.Status.AvailableReplicas < dep.Status.Replicas {
				alerts = append(alerts, ObservabilityAlert{
					Level: "warning", Category: "deployment",
					Namespace: dep.Namespace, Name: dep.Name,
					Message: fmt.Sprintf("可用副本 %d/%d", dep.Status.AvailableReplicas, dep.Status.Replicas),
				})
			}
		}
	}

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}
	return alerts, nil
}

func summarizePods(ctx context.Context, client *kubernetes.Clientset) (PodHealthSummary, error) {
	list, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return PodHealthSummary{}, err
	}
	summary := PodHealthSummary{Total: len(list.Items)}
	for i := range list.Items {
		pod := &list.Items[i]
		switch pod.Status.Phase {
		case corev1.PodRunning:
			summary.Running++
		case corev1.PodPending:
			summary.Pending++
		case corev1.PodFailed:
			summary.Failed++
		}
		if isCrashLoop(pod) {
			summary.CrashLoop++
		}
		if pod.Status.Phase == corev1.PodRunning && !podIsReady(pod) {
			summary.NotReady++
		}
	}
	return summary, nil
}

func summarizeWarningEvents(ctx context.Context, client *kubernetes.Clientset) (EventSummary, error) {
	events, err := client.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return EventSummary{}, err
	}
	cutoff := time.Now().Add(-1 * time.Hour)
	count := 0
	for i := range events.Items {
		e := &events.Items[i]
		if e.Type != corev1.EventTypeWarning {
			continue
		}
		if e.LastTimestamp.After(cutoff) {
			count++
		}
	}
	return EventSummary{WarningRecent: count}, nil
}

func isCrashLoop(pod *corev1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
			return true
		}
		if cs.RestartCount > 5 {
			return true
		}
	}
	return false
}

func podIsReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func podProblemMessage(pod *corev1.Pod) string {
	if pod.Status.Phase == corev1.PodFailed {
		return "Pod Failed"
	}
	if isCrashLoop(pod) {
		return "CrashLoopBackOff"
	}
	if pod.Status.Phase == corev1.PodPending {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
				return cond.Reason + ": " + cond.Message
			}
		}
		return "Pod Pending"
	}
	if pod.Status.Phase == corev1.PodRunning && !podIsReady(pod) {
		return "Pod NotReady"
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			reason := cs.State.Waiting.Reason
			if reason != "ContainerCreating" {
				return reason
			}
		}
	}
	return ""
}

func podAlertLevel(pod *corev1.Pod) string {
	if pod.Status.Phase == corev1.PodFailed || isCrashLoop(pod) {
		return "error"
	}
	if strings.Contains(podProblemMessage(pod), "ImagePull") {
		return "error"
	}
	return "warning"
}
