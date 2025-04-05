package api

import (
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

// ListNamespaces Get the list of namespaces for the specified cluster
// @Summary Get namespaces list
// @Description Retrieve all namespaces in the specified cluster
// @Tags Namespace
// @Accept json
// @Produce json
// @Param clusterName path string true "Cluster name"
// @Success 200 {object} Response{data=ListNamespacesResponse} "Successfully retrieved namespaces list"
// @Failure 400 {object} Response "Bad request"
// @Failure 500 {object} Response "Internal server error"
// @Router /api/v1/clusters/{clusterName}/namespaces [get]
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	clusterName := c.Param("cluster")
	logger.Info("Listing namespaces for cluster: " + clusterName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}

	namespaces, err := h.namespaceService.ListNamespaces(clusterName)
	if err != nil {
		logger.Error("Failed to list namespaces: " + err.Error())
		FailWithError(c, http.StatusInternalServerError, "namespace.fetchFailed", err)
		return
	}

	ResponseSuccess(c, ListNamespacesResponse{
		Namespaces: namespaces,
	})
}

// ListNamespacesResponse is the response structure for listing namespaces
type ListNamespacesResponse struct {
	Namespaces []string `json:"namespaces"`
}
