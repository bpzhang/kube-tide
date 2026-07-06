package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type NetworkPolicyHandler struct {
	service *k8s.NetworkPolicyService
}

func NewNetworkPolicyHandler(service *k8s.NetworkPolicyService) *NetworkPolicyHandler {
	return &NetworkPolicyHandler{service: service}
}

func (h *NetworkPolicyHandler) ListNetworkPolicies(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListNetworkPolicies(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "networkpolicy.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"networkpolicies": items})
}

func (h *NetworkPolicyHandler) GetNetworkPolicy(c *gin.Context) {
	item, err := h.service.GetNetworkPolicy(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("networkpolicy"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "networkpolicy.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"networkpolicy": item})
}

func (h *NetworkPolicyHandler) CreateNetworkPolicy(c *gin.Context) {
	var req k8s.CreateNetworkPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "networkpolicy.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateNetworkPolicy(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "networkpolicy.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"networkpolicy": item})
}

func (h *NetworkPolicyHandler) UpdateNetworkPolicy(c *gin.Context) {
	var req k8s.UpdateNetworkPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "networkpolicy.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateNetworkPolicy(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("networkpolicy"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "networkpolicy.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"networkpolicy": item})
}

func (h *NetworkPolicyHandler) DeleteNetworkPolicy(c *gin.Context) {
	if err := h.service.DeleteNetworkPolicy(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("networkpolicy")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "networkpolicy.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "NetworkPolicy deleted successfully"})
}
