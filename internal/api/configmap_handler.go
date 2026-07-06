package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// ConfigMapHandler ConfigMap 管理处理器
type ConfigMapHandler struct {
	service *k8s.ConfigMapService
}

// NewConfigMapHandler 创建 ConfigMap 处理器
func NewConfigMapHandler(service *k8s.ConfigMapService) *ConfigMapHandler {
	return &ConfigMapHandler{service: service}
}

func (h *ConfigMapHandler) ListConfigMaps(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	items, err := h.service.ListConfigMaps(context.Background(), clusterName)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"configmaps": items})
}

func (h *ConfigMapHandler) ListConfigMapsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	items, err := h.service.ListConfigMapsByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"configmaps": items})
}

func (h *ConfigMapHandler) GetConfigMap(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	item, err := h.service.GetConfigMap(context.Background(), clusterName, namespace, name)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"configmap": item})
}

func (h *ConfigMapHandler) CreateConfigMap(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	var req k8s.CreateConfigMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "common.invalidRequest")
		return
	}
	item, err := h.service.CreateConfigMap(context.Background(), clusterName, namespace, req)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.createFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"configmap": item})
}

func (h *ConfigMapHandler) UpdateConfigMap(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	var req struct {
		Data   map[string]string `json:"data"`
		Labels map[string]string `json:"labels,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "common.invalidRequest")
		return
	}
	item, err := h.service.UpdateConfigMap(context.Background(), clusterName, namespace, name, req.Data, req.Labels)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.updateFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"configmap": item})
}

func (h *ConfigMapHandler) DeleteConfigMap(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	if err := h.service.DeleteConfigMap(context.Background(), clusterName, namespace, name); err != nil {
		FailWithError(c, http.StatusInternalServerError, "configmap.deleteFailed", err)
		return
	}
	ResponseSuccess(c, nil)
}
