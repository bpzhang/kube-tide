package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckHandler health check handler
type HealthCheckHandler struct{}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// CheckHealth health check endpoint
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags System
// @Produce json
// @Success 200 {object} Response
// @Router /health [get]
func (h *HealthCheckHandler) CheckHealth(c *gin.Context) {
	ResponseError(c, http.StatusOK, "Service is running normally")
}
