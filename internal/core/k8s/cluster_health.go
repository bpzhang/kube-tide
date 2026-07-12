package k8s

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ComponentHealth 集群组件健康状态
type ComponentHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // healthy, degraded, missing, unknown
	Message string `json:"message,omitempty"`
}

// ClusterHealthReport 集群健康报告
type ClusterHealthReport struct {
	Overall    string            `json:"overall"` // healthy, degraded, unhealthy
	Components []ComponentHealth `json:"components"`
}

// CheckClusterHealth 检查集群关键组件状态
func CheckClusterHealth(ctx context.Context, client *kubernetes.Clientset) (*ClusterHealthReport, error) {
	if _, err := client.ServerVersion(); err != nil {
		return nil, fmt.Errorf("API Server 不可达: %w", err)
	}

	components := []ComponentHealth{
		checkAPIServer(ctx, client),
		checkDNS(ctx, client),
		checkMetricsServer(ctx, client),
		checkNodes(ctx, client),
	}

	overall := "healthy"
	hasDegraded := false
	hasUnhealthy := false
	for _, c := range components {
		switch c.Status {
		case "degraded":
			hasDegraded = true
		case "missing", "unhealthy":
			hasUnhealthy = true
		}
	}
	if hasUnhealthy {
		overall = "unhealthy"
	} else if hasDegraded {
		overall = "degraded"
	}

	return &ClusterHealthReport{Overall: overall, Components: components}, nil
}

func checkAPIServer(ctx context.Context, client *kubernetes.Clientset) ComponentHealth {
	version, err := client.ServerVersion()
	if err != nil {
		return ComponentHealth{Name: "api-server", Status: "unhealthy", Message: err.Error()}
	}
	return ComponentHealth{Name: "api-server", Status: "healthy", Message: version.String()}
}

func checkDNS(ctx context.Context, client *kubernetes.Clientset) ComponentHealth {
	names := []string{"coredns", "kube-dns"}
	for _, name := range names {
		if dep, err := client.AppsV1().Deployments("kube-system").Get(ctx, name, metav1.GetOptions{}); err == nil {
			return workloadHealth(name, dep.Status.ReadyReplicas, dep.Status.Replicas)
		}
	}
	// CoreDNS may run as DaemonSet on some clusters
	if ds, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "coredns", metav1.GetOptions{}); err == nil {
		return daemonSetHealth("coredns", ds)
	}
	return ComponentHealth{Name: "dns", Status: "missing", Message: "未找到 CoreDNS / kube-dns"}
}

func checkMetricsServer(ctx context.Context, client *kubernetes.Clientset) ComponentHealth {
	namespaces := []string{"kube-system", "metrics-server"}
	for _, ns := range namespaces {
		deployments, err := client.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		for i := range deployments.Items {
			dep := &deployments.Items[i]
			if strings.Contains(strings.ToLower(dep.Name), "metrics-server") {
				return workloadHealth(dep.Name, dep.Status.ReadyReplicas, dep.Status.Replicas)
			}
		}
	}
	return ComponentHealth{Name: "metrics-server", Status: "missing", Message: "未找到 metrics-server"}
}

func checkNodes(ctx context.Context, client *kubernetes.Clientset) ComponentHealth {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return ComponentHealth{Name: "nodes", Status: "unknown", Message: err.Error()}
	}
	ready, notReady := 0, 0
	for _, node := range nodes.Items {
		if nodeReady(&node) {
			ready++
		} else {
			notReady++
		}
	}
	if len(nodes.Items) == 0 {
		return ComponentHealth{Name: "nodes", Status: "missing", Message: "集群无节点"}
	}
	if notReady > 0 {
		return ComponentHealth{
			Name: "nodes", Status: "degraded",
			Message: fmt.Sprintf("%d/%d 节点 Ready", ready, len(nodes.Items)),
		}
	}
	return ComponentHealth{
		Name: "nodes", Status: "healthy",
		Message: fmt.Sprintf("%d/%d 节点 Ready", ready, len(nodes.Items)),
	}
}

func workloadHealth(name string, ready, total int32) ComponentHealth {
	if total == 0 {
		return ComponentHealth{Name: name, Status: "degraded", Message: "无副本"}
	}
	if ready < total {
		return ComponentHealth{
			Name: name, Status: "degraded",
			Message: fmt.Sprintf("%d/%d Ready", ready, total),
		}
	}
	return ComponentHealth{Name: name, Status: "healthy", Message: fmt.Sprintf("%d/%d Ready", ready, total)}
}

func daemonSetHealth(name string, ds *appsv1.DaemonSet) ComponentHealth {
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady
	if desired == 0 {
		return ComponentHealth{Name: name, Status: "degraded", Message: "无调度副本"}
	}
	if ready < desired {
		return ComponentHealth{
			Name: name, Status: "degraded",
			Message: fmt.Sprintf("%d/%d Ready", ready, desired),
		}
	}
	return ComponentHealth{Name: name, Status: "healthy", Message: fmt.Sprintf("%d/%d Ready", ready, desired)}
}

func nodeReady(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}
