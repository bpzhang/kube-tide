package k8s

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// TopologyNode 拓扑节点
type TopologyNode struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	Extra     map[string]any    `json:"extra,omitempty"`
}

// TopologyEdge 拓扑边
type TopologyEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	EdgeType string `json:"edgeType"`
	Port     string `json:"port,omitempty"`
	Inferred bool   `json:"inferred,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

// TrafficPath 外部到后端的流量路径
type TrafficPath struct {
	Namespace    string `json:"namespace"`
	IngressName  string `json:"ingressName,omitempty"`
	IngressHost  string `json:"ingressHost,omitempty"`
	Path         string `json:"path,omitempty"`
	ServiceName  string `json:"serviceName"`
	WorkloadType string `json:"workloadType,omitempty"`
	WorkloadName string `json:"workloadName,omitempty"`
	PodCount     int    `json:"podCount"`
}

// TrafficTopology 服务流量拓扑
type TrafficTopology struct {
	Nodes   []TopologyNode        `json:"nodes"`
	Edges   []TopologyEdge        `json:"edges"`
	Paths   []TrafficPath         `json:"paths"`
	Network *ClusterNetworkInfo   `json:"network,omitempty"`
	Hubble  *HubbleMetricsSummary `json:"hubble,omitempty"`
}

// TrafficTopologyService 流量拓扑服务
type TrafficTopologyService struct {
	clientManager *ClientManager
	prometheus    *PrometheusService
}

// NewTrafficTopologyService 创建流量拓扑服务
func NewTrafficTopologyService(clientManager *ClientManager, prometheus *PrometheusService) *TrafficTopologyService {
	return &TrafficTopologyService{clientManager: clientManager, prometheus: prometheus}
}

var (
	svcDNSPattern   = regexp.MustCompile(`(?i)(?:https?://)?([a-z0-9](?:[a-z0-9-]*[a-z0-9])?)(?:\.([a-z0-9][a-z0-9-]*))?\.svc(?:\.cluster\.local)?(?:[:/]|$)`)
	svcShortPattern = regexp.MustCompile(`(?i)(?:https?://)?([a-z0-9](?:[a-z0-9-]*[a-z0-9])?)(?:\.([a-z0-9][a-z0-9-]*))?\.svc(?:[:/]|$)`)
)

func nodeID(nodeType, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", nodeType, namespace, name)
}

// GetTrafficTopology 获取命名空间或全集群流量拓扑
func (s *TrafficTopologyService) GetTrafficTopology(ctx context.Context, clusterName, namespace string) (*TrafficTopology, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	nsList, err := listTargetNamespaces(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	topology := &TrafficTopology{
		Nodes: make([]TopologyNode, 0),
		Edges: make([]TopologyEdge, 0),
		Paths: make([]TrafficPath, 0),
	}
	nodeIndex := map[string]struct{}{}
	edgeKeys := map[string]struct{}{}

	addNode := func(n TopologyNode) {
		if _, ok := nodeIndex[n.ID]; ok {
			return
		}
		nodeIndex[n.ID] = struct{}{}
		topology.Nodes = append(topology.Nodes, n)
	}
	addEdge := func(e TopologyEdge) {
		key := e.Source + "->" + e.Target + ":" + e.EdgeType + ":" + e.Port + ":" + e.Evidence
		if _, ok := edgeKeys[key]; ok {
			return
		}
		edgeKeys[key] = struct{}{}
		topology.Edges = append(topology.Edges, e)
	}

	serviceByKey := map[string]corev1.Service{}
	workloadByPod := map[string]workloadRef{}
	podsByNS := map[string][]corev1.Pod{}

	for _, ns := range nsList {
		services, err := client.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Service 列表失败: %w", err)
		}
		for _, svc := range services.Items {
			serviceByKey[ns+"/"+svc.Name] = svc
			ports := make([]string, 0, len(svc.Spec.Ports))
			for _, p := range svc.Spec.Ports {
				if p.Port > 0 {
					ports = append(ports, fmt.Sprintf("%d", p.Port))
				}
			}
			addNode(TopologyNode{
				ID: nodeID("service", ns, svc.Name), Type: "service", Name: svc.Name, Namespace: ns,
				Extra: map[string]any{"type": string(svc.Spec.Type), "clusterIP": svc.Spec.ClusterIP, "ports": ports},
			})
		}

		ingresses, err := client.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Ingress 列表失败: %w", err)
		}
		for _, ing := range ingresses.Items {
			addNode(TopologyNode{
				ID: nodeID("ingress", ns, ing.Name), Type: "ingress", Name: ing.Name, Namespace: ns,
				Extra: map[string]any{"ingressClass": derefString(ing.Spec.IngressClassName)},
			})
			for _, rule := range ing.Spec.Rules {
				host := rule.Host
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service == nil {
						continue
					}
					svcName := path.Backend.Service.Name
					svcPort := formatServiceBackendPort(path.Backend.Service.Port)
					addEdge(TopologyEdge{
						Source:   nodeID("ingress", ns, ing.Name),
						Target:   nodeID("service", ns, svcName),
						EdgeType: "routes", Port: svcPort,
					})
					topology.Paths = append(topology.Paths, TrafficPath{
						Namespace: ns, IngressName: ing.Name, IngressHost: host,
						Path: path.Path, ServiceName: svcName,
					})
				}
			}
		}

		podList, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
		}
		podsByNS[ns] = podList.Items
		for i := range podList.Items {
			pod := &podList.Items[i]
			wl := resolvePodWorkload(ctx, client, pod)
			if wl.name != "" {
				workloadByPod[ns+"/"+pod.Name] = wl
				addNode(TopologyNode{
					ID: nodeID(wl.kind, ns, wl.name), Type: wl.kind, Name: wl.name, Namespace: ns,
				})
			}
		}
	}

	for _, ns := range nsList {
		for key, svc := range serviceByKey {
			if !strings.HasPrefix(key, ns+"/") {
				continue
			}
			if len(svc.Spec.Selector) == 0 {
				continue
			}
			selector := labels.Set(svc.Spec.Selector).AsSelector()
			matchedPods := 0
			var matchedWL workloadRef
			for i := range podsByNS[ns] {
				pod := &podsByNS[ns][i]
				if !selector.Matches(labels.Set(pod.Labels)) {
					continue
				}
				matchedPods++
				if wl, ok := workloadByPod[ns+"/"+pod.Name]; ok && wl.name != "" {
					matchedWL = wl
				}
			}
			if matchedWL.name == "" {
				continue
			}
			addEdge(TopologyEdge{
				Source:   nodeID("service", ns, svc.Name),
				Target:   nodeID(matchedWL.kind, ns, matchedWL.name),
				EdgeType: "selects",
			})
			enriched := false
			for idx := range topology.Paths {
				if topology.Paths[idx].Namespace == ns && topology.Paths[idx].ServiceName == svc.Name {
					topology.Paths[idx].WorkloadType = matchedWL.kind
					topology.Paths[idx].WorkloadName = matchedWL.name
					topology.Paths[idx].PodCount = matchedPods
					enriched = true
				}
			}
			if !enriched {
				topology.Paths = append(topology.Paths, TrafficPath{
					Namespace: ns, ServiceName: svc.Name,
					WorkloadType: matchedWL.kind, WorkloadName: matchedWL.name, PodCount: matchedPods,
				})
			}
		}
	}

	for _, ns := range nsList {
		for i := range podsByNS[ns] {
			pod := &podsByNS[ns][i]
			wl, ok := workloadByPod[ns+"/"+pod.Name]
			if !ok || wl.name == "" {
				continue
			}
			sourceID := nodeID(wl.kind, ns, wl.name)
			for _, ref := range extractServiceReferences(pod, ns) {
				targetNS := ref.namespace
				if targetNS == "" {
					targetNS = ns
				}
				if _, exists := serviceByKey[targetNS+"/"+ref.name]; !exists {
					continue
				}
				addEdge(TopologyEdge{
					Source: sourceID, Target: nodeID("service", targetNS, ref.name),
					EdgeType: "calls", Inferred: true, Evidence: ref.evidence,
				})
			}
		}
	}

	addNetworkPolicyEdges(ctx, client, nsList, podsByNS, workloadByPod, addEdge, nodeIndex)
	topology.Network = detectClusterNetwork(ctx, client)
	topology.Hubble = fetchHubbleMetrics(ctx, client, s.prometheus, clusterName, namespace, topology.Network)
	setNetworkMessage(topology.Network)

	sortTopology(topology)
	return topology, nil
}

type workloadRef struct {
	kind string
	name string
}

type serviceRef struct {
	name      string
	namespace string
	evidence  string
}

func listTargetNamespaces(ctx context.Context, client *kubernetes.Clientset, namespace string) ([]string, error) {
	if namespace != "" && namespace != "all" {
		return []string{namespace}, nil
	}
	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	return names, nil
}

func resolvePodWorkload(ctx context.Context, client *kubernetes.Clientset, pod *corev1.Pod) workloadRef {
	for _, owner := range pod.OwnerReferences {
		switch owner.Kind {
		case "ReplicaSet":
			rs, err := client.AppsV1().ReplicaSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			for _, rsOwner := range rs.OwnerReferences {
				if rsOwner.Kind == "Deployment" && rsOwner.Controller != nil && *rsOwner.Controller {
					return workloadRef{kind: "deployment", name: rsOwner.Name}
				}
			}
			return workloadRef{kind: "replicaset", name: owner.Name}
		case "StatefulSet", "DaemonSet", "Job":
			if owner.Controller != nil && *owner.Controller {
				return workloadRef{kind: strings.ToLower(owner.Kind), name: owner.Name}
			}
		}
	}
	return workloadRef{}
}

func extractServiceReferences(pod *corev1.Pod, defaultNS string) []serviceRef {
	refs := make([]serviceRef, 0)
	seen := map[string]struct{}{}
	add := func(name, ns, evidence string) {
		if name == "" {
			return
		}
		if ns == "" {
			ns = defaultNS
		}
		key := ns + "/" + name
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		refs = append(refs, serviceRef{name: strings.ToLower(name), namespace: ns, evidence: evidence})
	}

	for _, c := range pod.Spec.Containers {
		for _, e := range c.Env {
			if svc, ok := serviceNameFromK8sEnvKey(e.Name); ok {
				add(svc, defaultNS, fmt.Sprintf("env:%s", e.Name))
			}
			if e.Value != "" {
				parseServiceFromText(e.Value, func(name, ns string) {
					add(name, ns, fmt.Sprintf("env:%s", e.Name))
				})
			}
		}
		for _, arg := range c.Args {
			parseServiceFromText(arg, func(name, ns string) {
				add(name, ns, "arg")
			})
		}
	}
	return refs
}

func parseServiceFromText(text string, add func(name, ns string)) {
	for _, re := range []*regexp.Regexp{svcDNSPattern, svcShortPattern} {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			ns := ""
			if len(m) > 2 {
				ns = m[2]
			}
			add(m[1], ns)
		}
	}
}

func serviceNameFromK8sEnvKey(key string) (string, bool) {
	if !strings.HasSuffix(key, "_SERVICE_HOST") {
		return "", false
	}
	raw := strings.TrimSuffix(key, "_SERVICE_HOST")
	if raw == "" {
		return "", false
	}
	return strings.ToLower(strings.ReplaceAll(raw, "_", "-")), true
}

func formatServiceBackendPort(port networkingv1.ServiceBackendPort) string {
	if port.Name != "" {
		return port.Name
	}
	if port.Number > 0 {
		return fmt.Sprintf("%d", port.Number)
	}
	return ""
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func sortTopology(topology *TrafficTopology) {
	sort.Slice(topology.Nodes, func(i, j int) bool {
		if topology.Nodes[i].Type != topology.Nodes[j].Type {
			return topology.Nodes[i].Type < topology.Nodes[j].Type
		}
		if topology.Nodes[i].Namespace != topology.Nodes[j].Namespace {
			return topology.Nodes[i].Namespace < topology.Nodes[j].Namespace
		}
		return topology.Nodes[i].Name < topology.Nodes[j].Name
	})
	sort.Slice(topology.Edges, func(i, j int) bool {
		return topology.Edges[i].Source+topology.Edges[i].Target < topology.Edges[j].Source+topology.Edges[j].Target
	})
	sort.Slice(topology.Paths, func(i, j int) bool {
		return topology.Paths[i].Namespace+topology.Paths[i].ServiceName < topology.Paths[j].Namespace+topology.Paths[j].ServiceName
	})
}
