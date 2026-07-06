package api

import (
	"context"
	"kube-tide/internal/utils/logger"
	"net/http"

	"github.com/gin-gonic/gin"

	"kube-tide/internal/core/k8s"
)

// NamespaceHandler namespace management handler
type NamespaceHandler struct {
	namespaceService *k8s.NamespaceService
}

// NewNamespaceHandler create a new NamespaceHandler
func NewNamespaceHandler(namespaceService *k8s.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: namespaceService,
	}
}

// ListNamespacesResponse is the response structure for listing namespaces
type ListNamespacesResponse struct {
	Namespaces []string            `json:"namespaces"`
	Items      []k8s.NamespaceInfo `json:"items"`
}

// ListNamespaces Get the list of namespaces for the specified cluster
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	clusterName := c.Param("cluster")
	logger.Info("Listing namespaces for cluster: " + clusterName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}

	result, err := h.namespaceService.ListNamespaces(clusterName)
	if err != nil {
		logger.Error("Failed to list namespaces: " + err.Error())
		FailWithError(c, http.StatusInternalServerError, "namespace.fetchFailed", err)
		return
	}

	ResponseSuccess(c, ListNamespacesResponse{
		Namespaces: result.Namespaces,
		Items:      result.Items,
	})
}

// GetNamespace 获取命名空间详情
func (h *NamespaceHandler) GetNamespace(c *gin.Context) {
	item, err := h.namespaceService.GetNamespace(context.Background(), c.Param("cluster"), c.Param("namespace"))
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "namespace.fetchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"namespace": item})
}

// CreateNamespace 创建命名空间
func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
	var req k8s.CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "namespace.invalidRequest", err.Error())
		return
	}
	item, err := h.namespaceService.CreateNamespace(context.Background(), c.Param("cluster"), req)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "namespace.createFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"namespace": item})
}

// DeleteNamespace 删除命名空间
func (h *NamespaceHandler) DeleteNamespace(c *gin.Context) {
	if err := h.namespaceService.DeleteNamespace(context.Background(), c.Param("cluster"), c.Param("namespace")); err != nil {
		FailWithError(c, http.StatusInternalServerError, "namespace.deleteFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"message": "namespace.deleteSuccess"})
}

// PatchNamespaceLabels 更新命名空间标签
func (h *NamespaceHandler) PatchNamespaceLabels(c *gin.Context) {
	var req k8s.PatchNamespaceLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "namespace.invalidRequest", err.Error())
		return
	}
	item, err := h.namespaceService.PatchNamespaceLabels(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "namespace.patchFailed", err)
		return
	}
	ResponseSuccess(c, gin.H{"namespace": item})
}
