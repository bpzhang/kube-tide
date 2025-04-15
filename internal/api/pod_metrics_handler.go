package api

import (
	"context"
	"net/http"

	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

// GetPodMetrics 获取Pod的CPU和内存监控指标
func (h *PodHandler) GetPodMetrics(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		logger.Errorf("Failed to get pod metrics: namespace cannot be empty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod.podNameEmpty")
		logger.Errorf("Failed to get pod metrics: pod name cannot be empty")
		return
	}

	metrics, err := h.service.GetPodMetrics(context.Background(), clusterName, namespace, podName)
	if err != nil {
		logger.Errorf("Failed to get pod metrics: %s", err.Error())
		FailWithError(c, http.StatusInternalServerError, "pod.metricsFetchFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"metrics": metrics,
	})
}
