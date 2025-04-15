package handlers

import (
	"net/http"
	"kube-tide/internal/utils/logger"
	"kube-tide/internal/core/k8s"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

)

// GetPodMetrics 获取Pod的CPU和内存指标
func GetPodMetrics(c *gin.Context) {
	requestID := uuid.New().String()
	logger.Infof("Request ID: %s, Method: %s, Path: %s", requestID, c.Request.Method, c.Request.URL.Path)

	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	// 获取k8s客户端
	client, err := k8s.NewClientManager().GetClient(clusterName)
	if err != nil {
		logger.Errorf( "获取集群客户端失败: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群客户端失败: " + err.Error(),
		})
		return
	}

	// 获取Pod指标
	metrics, err := k8s.GetPodMetrics(client, namespace, podName)
	if err != nil {
		logger.Errorf( "获取Pod指标失败: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取Pod指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"metrics": metrics,
		},
	})
}
