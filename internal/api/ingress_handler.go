package api

import (
	"context"
	"fmt"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
	networkingv1 "k8s.io/api/networking/v1"
)

// IngressHandler Ingress 管理处理器。
type IngressHandler struct {
	manager *k8s.IngressManager
}

// NewIngressHandler 创建 Ingress 管理处理器。
func NewIngressHandler(manager *k8s.IngressManager) *IngressHandler {
	return &IngressHandler{manager: manager}
}

type ingressBackendResponse struct {
	ServiceName string `json:"serviceName,omitempty"`
	ServicePort string `json:"servicePort,omitempty"`
}

type ingressPathResponse struct {
	Path     string                 `json:"path,omitempty"`
	PathType string                 `json:"pathType,omitempty"`
	Backend  ingressBackendResponse `json:"backend"`
}

type ingressRuleResponse struct {
	Host  string                `json:"host,omitempty"`
	Paths []ingressPathResponse `json:"paths"`
}

type ingressTLSResponse struct {
	Hosts      []string `json:"hosts,omitempty"`
	SecretName string   `json:"secretName,omitempty"`
}

type ingressResponse struct {
	Name             string                `json:"name"`
	Namespace        string                `json:"namespace"`
	IngressClassName string                `json:"ingressClassName,omitempty"`
	Rules            []ingressRuleResponse `json:"rules"`
	TLS              []ingressTLSResponse  `json:"tls"`
}

// ListIngressesByNamespace 获取指定命名空间中的 Ingress 列表。
func (h *IngressHandler) ListIngressesByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or namespace cannot be empty")
		return
	}

	ingresses, err := h.manager.GetIngressesByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]ingressResponse, 0, len(ingresses))
	for _, ing := range ingresses {
		result = append(result, convertIngressToResponse(ing))
	}

	ResponseSuccess(c, gin.H{"ingresses": result})
}

func convertIngressToResponse(ing networkingv1.Ingress) ingressResponse {
	rules := make([]ingressRuleResponse, 0, len(ing.Spec.Rules))
	for _, rule := range ing.Spec.Rules {
		paths := make([]ingressPathResponse, 0)
		if rule.HTTP != nil {
			paths = make([]ingressPathResponse, 0, len(rule.HTTP.Paths))
			for _, path := range rule.HTTP.Paths {
				servicePort := ""
				serviceName := ""
				if path.Backend.Service != nil {
					serviceName = path.Backend.Service.Name
					if path.Backend.Service.Port.Name != "" {
						servicePort = path.Backend.Service.Port.Name
					} else if path.Backend.Service.Port.Number != 0 {
						servicePort = fmt.Sprintf("%d", path.Backend.Service.Port.Number)
					}
				}

				pathType := ""
				if path.PathType != nil {
					pathType = string(*path.PathType)
				}

				paths = append(paths, ingressPathResponse{
					Path:     path.Path,
					PathType: pathType,
					Backend: ingressBackendResponse{
						ServiceName: serviceName,
						ServicePort: servicePort,
					},
				})
			}
		}

		rules = append(rules, ingressRuleResponse{Host: rule.Host, Paths: paths})
	}

	tls := make([]ingressTLSResponse, 0, len(ing.Spec.TLS))
	for _, item := range ing.Spec.TLS {
		tls = append(tls, ingressTLSResponse{Hosts: item.Hosts, SecretName: item.SecretName})
	}

	ingressClassName := ""
	if ing.Spec.IngressClassName != nil {
		ingressClassName = *ing.Spec.IngressClassName
	}

	return ingressResponse{
		Name:             ing.Name,
		Namespace:        ing.Namespace,
		IngressClassName: ingressClassName,
		Rules:            rules,
		TLS:              tls,
	}
}
