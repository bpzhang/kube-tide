package k8s

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var prometheusNamespaces = []string{
	"arms-prom", "monitoring", "kube-system", "o11y-system", "prometheus", "ack-cmonitor",
}

// DiscoverPrometheusURL 从 ACK/集群内自动发现 Prometheus 查询端点
func DiscoverPrometheusURL(ctx context.Context, client *kubernetes.Clientset) string {
	for _, ns := range prometheusNamespaces {
		list, err := client.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		for i := range list.Items {
			if url := prometheusURLFromService(&list.Items[i]); url != "" {
				if prometheusHealthy(ctx, client, &list.Items[i]) {
					return url
				}
			}
		}
	}
	return ""
}

func prometheusURLFromService(svc *corev1.Service) string {
	name := strings.ToLower(svc.Name)
	if !strings.Contains(name, "prometheus") && svc.Annotations["prometheus.io/scrape"] != "true" {
		return ""
	}
	port := prometheusServicePort(svc)
	if port == 0 {
		return ""
	}
	return serviceClusterURL(svc.Namespace, svc.Name, port)
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
