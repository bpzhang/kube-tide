package k8s

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ACK 可观测性数据源（无需额外部署中间件）
const (
	MetricsSourceACKTerwayHubble      = "ack_terway_hubble"
	MetricsSourceACKManagedPrometheus = "ack_managed_prometheus"
	MetricsSourceK8sInferred          = "k8s_inferred"
	MetricsSourceUnconfigured         = "unconfigured"
)

// 阿里云 ACK 容器监控 / ARMS Prometheus 所在命名空间
var ackPrometheusNamespaces = []string{
	"arms-prom",
	"ack-cmonitor",
	"o11y-system",
}

// DiscoverACKPrometheusURL 自动发现 ACK 托管 Prometheus 查询端点（arms-prom 等）
func DiscoverACKPrometheusURL(ctx context.Context, client *kubernetes.Clientset) string {
	for _, ns := range ackPrometheusNamespaces {
		list, err := client.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		for i := range list.Items {
			svc := &list.Items[i]
			if !isACKPrometheusService(svc) {
				continue
			}
			if url := ackPrometheusURLFromService(svc); url != "" {
				if ackPrometheusHealthy(ctx, client, svc) {
					return url
				}
			}
		}
	}
	return ""
}

func isACKPrometheusService(svc *corev1.Service) bool {
	name := strings.ToLower(svc.Name)
	if strings.Contains(name, "arms-prometheus") || strings.Contains(name, "ack-prometheus") {
		return true
	}
	if strings.Contains(name, "prometheus") && svc.Namespace == "arms-prom" {
		return true
	}
	if strings.Contains(name, "cmonitor") || strings.Contains(name, "o11y") {
		return true
	}
	return false
}

func ackPrometheusURLFromService(svc *corev1.Service) string {
	port := prometheusServicePort(svc)
	if port == 0 {
		return ""
	}
	return serviceClusterURL(svc.Namespace, svc.Name, port)
}

func ackPrometheusHealthy(ctx context.Context, client *kubernetes.Clientset, svc *corev1.Service) bool {
	return prometheusHealthy(ctx, client, svc)
}

func prometheusServicePort(svc *corev1.Service) int32 {
	if p := svc.Annotations["prometheus.io/port"]; p != "" {
		for _, sp := range svc.Spec.Ports {
			if sp.Name == p || strconv.Itoa(int(sp.Port)) == p {
				return sp.Port
			}
		}
	}
	for _, sp := range svc.Spec.Ports {
		if sp.Port == 9090 || sp.Name == "http" || sp.Name == "web" {
			return sp.Port
		}
	}
	if len(svc.Spec.Ports) > 0 {
		return svc.Spec.Ports[0].Port
	}
	return 0
}

func prometheusHealthy(ctx context.Context, client *kubernetes.Clientset, svc *corev1.Service) bool {
	port := strconv.Itoa(int(prometheusServicePort(svc)))
	for _, path := range []string{"/-/ready", "/-/healthy", "/api/v1/status/buildinfo"} {
		if _, err := proxyServiceGET(ctx, client, svc.Namespace, svc.Name, port, path); err == nil {
			return true
		}
	}
	return false
}

// FetchACKTerwayHubbleMetrics 读取 ACK Terway 内置 hubble-metrics 原始指标
func FetchACKTerwayHubbleMetrics(ctx context.Context, client *kubernetes.Clientset) ([]byte, bool) {
	for _, port := range []string{"9091", "9965"} {
		body, err := proxyServiceGET(ctx, client, "kube-system", "hubble-metrics", port, "/metrics")
		if err == nil && len(body) > 0 {
			return body, true
		}
	}
	return nil, false
}

// IsACKTerwayObservable 集群是否具备 ACK Terway 网络可观测能力
func IsACKTerwayObservable(network *ClusterNetworkInfo) bool {
	if network == nil {
		return false
	}
	return network.CNI == "terway" && (network.HubbleMetricsSvc || network.HubbleEnabled)
}

// DiscoverPrometheusURL 兼容入口：仅发现 ACK 托管 Prometheus
func DiscoverPrometheusURL(ctx context.Context, client *kubernetes.Clientset) string {
	return DiscoverACKPrometheusURL(ctx, client)
}
