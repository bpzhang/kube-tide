package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// EdgeMetrics 边上的可观测指标
type EdgeMetrics struct {
	FlowsPerSec float64 `json:"flowsPerSec,omitempty"`
	BytesPerSec float64 `json:"bytesPerSec,omitempty"`
	Drops       float64 `json:"drops,omitempty"`
	Observed    bool    `json:"observed"`
}

// CallFlowStat 服务间调用流量（Hubble 观测或推断）
type CallFlowStat struct {
	SourceNS      string  `json:"sourceNamespace"`
	Source        string  `json:"source"`
	SourceType    string  `json:"sourceType,omitempty"`
	TargetNS      string  `json:"targetNamespace"`
	Target        string  `json:"target"`
	TargetType    string  `json:"targetType,omitempty"`
	Port          string  `json:"port,omitempty"`
	Protocol      string  `json:"protocol,omitempty"`
	FlowsPerSec   float64 `json:"flowsPerSec"`
	BytesPerSec   float64 `json:"bytesPerSec,omitempty"`
	Drops         float64 `json:"drops,omitempty"`
	Observed      bool    `json:"observed"`
	Evidence      string  `json:"evidence,omitempty"`
}

// CallChainHop 调用链一跳
type CallChainHop struct {
	NodeID    string `json:"nodeId"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// CallChainPath 完整调用链路径
type CallChainPath struct {
	ID         string         `json:"id"`
	Entry      string         `json:"entry,omitempty"`
	Hops       []CallChainHop `json:"hops"`
	Depth      int            `json:"depth"`
	Observed   bool           `json:"observed"`
	FlowsPerSec float64       `json:"flowsPerSec,omitempty"`
}

type flowEndpoint struct {
	namespace string
	name      string
	kind      string // service, deployment, etc.
	nodeID    string
}

type hubbleFlowSample struct {
	sourceNS, sourceName string
	destNS, destName     string
	port, protocol       string
	flowsPerSec          float64
	bytesPerSec          float64
	drops                float64
}

func enrichCallChainObservability(
	ctx context.Context,
	client *kubernetes.Clientset,
	prom *PrometheusService,
	clusterName string,
	topology *TrafficTopology,
	serviceByKey map[string]corev1.Service,
	workloadByPod map[string]workloadRef,
) {
	if topology == nil {
		return
	}

	flows, flowSource := fetchCallFlows(ctx, client, prom, clusterName, topology.Network)
	resolved := resolveFlowEndpoints(flows, serviceByKey, workloadByPod, flowSource)
	topology.CallFlows = resolved

	enrichEdgeMetrics(topology, resolved)
	topology.CallChains = buildCallChains(topology)
}

func fetchCallFlows(ctx context.Context, client *kubernetes.Clientset, prom *PrometheusService, clusterName string, network *ClusterNetworkInfo) ([]hubbleFlowSample, string) {
	// 1. ACK Terway 内置 hubble-metrics（kube-system，无需额外组件）
	if IsACKTerwayObservable(network) {
		if flows := fetchFlowsFromACKHubbleMetrics(ctx, client); len(flows) > 0 {
			return flows, MetricsSourceACKTerwayHubble
		}
	}
	// 2. ACK 托管 Prometheus 中已采集的 Hubble 指标
	if prom != nil {
		if flows, ok := fetchFlowsViaACKPrometheus(ctx, prom, clusterName); ok {
			return flows, MetricsSourceACKManagedPrometheus
		}
	}
	return nil, ""
}

func fetchFlowsFromACKHubbleMetrics(ctx context.Context, client *kubernetes.Clientset) []hubbleFlowSample {
	body, ok := FetchACKTerwayHubbleMetrics(ctx, client)
	if !ok {
		return nil
	}
	return parseFlowsFromTextSamples(parsePrometheusText(body))
}

func fetchFlowsViaACKPrometheus(ctx context.Context, prom *PrometheusService, clusterName string) ([]hubbleFlowSample, bool) {
	ackURL, err := prom.ResolveACKPrometheusURL(ctx, clusterName)
	if err != nil || ackURL == "" {
		return nil, false
	}
	return fetchFlowsViaPrometheusOnURL(ctx, prom, ackURL)
}

func fetchFlowsViaPrometheusOnURL(ctx context.Context, prom *PrometheusService, promURL string) ([]hubbleFlowSample, bool) {
	flowQuery := `sum by (source, destination) (rate(hubble_flows_processed_total{verdict="FORWARDED"}[5m]))`
	dropQuery := `sum by (source, destination) (rate(hubble_drop_total[5m]))`

	flowRaw, err := prom.QueryInstantOnURL(ctx, promURL, flowQuery, 15*time.Second)
	if err != nil {
		return nil, false
	}

	drops := parseFlowDropsOnURL(dropQuery, prom, ctx, promURL)
	flows := parseHubbleFlowMetrics(flowRaw, drops)
	return flows, len(flows) > 0
}

func parseFlowDropsOnURL(query string, prom *PrometheusService, ctx context.Context, promURL string) map[string]float64 {
	result := map[string]float64{}
	raw, err := prom.QueryInstantOnURL(ctx, promURL, query, 15*time.Second)
	if err != nil {
		return result
	}
	var resp promInstantResponse
	if err := json.Unmarshal(raw, &resp); err != nil || resp.Status != "success" {
		return result
	}
	for _, item := range resp.Data.Result {
		src := item.Metric["source"]
		dst := item.Metric["destination"]
		if src == "" || dst == "" {
			continue
		}
		result[src+"->"+dst] = instantValue(item.Value)
	}
	return result
}

func parseHubbleFlowMetrics(raw json.RawMessage, drops map[string]float64) []hubbleFlowSample {
	var resp promInstantResponse
	if err := json.Unmarshal(raw, &resp); err != nil || resp.Status != "success" {
		return nil
	}
	flows := make([]hubbleFlowSample, 0, len(resp.Data.Result))
	for _, item := range resp.Data.Result {
		src := item.Metric["source"]
		dst := item.Metric["destination"]
		if src == "" || dst == "" {
			continue
		}
		rate := instantValue(item.Value)
		if rate <= 0 {
			continue
		}
		srcNS, srcName := splitHubbleEndpoint(src)
		dstNS, dstName := splitHubbleEndpoint(dst)
		flows = append(flows, hubbleFlowSample{
			sourceNS: srcNS, sourceName: srcName,
			destNS: dstNS, destName: dstName,
			port: item.Metric["port"], protocol: item.Metric["protocol"],
			flowsPerSec: rate,
			drops:       drops[src+"->"+dst],
		})
	}
	return flows
}

func splitHubbleEndpoint(raw string) (namespace, name string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	if i := strings.Index(raw, "/"); i >= 0 {
		return raw[:i], raw[i+1:]
	}
	return "", raw
}

func parseFlowsFromTextSamples(samples []promSample) []hubbleFlowSample {
	flows := make([]hubbleFlowSample, 0)
	for _, s := range samples {
		if s.Name != "hubble_flows_processed_total" || s.Value <= 0 {
			continue
		}
		if s.Labels["verdict"] != "" && s.Labels["verdict"] != "FORWARDED" {
			continue
		}
		src := s.Labels["source"]
		dst := s.Labels["destination"]
		if src == "" || dst == "" {
			continue
		}
		srcNS, srcName := splitHubbleEndpoint(src)
		dstNS, dstName := splitHubbleEndpoint(dst)
		flows = append(flows, hubbleFlowSample{
			sourceNS: srcNS, sourceName: srcName,
			destNS: dstNS, destName: dstName,
			port: s.Labels["port"], protocol: s.Labels["protocol"],
			flowsPerSec: s.Value,
		})
	}
	return flows
}

func resolveFlowEndpoints(
	flows []hubbleFlowSample,
	serviceByKey map[string]corev1.Service,
	workloadByPod map[string]workloadRef,
	evidence string,
) []CallFlowStat {
	result := make([]CallFlowStat, 0, len(flows))
	for _, f := range flows {
		src := resolveEndpoint(f.sourceNS, f.sourceName, serviceByKey, workloadByPod)
		dst := resolveEndpoint(f.destNS, f.destName, serviceByKey, workloadByPod)
		if src.name == "" || dst.name == "" {
			continue
		}
		result = append(result, CallFlowStat{
			SourceNS: src.namespace, Source: src.name, SourceType: src.kind,
			TargetNS: dst.namespace, Target: dst.name, TargetType: dst.kind,
			Port: f.port, Protocol: f.protocol,
			FlowsPerSec: f.flowsPerSec, BytesPerSec: f.bytesPerSec, Drops: f.drops,
			Observed: true, Evidence: evidence,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FlowsPerSec > result[j].FlowsPerSec
	})
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func resolveEndpoint(ns, name string, serviceByKey map[string]corev1.Service, workloadByPod map[string]workloadRef) flowEndpoint {
	ep := flowEndpoint{namespace: ns, name: name}
	if ns == "" {
		ns = "default"
		ep.namespace = ns
	}

	// 直接匹配 Service
	if _, ok := serviceByKey[ns+"/"+name]; ok {
		ep.kind = "service"
		ep.nodeID = nodeID("service", ns, name)
		return ep
	}

	// Pod 名匹配 workload
	for podKey, wl := range workloadByPod {
		if !strings.HasPrefix(podKey, ns+"/") {
			continue
		}
		podName := strings.TrimPrefix(podKey, ns+"/")
		if podName == name || strings.HasPrefix(podName, name+"-") || strings.HasPrefix(name, podName) {
			ep.kind = wl.kind
			ep.name = wl.name
			ep.nodeID = nodeID(wl.kind, ns, wl.name)
			return ep
		}
	}

	// 短名匹配 service（去掉 pod hash 后缀）
	baseName := trimPodHash(name)
	if _, ok := serviceByKey[ns+"/"+baseName]; ok {
		ep.kind = "service"
		ep.name = baseName
		ep.nodeID = nodeID("service", ns, baseName)
		return ep
	}

	ep.kind = "unknown"
	ep.nodeID = nodeID("unknown", ns, name)
	return ep
}

func trimPodHash(name string) string {
	parts := strings.Split(name, "-")
	if len(parts) >= 3 {
		last := parts[len(parts)-1]
		if len(last) >= 5 && isPodSuffix(last) {
			return strings.Join(parts[:len(parts)-2], "-")
		}
	}
	return name
}

func isPodSuffix(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

func enrichEdgeMetrics(topology *TrafficTopology, flows []CallFlowStat) {
	if len(flows) == 0 {
		return
	}
	flowIndex := map[string]CallFlowStat{}
	for _, f := range flows {
		srcID := nodeID(f.SourceType, f.SourceNS, f.Source)
		if f.SourceType == "service" {
			srcID = nodeID("service", f.SourceNS, f.Source)
		} else if f.SourceType != "" && f.SourceType != "unknown" {
			srcID = nodeID(f.SourceType, f.SourceNS, f.Source)
		}
		dstID := nodeID("service", f.TargetNS, f.Target)
		if f.TargetType != "service" && f.TargetType != "" && f.TargetType != "unknown" {
			dstID = nodeID(f.TargetType, f.TargetNS, f.Target)
		}
		key := srcID + "->" + dstID
		existing := flowIndex[key]
		existing.FlowsPerSec += f.FlowsPerSec
		existing.Drops += f.Drops
		existing.BytesPerSec += f.BytesPerSec
		existing.Observed = true
		flowIndex[key] = existing
	}

	for i := range topology.Edges {
		e := &topology.Edges[i]
		key := e.Source + "->" + e.Target
		if m, ok := flowIndex[key]; ok {
			e.Metrics = &EdgeMetrics{
				FlowsPerSec: m.FlowsPerSec,
				BytesPerSec: m.BytesPerSec,
				Drops:       m.Drops,
				Observed:    true,
			}
			e.Inferred = false
		}
		// workload -> service 反向匹配
		for fk, m := range flowIndex {
			if strings.HasPrefix(fk, e.Source+"->") && strings.HasSuffix(fk, e.Target) {
				if e.Metrics == nil {
					e.Metrics = &EdgeMetrics{Observed: true}
				}
				e.Metrics.FlowsPerSec += m.FlowsPerSec
				e.Metrics.Drops += m.Drops
				e.Inferred = false
			}
		}
	}
}

func buildCallChains(topology *TrafficTopology) []CallChainPath {
	if topology == nil {
		return nil
	}

	adj := map[string][]TopologyEdge{}
	for _, e := range topology.Edges {
		if e.EdgeType == "calls" || e.EdgeType == "routes" || e.EdgeType == "selects" {
			adj[e.Source] = append(adj[e.Source], e)
		}
	}

	nodeByID := map[string]TopologyNode{}
	for _, n := range topology.Nodes {
		nodeByID[n.ID] = n
	}

	chains := make([]CallChainPath, 0)
	seen := map[string]struct{}{}

	addChain := func(hops []CallChainHop, entry string, observed bool, flows float64) {
		if len(hops) < 2 {
			return
		}
		key := ""
		for _, h := range hops {
			key += h.NodeID + ">"
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		chains = append(chains, CallChainPath{
			ID: fmt.Sprintf("chain-%d", len(chains)+1),
			Entry: entry, Hops: hops, Depth: len(hops),
			Observed: observed, FlowsPerSec: flows,
		})
	}

	hopFromNode := func(id string) CallChainHop {
		n, ok := nodeByID[id]
		if !ok {
			parts := strings.Split(id, "/")
			if len(parts) >= 3 {
				return CallChainHop{NodeID: id, Type: parts[0], Namespace: parts[1], Name: parts[2]}
			}
			return CallChainHop{NodeID: id}
		}
		return CallChainHop{NodeID: id, Type: n.Type, Name: n.Name, Namespace: n.Namespace}
	}

	// 从 Ingress 入口路径扩展
	for _, path := range topology.Paths {
		startHops := make([]CallChainHop, 0)
		entry := path.ServiceName
		if path.IngressName != "" {
			ingID := nodeID("ingress", path.Namespace, path.IngressName)
			startHops = append(startHops, hopFromNode(ingID))
			entry = path.IngressHost
			if entry == "" {
				entry = path.IngressName
			}
		}
		svcID := nodeID("service", path.Namespace, path.ServiceName)
		startHops = append(startHops, hopFromNode(svcID))
		if path.WorkloadName != "" {
			wlID := nodeID(path.WorkloadType, path.Namespace, path.WorkloadName)
			startHops = append(startHops, hopFromNode(wlID))
		}

		walkChain(adj, hopFromNode, startHops, entry, addChain, 0, 6)
	}

	// 从观测到的 call flow 构建
	for _, flow := range topology.CallFlows {
		if !flow.Observed {
			continue
		}
		srcID := nodeID(flow.SourceType, flow.SourceNS, flow.Source)
		if flow.SourceType == "service" {
			srcID = nodeID("service", flow.SourceNS, flow.Source)
		}
		dstID := nodeID("service", flow.TargetNS, flow.Target)
		hops := []CallChainHop{hopFromNode(srcID), hopFromNode(dstID)}
		addChain(hops, flow.Source, true, flow.FlowsPerSec)
	}

	sort.Slice(chains, func(i, j int) bool {
		if chains[i].Observed != chains[j].Observed {
			return chains[i].Observed
		}
		return chains[i].FlowsPerSec > chains[j].FlowsPerSec
	})
	if len(chains) > 30 {
		chains = chains[:30]
	}
	return chains
}

func walkChain(
	adj map[string][]TopologyEdge,
	hopFromNode func(string) CallChainHop,
	hops []CallChainHop,
	entry string,
	addChain func([]CallChainHop, string, bool, float64),
	depth, maxDepth int,
) {
	if depth >= maxDepth {
		addChain(hops, entry, false, 0)
		return
	}
	last := hops[len(hops)-1].NodeID
	edges := adj[last]
	if len(edges) == 0 {
		addChain(hops, entry, false, 0)
		return
	}
	extended := false
	for _, e := range edges {
		if e.Target == last {
			continue
		}
		already := false
		for _, h := range hops {
			if h.NodeID == e.Target {
				already = true
				break
			}
		}
		if already {
			continue
		}
		extended = true
		next := append(append([]CallChainHop{}, hops...), hopFromNode(e.Target))
		flows := float64(0)
		observed := false
		if e.Metrics != nil {
			flows = e.Metrics.FlowsPerSec
			observed = e.Metrics.Observed
		}
		if e.EdgeType == "calls" && depth+1 >= maxDepth-1 {
			addChain(next, entry, observed, flows)
			continue
		}
		walkChain(adj, hopFromNode, next, entry, addChain, depth+1, maxDepth)
	}
	if !extended {
		addChain(hops, entry, false, 0)
	}
}
