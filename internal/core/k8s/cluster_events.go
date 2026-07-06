package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventFilterOptions 集群事件过滤选项
type EventFilterOptions struct {
	Namespace          string
	Kind               string
	Type               string
	Reason             string
	InvolvedObjectName string
	Limit              int
}

// ClusterEventService 集群事件服务
type ClusterEventService struct {
	clientManager *ClientManager
}

// NewClusterEventService 创建集群事件服务
func NewClusterEventService(clientManager *ClientManager) *ClusterEventService {
	return &ClusterEventService{clientManager: clientManager}
}

// ListClusterEvents 获取过滤后的集群事件
func (s *ClusterEventService) ListClusterEvents(ctx context.Context, clusterName string, filter EventFilterOptions) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	return listFilteredEvents(ctx, client, filter)
}

func listFilteredEvents(ctx context.Context, client *kubernetes.Clientset, filter EventFilterOptions) ([]corev1.Event, error) {
	listOpts := metav1.ListOptions{}
	var selectors []string
	if filter.InvolvedObjectName != "" {
		selectors = append(selectors, "involvedObject.name="+filter.InvolvedObjectName)
	}
	if filter.Kind != "" {
		selectors = append(selectors, "involvedObject.kind="+filter.Kind)
	}
	if len(selectors) > 0 {
		listOpts.FieldSelector = strings.Join(selectors, ",")
	}

	namespace := filter.Namespace
	if namespace == "" {
		namespace = ""
	}

	events, err := client.CoreV1().Events(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("获取集群事件失败: %w", err)
	}

	items := events.Items
	if filter.Type != "" {
		filtered := make([]corev1.Event, 0)
		for _, e := range items {
			if string(e.Type) == filter.Type {
				filtered = append(filtered, e)
			}
		}
		items = filtered
	}
	if filter.Reason != "" {
		filtered := make([]corev1.Event, 0)
		for _, e := range items {
			if e.Reason == filter.Reason {
				filtered = append(filtered, e)
			}
		}
		items = filtered
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastTimestamp.After(items[j].LastTimestamp.Time)
	})

	if filter.Limit > 0 && len(items) > filter.Limit {
		items = items[:filter.Limit]
	}
	return items, nil
}
