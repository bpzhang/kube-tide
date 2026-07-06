package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// SecretHandler Secret 管理处理器
type SecretHandler struct {
	service *k8s.SecretService
}

// NewSecretHandler 创建 Secret 处理器
func NewSecretHandler(service *k8s.SecretService) *SecretHandler {
	return &SecretHandler{service: service}
}

func (h *SecretHandler) ListSecrets(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	items, err := h.service.ListSecrets(context.Background(), clusterName)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"secrets": items})
}

func (h *SecretHandler) ListSecretsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	items, err := h.service.ListSecretsByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"secrets": items})
}

func (h *SecretHandler) GetSecret(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	item, err := h.service.GetSecret(context.Background(), clusterName, namespace, name)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"secret": item})
}

func (h *SecretHandler) CreateSecret(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	var req k8s.CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "common.invalidRequest")
		return
	}
	item, err := h.service.CreateSecret(context.Background(), clusterName, namespace, req)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.createFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"secret": item})
}

func (h *SecretHandler) UpdateSecret(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	var req struct {
		StringData map[string]string `json:"stringData"`
		Labels     map[string]string `json:"labels,omitempty"`
		Type       string            `json:"type,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "common.invalidRequest")
		return
	}
	item, err := h.service.UpdateSecret(context.Background(), clusterName, namespace, name, req.StringData, req.Labels, req.Type)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.updateFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"secret": item})
}

func (h *SecretHandler) DeleteSecret(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	if err := h.service.DeleteSecret(context.Background(), clusterName, namespace, name); err != nil {
		FailWithError(c, http.StatusInternalServerError, "secret.deleteFailed", err)
		return
	}
	ResponseSuccess(c, nil)
}
