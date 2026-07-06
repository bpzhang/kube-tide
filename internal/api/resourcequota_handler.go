package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type ResourceQuotaHandler struct {
	service *k8s.ResourceQuotaService
}

func NewResourceQuotaHandler(service *k8s.ResourceQuotaService) *ResourceQuotaHandler {
	return &ResourceQuotaHandler{service: service}
}

func (h *ResourceQuotaHandler) ListResourceQuotas(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListResourceQuotas(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "resourcequota.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"resourcequotas": items})
}

func (h *ResourceQuotaHandler) GetResourceQuota(c *gin.Context) {
	item, err := h.service.GetResourceQuota(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("resourcequota"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "resourcequota.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"resourcequota": item})
}

func (h *ResourceQuotaHandler) CreateResourceQuota(c *gin.Context) {
	var req k8s.CreateResourceQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "resourcequota.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateResourceQuota(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "resourcequota.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"resourcequota": item})
}

func (h *ResourceQuotaHandler) UpdateResourceQuota(c *gin.Context) {
	var req k8s.UpdateResourceQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "resourcequota.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateResourceQuota(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("resourcequota"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "resourcequota.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"resourcequota": item})
}

func (h *ResourceQuotaHandler) DeleteResourceQuota(c *gin.Context) {
	if err := h.service.DeleteResourceQuota(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("resourcequota")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "resourcequota.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "ResourceQuota deleted successfully"})
}
