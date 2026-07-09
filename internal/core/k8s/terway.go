package k8s

import (
	"context"
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterNetworkInfo 集群 CNI / 网络可观测性信息
type ClusterNetworkInfo struct {
	CNI              string `json:"cni"`
	TerwayMode       string `json:"terwayMode,omitempty"`
	HubbleEnabled    bool   `json:"hubbleEnabled,omitempty"`
	HubbleRelayReady bool   `json:"hubbleRelayReady,omitempty"`
	HubbleMetricsSvc bool   `json:"hubbleMetricsSvc,omitempty"`
	MetricsSource    string `json:"metricsSource,omitempty"`
	Message          string `json:"message,omitempty"`
}

// HubbleDropStat Hubble 丢包统计
type HubbleDropStat struct {
	Reason string  `json:"reason"`
	Count  float64 `json:"count"`
}

// HubblePortStat Hubble 端口分布
type HubblePortStat struct {
	Protocol string  `json:"protocol"`
	Port     string  `json:"port"`
	Count    float64 `json:"count"`
}

// HubbleMetricsSummary Hubble L3/L4 指标摘要
type HubbleMetricsSummary struct {
	Available bool             `json:"available"`
	Drops     []HubbleDropStat `json:"drops,omitempty"`
	TopPorts  []HubblePortStat `json:"topPorts,omitempty"`
	Message   string           `json:"message,omitempty"`
}

type terwayConfig struct {
	ENIIPVirtualType          string `json:"eniip_virtual_type"`
	CiliumEnableHubble        string `json:"cilium_enable_hubble"`
	CiliumHubbleMetrics       string `json:"cilium_hubble_metrics"`
	CiliumHubbleListenAddress string `json:"cilium_hubble_listen_address"`
}

func detectClusterNetwork(ctx context.Context, client *kubernetes.Clientset) *ClusterNetworkInfo {
	info := &ClusterNetworkInfo{CNI: "unknown", Message: "cni_unknown"}

	if isTerwayCluster(ctx, client) {
		info.CNI = "terway"
		parseTerwayConfig(ctx, client, info)
	}
	checkHubbleComponents(ctx, client, info)
	setNetworkMessage(info)
	return info
}

func isTerwayCluster(ctx context.Context, client *kubernetes.Clientset) bool {
	knownDS := []string{"terway-eniip", "terway-eni", "terway", "terway-veth", "terway-controlplane"}
	for _, name := range knownDS {
		if _, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, name, metav1.GetOptions{}); err == nil {
			return true
		}
	}

	if list, err := client.AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{}); err == nil {
		for _, ds := range list.Items {
			if strings.Contains(strings.ToLower(ds.Name), "terway") {
				return true
			}
		}
	}

	for _, cmName := range []string{"eni-config", "terway-config"} {
		if _, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, cmName, metav1.GetOptions{}); err == nil {
			return true
		}
	}

	for _, sel := range []string{
		"app=terway-eniip",
		"app.kubernetes.io/name=terway-eniip",
		"k8s-app=terway-eniip",
	} {
		if pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{LabelSelector: sel}); err == nil && len(pods.Items) > 0 {
			return true
		}
	}
	return false
}

func parseTerwayConfig(ctx context.Context, client *kubernetes.Clientset, info *ClusterNetworkInfo) {
	for _, cmName := range []string{"eni-config", "terway-config"} {
		cm, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, cmName, metav1.GetOptions{})
		if err != nil {
			continue
		}
		for _, key := range []string{"10-terway.conf", "config", "terway.conf"} {
			raw := cm.Data[key]
			if raw == "" {
				continue
			}
			var cfg terwayConfig
			if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
				continue
			}
			switch strings.ToUpper(cfg.ENIIPVirtualType) {
			case "IPVLAN":
				info.TerwayMode = "ipvlan"
			default:
				if cfg.ENIIPVirtualType != "" {
					info.TerwayMode = strings.ToLower(cfg.ENIIPVirtualType)
				}
			}
			if strings.EqualFold(cfg.CiliumEnableHubble, "true") {
				info.HubbleEnabled = true
			}
			return
		}
	}
}

func checkHubbleComponents(ctx context.Context, client *kubernetes.Clientset, info *ClusterNetworkInfo) {
	for _, name := range []string{"hubble-metrics", "cilium-agent"} {
		if svc, err := client.CoreV1().Services("kube-system").Get(ctx, name, metav1.GetOptions{}); err == nil && svc != nil {
			info.HubbleMetricsSvc = true
			if name == "hubble-metrics" {
				info.HubbleEnabled = true
			}
			break
		}
	}
	for _, name := range []string{"hubble-relay", "ack-terway-hubble-relay", "hubble-ui"} {
		dep, err := client.AppsV1().Deployments("kube-system").Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			continue
		}
		if dep.Status.ReadyReplicas > 0 {
			info.HubbleRelayReady = true
			info.HubbleEnabled = true
			break
		}
	}
}

func setNetworkMessage(info *ClusterNetworkInfo) {
	if info.MetricsSource != "" {
		return
	}
	switch {
	case info.CNI == "terway" && info.HubbleMetricsSvc:
		info.Message = "terway_observable"
	case info.CNI == "terway" && info.HubbleEnabled && info.HubbleRelayReady:
		info.Message = "terway_hubble_ready"
	case info.CNI == "terway":
		info.Message = "terway_detected"
	case info.HubbleMetricsSvc:
		info.Message = "hubble_available"
	default:
		info.Message = "cni_unknown"
	}
}
