package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type DaemonSetHandler struct {
	service *k8s.DaemonSetService
}

func NewDaemonSetHandler(service *k8s.DaemonSetService) *DaemonSetHandler {
	return &DaemonSetHandler{service: service}
}

func (h *DaemonSetHandler) ListDaemonSets(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListDaemonSets(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"daemonsets": items})
}

func (h *DaemonSetHandler) GetDaemonSet(c *gin.Context) {
	item, err := h.service.GetDaemonSet(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("daemonset"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"daemonset": item})
}

func (h *DaemonSetHandler) CreateDaemonSet(c *gin.Context) {
	var req k8s.CreateDaemonSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "daemonset.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateDaemonSet(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"daemonset": item})
}

func (h *DaemonSetHandler) UpdateDaemonSet(c *gin.Context) {
	var req k8s.UpdateDaemonSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "daemonset.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateDaemonSet(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("daemonset"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"daemonset": item})
}

func (h *DaemonSetHandler) DeleteDaemonSet(c *gin.Context) {
	if err := h.service.DeleteDaemonSet(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("daemonset")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "DaemonSet deleted successfully"})
}

func (h *DaemonSetHandler) GetDaemonSetPods(c *gin.Context) {
	pods, err := h.service.GetDaemonSetPods(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("daemonset"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "daemonset.getPodsFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pods": pods})
}
