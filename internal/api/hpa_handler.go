package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

type HPAHandler struct {
	service *k8s.HPAService
}

func NewHPAHandler(service *k8s.HPAService) *HPAHandler {
	return &HPAHandler{service: service}
}

func (h *HPAHandler) ListHPAs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListHPAs(context.Background(), clusterName, namespace)
	if err != nil {
		logger.Errorf("获取 HPA 列表失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "hpa.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"hpas": items})
}

func (h *HPAHandler) GetHPA(c *gin.Context) {
	item, err := h.service.GetHPA(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("hpa"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "hpa.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"hpa": item})
}

func (h *HPAHandler) CreateHPA(c *gin.Context) {
	var req k8s.CreateHPARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "hpa.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateHPA(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "hpa.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"hpa": item})
}

func (h *HPAHandler) UpdateHPA(c *gin.Context) {
	var req k8s.UpdateHPARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "hpa.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateHPA(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("hpa"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "hpa.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"hpa": item})
}

func (h *HPAHandler) DeleteHPA(c *gin.Context) {
	if err := h.service.DeleteHPA(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("hpa")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "hpa.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "HPA deleted successfully"})
}
