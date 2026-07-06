package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NetworkPolicyService NetworkPolicy 管理服务
type NetworkPolicyService struct {
	clientManager *ClientManager
}

// NewNetworkPolicyService 创建 NetworkPolicy 服务
func NewNetworkPolicyService(clientManager *ClientManager) *NetworkPolicyService {
	return &NetworkPolicyService{clientManager: clientManager}
}

// NetworkPolicyInfo NetworkPolicy 摘要
type NetworkPolicyInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	PolicyTypes    []string          `json:"policyTypes,omitempty"`
	PodSelector    map[string]string `json:"podSelector,omitempty"`
	IngressRuleCount int             `json:"ingressRuleCount"`
	EgressRuleCount  int             `json:"egressRuleCount"`
	CreationTime   time.Time         `json:"creationTime"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// NetworkPolicySpecRequest NetworkPolicy 规格请求
type NetworkPolicySpecRequest struct {
	PodSelector       map[string]string              `json:"podSelector"`
	PolicyTypes       []string                       `json:"policyTypes,omitempty"`
	Ingress           []NetworkPolicyIngressRule     `json:"ingress,omitempty"`
	Egress            []NetworkPolicyEgressRule      `json:"egress,omitempty"`
}

// NetworkPolicyIngressRule 入站规则
type NetworkPolicyIngressRule struct {
	From  []NetworkPolicyPeer `json:"from,omitempty"`
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
}

// NetworkPolicyEgressRule 出站规则
type NetworkPolicyEgressRule struct {
	To    []NetworkPolicyPeer `json:"to,omitempty"`
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
}

// NetworkPolicyPeer 网络策略对等体
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"podSelector,omitempty"`
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
	IPBlock           *IPBlock          `json:"ipBlock,omitempty"`
}

// IPBlock IP 块
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// NetworkPolicyPort 端口
type NetworkPolicyPort struct {
	Protocol string `json:"protocol,omitempty"`
	Port     *int32 `json:"port,omitempty"`
}

// CreateNetworkPolicyRequest 创建 NetworkPolicy 请求
type CreateNetworkPolicyRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Namespace   string                   `json:"namespace"`
	Labels      map[string]string        `json:"labels,omitempty"`
	Annotations map[string]string        `json:"annotations,omitempty"`
	Spec        NetworkPolicySpecRequest `json:"spec" binding:"required"`
}

// UpdateNetworkPolicyRequest 更新 NetworkPolicy 请求
type UpdateNetworkPolicyRequest struct {
	Labels      map[string]string         `json:"labels,omitempty"`
	Annotations map[string]string         `json:"annotations,omitempty"`
	Spec        *NetworkPolicySpecRequest `json:"spec,omitempty"`
}

// ListNetworkPolicies 获取 NetworkPolicy 列表
func (s *NetworkPolicyService) ListNetworkPolicies(ctx context.Context, clusterName, namespace string) ([]NetworkPolicyInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *networkingv1.NetworkPolicyList
	if namespace == "all" || namespace == "" {
		list, err = client.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 NetworkPolicy 列表失败: %w", err)
	}
	result := make([]NetworkPolicyInfo, 0, len(list.Items))
	for _, np := range list.Items {
		result = append(result, convertNetworkPolicyInfo(&np))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetNetworkPolicy 获取 NetworkPolicy 详情
func (s *NetworkPolicyService) GetNetworkPolicy(ctx context.Context, clusterName, namespace, name string) (*networkingv1.NetworkPolicy, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateNetworkPolicy 创建 NetworkPolicy
func (s *NetworkPolicyService) CreateNetworkPolicy(ctx context.Context, clusterName, namespace string, req CreateNetworkPolicyRequest) (*NetworkPolicyInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: buildNetworkPolicySpec(req.Spec),
	}
	created, err := client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, np, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 NetworkPolicy 失败: %w", err)
	}
	info := convertNetworkPolicyInfo(created)
	return &info, nil
}

// UpdateNetworkPolicy 更新 NetworkPolicy
func (s *NetworkPolicyService) UpdateNetworkPolicy(ctx context.Context, clusterName, namespace, name string, req UpdateNetworkPolicyRequest) (*NetworkPolicyInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	np, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 NetworkPolicy 失败: %w", err)
	}
	if req.Labels != nil {
		np.Labels = req.Labels
	}
	if req.Annotations != nil {
		np.Annotations = req.Annotations
	}
	if req.Spec != nil {
		np.Spec = buildNetworkPolicySpec(*req.Spec)
	}
	updated, err := client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, np, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 NetworkPolicy 失败: %w", err)
	}
	info := convertNetworkPolicyInfo(updated)
	return &info, nil
}

// DeleteNetworkPolicy 删除 NetworkPolicy
func (s *NetworkPolicyService) DeleteNetworkPolicy(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertNetworkPolicyInfo(np *networkingv1.NetworkPolicy) NetworkPolicyInfo {
	types := make([]string, 0, len(np.Spec.PolicyTypes))
	for _, t := range np.Spec.PolicyTypes {
		types = append(types, string(t))
	}
	return NetworkPolicyInfo{
		Name:             np.Name,
		Namespace:        np.Namespace,
		PolicyTypes:      types,
		PodSelector:      np.Spec.PodSelector.MatchLabels,
		IngressRuleCount: len(np.Spec.Ingress),
		EgressRuleCount:  len(np.Spec.Egress),
		CreationTime:     np.CreationTimestamp.Time,
		Labels:           np.Labels,
	}
}

func buildNetworkPolicySpec(req NetworkPolicySpecRequest) networkingv1.NetworkPolicySpec {
	spec := networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{MatchLabels: req.PodSelector},
	}
	for _, t := range req.PolicyTypes {
		spec.PolicyTypes = append(spec.PolicyTypes, networkingv1.PolicyType(t))
	}
	for _, ing := range req.Ingress {
		rule := networkingv1.NetworkPolicyIngressRule{}
		for _, from := range ing.From {
			rule.From = append(rule.From, buildNetworkPolicyPeer(from))
		}
		for _, p := range ing.Ports {
			rule.Ports = append(rule.Ports, buildNetworkPolicyPort(p))
		}
		spec.Ingress = append(spec.Ingress, rule)
	}
	for _, eg := range req.Egress {
		rule := networkingv1.NetworkPolicyEgressRule{}
		for _, to := range eg.To {
			rule.To = append(rule.To, buildNetworkPolicyPeer(to))
		}
		for _, p := range eg.Ports {
			rule.Ports = append(rule.Ports, buildNetworkPolicyPort(p))
		}
		spec.Egress = append(spec.Egress, rule)
	}
	return spec
}

func buildNetworkPolicyPeer(peer NetworkPolicyPeer) networkingv1.NetworkPolicyPeer {
	p := networkingv1.NetworkPolicyPeer{}
	if peer.PodSelector != nil {
		p.PodSelector = &metav1.LabelSelector{MatchLabels: peer.PodSelector}
	}
	if peer.NamespaceSelector != nil {
		p.NamespaceSelector = &metav1.LabelSelector{MatchLabels: peer.NamespaceSelector}
	}
	if peer.IPBlock != nil {
		p.IPBlock = &networkingv1.IPBlock{CIDR: peer.IPBlock.CIDR, Except: peer.IPBlock.Except}
	}
	return p
}

func buildNetworkPolicyPort(p NetworkPolicyPort) networkingv1.NetworkPolicyPort {
	port := networkingv1.NetworkPolicyPort{}
	if p.Protocol != "" {
		proto := corev1.Protocol(p.Protocol)
		port.Protocol = &proto
	}
	if p.Port != nil {
		pv := intstr.FromInt(int(*p.Port))
		port.Port = &pv
	}
	return port
}
