package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type LimitRangeHandler struct {
	service *k8s.LimitRangeService
}

func NewLimitRangeHandler(service *k8s.LimitRangeService) *LimitRangeHandler {
	return &LimitRangeHandler{service: service}
}

func (h *LimitRangeHandler) ListLimitRanges(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListLimitRanges(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "limitrange.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"limitranges": items})
}

func (h *LimitRangeHandler) GetLimitRange(c *gin.Context) {
	item, err := h.service.GetLimitRange(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("limitrange"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "limitrange.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"limitrange": item})
}

func (h *LimitRangeHandler) CreateLimitRange(c *gin.Context) {
	var req k8s.CreateLimitRangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "limitrange.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateLimitRange(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "limitrange.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"limitrange": item})
}

func (h *LimitRangeHandler) UpdateLimitRange(c *gin.Context) {
	var req k8s.UpdateLimitRangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "limitrange.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateLimitRange(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("limitrange"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "limitrange.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"limitrange": item})
}

func (h *LimitRangeHandler) DeleteLimitRange(c *gin.Context) {
	if err := h.service.DeleteLimitRange(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("limitrange")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "limitrange.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "LimitRange deleted successfully"})
}
