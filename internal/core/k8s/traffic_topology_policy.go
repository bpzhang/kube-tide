package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func addNetworkPolicyEdges(
	ctx context.Context,
	client *kubernetes.Clientset,
	nsList []string,
	podsByNS map[string][]corev1.Pod,
	workloadByPod map[string]workloadRef,
	addEdge func(TopologyEdge),
	nodeIndex map[string]struct{},
) {
	for _, ns := range nsList {
		list, err := client.NetworkingV1().NetworkPolicies(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		for i := range list.Items {
			np := &list.Items[i]
			targets := workloadsForSelector(ns, &np.Spec.PodSelector, podsByNS, workloadByPod)
			if len(targets) == 0 {
				continue
			}

			for _, rule := range np.Spec.Ingress {
				peers := workloadsFromPeers(ctx, client, ns, nsList, rule.From, podsByNS, workloadByPod)
				for _, from := range peers {
					fromID := nodeID(from.kind, ns, from.name)
					if _, ok := nodeIndex[fromID]; !ok {
						continue
					}
					for _, to := range targets {
						toID := nodeID(to.kind, ns, to.name)
						if fromID == toID {
							continue
						}
						port := formatPolicyPorts(rule.Ports)
						addEdge(TopologyEdge{
							Source:   fromID,
							Target:   toID,
							EdgeType: "policy_allow",
							Port:     port,
							Evidence: fmt.Sprintf("ingress:%s", np.Name),
						})
					}
				}
			}

			for _, rule := range np.Spec.Egress {
				peers := workloadsFromPeers(ctx, client, ns, nsList, rule.To, podsByNS, workloadByPod)
				for _, from := range targets {
					fromID := nodeID(from.kind, ns, from.name)
					for _, to := range peers {
						toID := nodeID(to.kind, to.namespace, to.name)
						if _, ok := nodeIndex[toID]; !ok {
							continue
						}
						if fromID == toID {
							continue
						}
						port := formatPolicyPorts(rule.Ports)
						addEdge(TopologyEdge{
							Source:   fromID,
							Target:   toID,
							EdgeType: "policy_allow",
							Port:     port,
							Evidence: fmt.Sprintf("egress:%s", np.Name),
						})
					}
				}
			}
		}
	}
}

type workloadWithNS struct {
	kind      string
	name      string
	namespace string
}

func workloadsForSelector(
	ns string,
	selector *metav1.LabelSelector,
	podsByNS map[string][]corev1.Pod,
	workloadByPod map[string]workloadRef,
) []workloadWithNS {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]workloadWithNS, 0)
	for i := range podsByNS[ns] {
		pod := &podsByNS[ns][i]
		if !sel.Matches(labels.Set(pod.Labels)) {
			continue
		}
		wl, ok := workloadByPod[ns+"/"+pod.Name]
		if !ok || wl.name == "" {
			continue
		}
		key := wl.kind + "/" + wl.name
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, workloadWithNS{kind: wl.kind, name: wl.name, namespace: ns})
	}
	return result
}

func workloadsFromPeers(
	ctx context.Context,
	client *kubernetes.Clientset,
	policyNS string,
	nsList []string,
	peers []networkingv1.NetworkPolicyPeer,
	podsByNS map[string][]corev1.Pod,
	workloadByPod map[string]workloadRef,
) []workloadWithNS {
	result := make([]workloadWithNS, 0)
	seen := map[string]struct{}{}
	add := func(wl workloadWithNS) {
		key := wl.namespace + "/" + wl.kind + "/" + wl.name
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, wl)
	}

	for _, peer := range peers {
		if peer.PodSelector != nil {
			for _, wl := range workloadsForSelector(policyNS, peer.PodSelector, podsByNS, workloadByPod) {
				add(wl)
			}
		}
		if peer.NamespaceSelector != nil {
			nsSel, err := metav1.LabelSelectorAsSelector(peer.NamespaceSelector)
			if err != nil {
				continue
			}
			for _, ns := range nsList {
				nsObj, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
				if err != nil {
					continue
				}
				if !nsSel.Matches(labels.Set(nsObj.Labels)) {
					continue
				}
				for _, wl := range workloadsForSelector(ns, &metav1.LabelSelector{}, podsByNS, workloadByPod) {
					add(wl)
				}
			}
		}
	}
	return result
}

func formatPolicyPorts(ports []networkingv1.NetworkPolicyPort) string {
	if len(ports) == 0 {
		return ""
	}
	if ports[0].Port != nil {
		return ports[0].Port.String()
	}
	return ""
}
