package k8s

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const maxPrometheusQueryLen = 4096

// MaxPrometheusQueryLen is the maximum allowed PromQL query length.
func MaxPrometheusQueryLen() int {
	return maxPrometheusQueryLen
}

// ValidatePrometheusURL checks that a Prometheus base URL is safe to proxy.
func ValidatePrometheusURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("无效的 Prometheus URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("Prometheus URL 仅支持 http/https")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("Prometheus URL 缺少主机名")
	}
	if isBlockedPrometheusHost(host) {
		return fmt.Errorf("Prometheus URL 不允许指向内网或本地地址")
	}
	return nil
}

func isBlockedPrometheusHost(host string) bool {
	lower := strings.ToLower(strings.TrimSpace(host))
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return true
	}
	// Kubernetes cluster DNS is allowed (e.g. prometheus.monitoring.svc.cluster.local)
	if strings.HasSuffix(lower, ".cluster.local") || strings.HasSuffix(lower, ".svc") {
		return false
	}
	// Block mDNS *.local hostnames (e.g. mydevice.local), but not *.cluster.local above
	if strings.HasSuffix(lower, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.Equal(net.IPv4(169, 254, 169, 254))
}
