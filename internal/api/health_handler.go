package api

import (
	"context"
	"net/http"
	"time"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

const healthProbeTimeout = 5 * time.Second

// HealthCheckHandler Health check handler
type HealthCheckHandler struct {
	clientManager *k8s.ClientManager
}

// NewHealthCheckHandler Create health check handler
func NewHealthCheckHandler(clientManager *k8s.ClientManager) *HealthCheckHandler {
	return &HealthCheckHandler{clientManager: clientManager}
}

// CheckHealth Check system liveness
func (h *HealthCheckHandler) CheckHealth(c *gin.Context) {
	ResponseSuccess(c, gin.H{
		"status": "ok",
	})
}

// CheckReadiness Check system readiness (probe registered cluster connectivity)
func (h *HealthCheckHandler) CheckReadiness(c *gin.Context) {
	clusters := h.clientManager.ListClusters()
	if len(clusters) == 0 {
		ResponseSuccess(c, gin.H{
			"status":  "ready",
			"message": "no clusters registered",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), healthProbeTimeout)
	defer cancel()

	reachable := make([]string, 0, len(clusters))
	unreachable := make([]string, 0)

	for _, name := range clusters {
		client, err := h.clientManager.GetClient(name)
		if err != nil {
			unreachable = append(unreachable, name)
			continue
		}
		if _, err = client.ServerVersion(); err != nil {
			unreachable = append(unreachable, name)
			continue
		}
		reachable = append(reachable, name)
	}

	if len(reachable) == 0 {
		c.JSON(http.StatusServiceUnavailable, Response{
			Code:    http.StatusServiceUnavailable,
			Message: "no registered cluster is reachable",
			Data: gin.H{
				"status":      "not_ready",
				"reachable":   reachable,
				"unreachable": unreachable,
			},
		})
		return
	}

	_ = ctx
	ResponseSuccess(c, gin.H{
		"status":      "ready",
		"reachable":   reachable,
		"unreachable": unreachable,
	})
}
