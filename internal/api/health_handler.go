package api

import (
	"github.com/gin-gonic/gin"
)

// HealthCheckHandler Health check handler
type HealthCheckHandler struct{}

// NewHealthCheckHandler Create health check handler
func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// CheckHealth Check system health
func (h *HealthCheckHandler) CheckHealth(c *gin.Context) {
	ResponseSuccess(c, gin.H{
		"status": "system.healthCheck",
	})
}
