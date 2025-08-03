package api

import (
	"context"
	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AutoScalerHandler 自动扩缩容处理器
type AutoScalerHandler struct {
	service *k8s.AutoScalerService
}

// NewAutoScalerHandler 创建自动扩缩容处理器
func NewAutoScalerHandler(service *k8s.AutoScalerService) *AutoScalerHandler {
	return &AutoScalerHandler{
		service: service,
	}
}

// GetAutoScalerConfig 获取自动扩缩容配置
func (h *AutoScalerHandler) GetAutoScalerConfig(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "autoscaler.clusterNameEmpty")
		return
	}

	config, err := h.service.GetAutoScalerConfig(context.Background(), clusterName)
	if err != nil {
		logger.Errorf("Failed to get autoscaler config: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, "autoscaler.getConfigFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"config": config,
	})
}

// UpdateAutoScalerConfig 更新自动扩缩容配置
func (h *AutoScalerHandler) UpdateAutoScalerConfig(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "autoscaler.clusterNameEmpty")
		return
	}

	var config k8s.AutoScalerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		ResponseError(c, http.StatusBadRequest, "api.invalidParameters", err.Error())
		return
	}

	err := h.service.UpdateAutoScalerConfig(context.Background(), clusterName, &config)
	if err != nil {
		logger.Errorf("Failed to update autoscaler config: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, "autoscaler.updateConfigFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "autoscaler.updateConfigSuccess",
	})
}

// GetAutoScalerStatus 获取自动扩缩容状态
func (h *AutoScalerHandler) GetAutoScalerStatus(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "autoscaler.clusterNameEmpty")
		return
	}

	status, err := h.service.GetAutoScalerStatus(context.Background(), clusterName)
	if err != nil {
		logger.Errorf("Failed to get autoscaler status: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, "autoscaler.getStatusFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"status": status,
	})
}
