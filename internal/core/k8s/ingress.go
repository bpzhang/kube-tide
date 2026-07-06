package k8s

import (
	"context"
	"fmt"
	"strconv"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressManager Ingress 资源管理器。
type IngressManager struct {
	clientManager *ClientManager
}

// NewIngressManager 创建 Ingress 管理器。
func NewIngressManager(clientManager *ClientManager) *IngressManager {
	return &IngressManager{clientManager: clientManager}
}

// IngressBackendSpec Ingress 后端规格
type IngressBackendSpec struct {
	ServiceName string `json:"serviceName"`
	ServicePort string `json:"servicePort"`
}

// IngressPathSpec Ingress 路径规格
type IngressPathSpec struct {
	Path     string             `json:"path,omitempty"`
	PathType string             `json:"pathType,omitempty"`
	Backend  IngressBackendSpec `json:"backend"`
}

// IngressRuleSpec Ingress 规则规格
type IngressRuleSpec struct {
	Host  string            `json:"host,omitempty"`
	Paths []IngressPathSpec `json:"paths"`
}

// IngressTLSSpec Ingress TLS 规格
type IngressTLSSpec struct {
	Hosts      []string `json:"hosts,omitempty"`
	SecretName string   `json:"secretName,omitempty"`
}

// CreateIngressRequest 创建 Ingress 请求
type CreateIngressRequest struct {
	Name             string            `json:"name" binding:"required"`
	Namespace        string            `json:"namespace"`
	Labels           map[string]string `json:"labels,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	IngressClassName string            `json:"ingressClassName,omitempty"`
	Rules            []IngressRuleSpec `json:"rules" binding:"required"`
	TLS              []IngressTLSSpec  `json:"tls,omitempty"`
}

// UpdateIngressRequest 更新 Ingress 请求
type UpdateIngressRequest struct {
	Labels           map[string]string  `json:"labels,omitempty"`
	Annotations      map[string]string  `json:"annotations,omitempty"`
	IngressClassName *string            `json:"ingressClassName,omitempty"`
	Rules            []IngressRuleSpec  `json:"rules,omitempty"`
	TLS              []IngressTLSSpec   `json:"tls,omitempty"`
}

// GetIngressesByNamespace 获取指定命名空间中的 Ingress 列表。
func (m *IngressManager) GetIngressesByNamespace(ctx context.Context, clusterName, namespace string) ([]networkingv1.Ingress, error) {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间 %s 的 Ingress 列表失败: %w", namespace, err)
	}

	return ingressList.Items, nil
}

// GetIngress 获取单个 Ingress
func (m *IngressManager) GetIngress(ctx context.Context, clusterName, namespace, name string) (*networkingv1.Ingress, error) {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ing, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 失败: %w", err)
	}
	return ing, nil
}

// CreateIngress 创建 Ingress
func (m *IngressManager) CreateIngress(ctx context.Context, clusterName, namespace string, req CreateIngressRequest) (*networkingv1.Ingress, error) {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: buildIngressSpec(req.IngressClassName, req.Rules, req.TLS),
	}
	return client.NetworkingV1().Ingresses(namespace).Create(ctx, ing, metav1.CreateOptions{})
}

// UpdateIngress 更新 Ingress
func (m *IngressManager) UpdateIngress(ctx context.Context, clusterName, namespace, name string, req UpdateIngressRequest) (*networkingv1.Ingress, error) {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	ing, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 失败: %w", err)
	}
	if req.Labels != nil {
		ing.Labels = req.Labels
	}
	if req.Annotations != nil {
		ing.Annotations = req.Annotations
	}
	className := ""
	if ing.Spec.IngressClassName != nil {
		className = *ing.Spec.IngressClassName
	}
	if req.IngressClassName != nil {
		className = *req.IngressClassName
	}
	if req.Rules != nil || req.TLS != nil || req.IngressClassName != nil {
		rules := req.Rules
		tls := req.TLS
		if rules == nil {
			rules = convertIngressRulesFromK8s(ing.Spec.Rules)
		}
		if tls == nil {
			tls = convertIngressTLSFromK8s(ing.Spec.TLS)
		}
		ing.Spec = buildIngressSpec(className, rules, tls)
	}
	return client.NetworkingV1().Ingresses(namespace).Update(ctx, ing, metav1.UpdateOptions{})
}

// DeleteIngress 删除 Ingress
func (m *IngressManager) DeleteIngress(ctx context.Context, clusterName, namespace, name string) error {
	client, err := m.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func buildIngressSpec(className string, rules []IngressRuleSpec, tls []IngressTLSSpec) networkingv1.IngressSpec {
	spec := networkingv1.IngressSpec{
		Rules: buildIngressRules(rules),
		TLS:   buildIngressTLS(tls),
	}
	if className != "" {
		spec.IngressClassName = &className
	}
	return spec
}

func buildIngressRules(rules []IngressRuleSpec) []networkingv1.IngressRule {
	result := make([]networkingv1.IngressRule, 0, len(rules))
	for _, rule := range rules {
		paths := make([]networkingv1.HTTPIngressPath, 0, len(rule.Paths))
		for _, p := range rule.Paths {
			pathType := networkingv1.PathTypePrefix
			if p.PathType != "" {
				pt := networkingv1.PathType(p.PathType)
				pathType = pt
			}
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     p.Path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: p.Backend.ServiceName,
						Port: parseServicePort(p.Backend.ServicePort),
					},
				},
			})
		}
		result = append(result, networkingv1.IngressRule{Host: rule.Host, IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{Paths: paths}}})
	}
	return result
}

func buildIngressTLS(tls []IngressTLSSpec) []networkingv1.IngressTLS {
	result := make([]networkingv1.IngressTLS, 0, len(tls))
	for _, t := range tls {
		result = append(result, networkingv1.IngressTLS{Hosts: t.Hosts, SecretName: t.SecretName})
	}
	return result
}

func parseServicePort(portStr string) networkingv1.ServiceBackendPort {
	if portStr == "" {
		return networkingv1.ServiceBackendPort{}
	}
	if num, err := strconv.Atoi(portStr); err == nil {
		p := int32(num)
		return networkingv1.ServiceBackendPort{Number: p}
	}
	return networkingv1.ServiceBackendPort{Name: portStr}
}

func convertIngressRulesFromK8s(rules []networkingv1.IngressRule) []IngressRuleSpec {
	result := make([]IngressRuleSpec, 0, len(rules))
	for _, rule := range rules {
		spec := IngressRuleSpec{Host: rule.Host}
		if rule.HTTP != nil {
			for _, p := range rule.HTTP.Paths {
				pathType := ""
				if p.PathType != nil {
					pathType = string(*p.PathType)
				}
				backend := IngressBackendSpec{}
				if p.Backend.Service != nil {
					backend.ServiceName = p.Backend.Service.Name
					if p.Backend.Service.Port.Name != "" {
						backend.ServicePort = p.Backend.Service.Port.Name
					} else if p.Backend.Service.Port.Number != 0 {
						backend.ServicePort = strconv.Itoa(int(p.Backend.Service.Port.Number))
					}
				}
				spec.Paths = append(spec.Paths, IngressPathSpec{Path: p.Path, PathType: pathType, Backend: backend})
			}
		}
		result = append(result, spec)
	}
	return result
}

func convertIngressTLSFromK8s(tls []networkingv1.IngressTLS) []IngressTLSSpec {
	result := make([]IngressTLSSpec, 0, len(tls))
	for _, t := range tls {
		result = append(result, IngressTLSSpec{Hosts: t.Hosts, SecretName: t.SecretName})
	}
	return result
}
