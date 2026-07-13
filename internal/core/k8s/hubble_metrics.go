package k8s

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

func fetchHubbleMetrics(
	ctx context.Context,
	client *kubernetes.Clientset,
	prom *PrometheusService,
	clusterName, namespace string,
	network *ClusterNetworkInfo,
) *HubbleMetricsSummary {
	summary := &HubbleMetricsSummary{}
	_ = namespace
	if network == nil {
		network = &ClusterNetworkInfo{}
	}

	// 1. ACK Terway 内置 hubble-metrics
	if IsACKTerwayObservable(network) {
		if body, ok := FetchACKTerwayHubbleMetrics(ctx, client); ok {
			samples := parsePrometheusText(body)
			summary.Drops = topHubbleDrops(samples, 8)
			summary.TopPorts = topHubblePorts(samples, 10)
			if len(summary.Drops) > 0 || len(summary.TopPorts) > 0 {
				summary.Available = true
				summary.Message = MetricsSourceACKTerwayHubble
				network.MetricsSource = MetricsSourceACKTerwayHubble
				network.HubbleMetricsSvc = true
				return summary
			}
		}
	}

	// 2. ACK 托管 Prometheus 中已采集的 Hubble 指标
	if prom != nil {
		if drops, ports, ok := fetchHubbleViaACKPrometheus(ctx, prom, clusterName); ok {
			summary.Drops = drops
			summary.TopPorts = ports
			summary.Available = true
			summary.Message = MetricsSourceACKManagedPrometheus
			network.MetricsSource = MetricsSourceACKManagedPrometheus
			return summary
		}
	}

	if network.CNI == "terway" && !network.HubbleMetricsSvc && !network.HubbleEnabled {
		summary.Message = "terway_enable_hubble"
	} else if network.HubbleMetricsSvc || network.HubbleEnabled {
		summary.Message = "hubble_no_data"
	}
	return summary
}

func fetchHubbleViaACKPrometheus(ctx context.Context, prom *PrometheusService, clusterName string) ([]HubbleDropStat, []HubblePortStat, bool) {
	ackURL, err := prom.ResolveACKPrometheusURL(ctx, clusterName)
	if err != nil || ackURL == "" {
		return nil, nil, false
	}
	return fetchHubbleViaPrometheusOnURL(ctx, prom, ackURL)
}

func fetchHubbleViaPrometheus(ctx context.Context, prom *PrometheusService, clusterName string) ([]HubbleDropStat, []HubblePortStat, bool) {
	ackURL, err := prom.ResolveACKPrometheusURL(ctx, clusterName)
	if err != nil || ackURL == "" {
		return nil, nil, false
	}
	return fetchHubbleViaPrometheusOnURL(ctx, prom, ackURL)
}

func fetchHubbleViaPrometheusOnURL(ctx context.Context, prom *PrometheusService, promURL string) ([]HubbleDropStat, []HubblePortStat, bool) {
	dropQuery := `topk(8, sum by (reason) (increase(hubble_drop_total[5m])))`
	portQuery := `topk(10, sum by (protocol, port) (increase(hubble_port_distribution_total[5m])))`

	dropsRaw, err := prom.QueryInstantOnURL(ctx, promURL, dropQuery, 15*time.Second)
	if err != nil {
		return nil, nil, false
	}
	portsRaw, _ := prom.QueryInstantOnURL(ctx, promURL, portQuery, 15*time.Second)

	drops := parseVectorMetrics(dropsRaw, "reason", "")
	ports := parsePortMetrics(portsRaw)
	if len(drops) == 0 && len(ports) == 0 {
		return nil, nil, false
	}
	return drops, ports, true
}

type promInstantResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func parseVectorMetrics(raw json.RawMessage, labelKey, secondaryKey string) []HubbleDropStat {
	var resp promInstantResponse
	if err := json.Unmarshal(raw, &resp); err != nil || resp.Status != "success" {
		return nil
	}
	stats := make([]HubbleDropStat, 0, len(resp.Data.Result))
	for _, item := range resp.Data.Result {
		val := instantValue(item.Value)
		if val <= 0 {
			continue
		}
		reason := item.Metric[labelKey]
		if secondaryKey != "" {
			reason = strings.TrimSpace(reason + " " + item.Metric[secondaryKey])
		}
		if reason == "" {
			reason = "unknown"
		}
		stats = append(stats, HubbleDropStat{Reason: reason, Count: val})
	}
	return stats
}

func parsePortMetrics(raw json.RawMessage) []HubblePortStat {
	var resp promInstantResponse
	if err := json.Unmarshal(raw, &resp); err != nil || resp.Status != "success" {
		return nil
	}
	stats := make([]HubblePortStat, 0, len(resp.Data.Result))
	for _, item := range resp.Data.Result {
		val := instantValue(item.Value)
		if val <= 0 {
			continue
		}
		stats = append(stats, HubblePortStat{
			Protocol: item.Metric["protocol"],
			Port:     item.Metric["port"],
			Count:    val,
		})
	}
	return stats
}

func instantValue(pair []any) float64 {
	if len(pair) < 2 {
		return 0
	}
	switch v := pair[1].(type) {
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case float64:
		return v
	default:
		return 0
	}
}
