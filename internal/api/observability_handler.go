package api

import (
	"context"
	"net/http"
	"strconv"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type ObservabilityHandler struct {
	service *k8s.ObservabilityService
}

func NewObservabilityHandler(service *k8s.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{service: service}
}

func (h *ObservabilityHandler) GetSummary(c *gin.Context) {
	clusterName := c.Param("cluster")
	summary, err := h.service.GetSummary(context.Background(), clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "observability.summaryFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"summary": summary})
}

func (h *ObservabilityHandler) ListAlerts(c *gin.Context) {
	clusterName := c.Param("cluster")
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	alerts, err := h.service.ListAlerts(context.Background(), clusterName, limit)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "observability.alertsFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"alerts": alerts})
}

func (h *ObservabilityHandler) GetPrometheusInfo(c *gin.Context) {
	clusterName := c.Param("cluster")
	info, err := h.service.GetPrometheusInfo(context.Background(), clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "prometheus.infoFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"prometheus": info})
}
