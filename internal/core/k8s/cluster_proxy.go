package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

func proxyServiceGET(ctx context.Context, client *kubernetes.Clientset, namespace, name, port, path string) ([]byte, error) {
	if path == "" {
		path = "/"
	}
	return client.CoreV1().Services(namespace).ProxyGet(
		"http", name, port, path, nil,
	).DoRaw(ctx)
}

func serviceClusterURL(namespace, name string, port int32) string {
	if port <= 0 {
		port = 9090
	}
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", name, namespace, port)
}
