package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-tide/internal/core"
	"kube-tide/internal/database/models"
)

// DBDeploymentHandler handles deployment operations with database persistence
type DBDeploymentHandler struct {
	service core.DeploymentService
	logger  *zap.Logger
}

// NewDBDeploymentHandler creates a new database deployment handler
func NewDBDeploymentHandler(service core.DeploymentService, logger *zap.Logger) *DBDeploymentHandler {
	return &DBDeploymentHandler{
		service: service,
		logger:  logger,
	}
}

// CreateDeployment creates a new deployment in the database
func (h *DBDeploymentHandler) CreateDeployment(c *gin.Context) {
	var deployment models.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		h.logger.Error("failed to bind deployment JSON", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.service.Create(c.Request.Context(), &deployment); err != nil {
		h.logger.Error("failed to create deployment", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, "Failed to create deployment: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployment": deployment,
	})
}

// GetDeployment retrieves a deployment by ID
func (h *DBDeploymentHandler) GetDeployment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment ID is required")
		return
	}

	deployment, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get deployment", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusNotFound, "Deployment not found: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployment": deployment,
	})
}

// GetDeploymentByName retrieves a deployment by cluster, namespace, and name
func (h *DBDeploymentHandler) GetDeploymentByName(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if clusterID == "" || namespace == "" || name == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID, namespace, and name are required")
		return
	}

	deployment, err := h.service.GetByClusterNamespaceAndName(c.Request.Context(), clusterID, namespace, name)
	if err != nil {
		h.logger.Error("failed to get deployment by name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.String("name", name))
		ResponseError(c, http.StatusNotFound, "Deployment not found: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployment": deployment,
	})
}

// ListDeploymentsByCluster lists deployments in a cluster with pagination
func (h *DBDeploymentHandler) ListDeploymentsByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	params := h.parsePaginationParams(c)
	result, err := h.service.ListByCluster(c.Request.Context(), clusterID, params)
	if err != nil {
		h.logger.Error("failed to list deployments by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to list deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployments":  result.Data,
		"total_count":  result.TotalCount,
		"page":         result.Page,
		"page_size":    result.PageSize,
		"total_pages":  result.TotalPages,
		"has_next":     result.HasNext,
		"has_previous": result.HasPrevious,
	})
}

// ListDeploymentsByNamespace lists deployments in a namespace with pagination
func (h *DBDeploymentHandler) ListDeploymentsByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	params := h.parsePaginationParams(c)
	result, err := h.service.ListByNamespace(c.Request.Context(), clusterID, namespace, params)
	if err != nil {
		h.logger.Error("failed to list deployments by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to list deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployments":  result.Data,
		"total_count":  result.TotalCount,
		"page":         result.Page,
		"page_size":    result.PageSize,
		"total_pages":  result.TotalPages,
		"has_next":     result.HasNext,
		"has_previous": result.HasPrevious,
	})
}

// UpdateDeployment updates a deployment
func (h *DBDeploymentHandler) UpdateDeployment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment ID is required")
		return
	}

	var updates models.DeploymentUpdateRequest
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("failed to bind deployment update JSON", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.service.Update(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("failed to update deployment", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusInternalServerError, "Failed to update deployment: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment updated successfully",
	})
}

// DeleteDeployment deletes a deployment
func (h *DBDeploymentHandler) DeleteDeployment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment ID is required")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete deployment", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete deployment: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment deleted successfully",
	})
}

// DeleteDeploymentsByCluster deletes all deployments in a cluster
func (h *DBDeploymentHandler) DeleteDeploymentsByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	if err := h.service.DeleteByCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to delete deployments by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "All deployments in cluster deleted successfully",
	})
}

// DeleteDeploymentsByNamespace deletes all deployments in a namespace
func (h *DBDeploymentHandler) DeleteDeploymentsByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	if err := h.service.DeleteByNamespace(c.Request.Context(), clusterID, namespace); err != nil {
		h.logger.Error("failed to delete deployments by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "All deployments in namespace deleted successfully",
	})
}

// CountDeploymentsByCluster counts deployments in a cluster
func (h *DBDeploymentHandler) CountDeploymentsByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	count, err := h.service.Count(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to count deployments by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to count deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"count": count,
	})
}

// CountDeploymentsByNamespace counts deployments in a namespace
func (h *DBDeploymentHandler) CountDeploymentsByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	count, err := h.service.CountByNamespace(c.Request.Context(), clusterID, namespace)
	if err != nil {
		h.logger.Error("failed to count deployments by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to count deployments: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"count": count,
	})
}

// parsePaginationParams parses pagination parameters from query string
func (h *DBDeploymentHandler) parsePaginationParams(c *gin.Context) models.PaginationParams {
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return models.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}
