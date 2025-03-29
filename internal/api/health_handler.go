package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckHandler 处理健康检查请求
type HealthCheckHandler struct{}

// NewHealthCheckHandler 创建一个新的健康检查处理器
func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// CheckHealth 返回服务健康状态
// @Summary 健康检查接口
// @Description 返回服务的健康状态
// @Tags 系统
// @Produce json
// @Success 200 {object} Response
// @Router /health [get]
func (h *HealthCheckHandler) CheckHealth(c *gin.Context) {
	ResponseError(c, http.StatusOK, "服务运行正常")
}
