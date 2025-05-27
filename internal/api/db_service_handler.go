package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-tide/internal/core"
	"kube-tide/internal/database/models"
)

// DBServiceHandler handles service operations with database persistence
type DBServiceHandler struct {
	service core.ServiceService
	logger  *zap.Logger
}

// NewDBServiceHandler creates a new database service handler
func NewDBServiceHandler(service core.ServiceService, logger *zap.Logger) *DBServiceHandler {
	return &DBServiceHandler{
		service: service,
		logger:  logger,
	}
}

// CreateService creates a new service in the database
func (h *DBServiceHandler) CreateService(c *gin.Context) {
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		h.logger.Error("failed to bind service JSON", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.service.Create(c.Request.Context(), &service); err != nil {
		h.logger.Error("failed to create service", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, "Failed to create service: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"service": service,
	})
}

// GetService retrieves a service by ID
func (h *DBServiceHandler) GetService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	service, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get service", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusNotFound, "Service not found: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"service": service,
	})
}

// GetServiceByName retrieves a service by cluster, namespace, and name
func (h *DBServiceHandler) GetServiceByName(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if clusterID == "" || namespace == "" || name == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID, namespace, and name are required")
		return
	}

	service, err := h.service.GetByClusterNamespaceAndName(c.Request.Context(), clusterID, namespace, name)
	if err != nil {
		h.logger.Error("failed to get service by name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.String("name", name))
		ResponseError(c, http.StatusNotFound, "Service not found: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"service": service,
	})
}

// ListServicesByCluster lists services in a cluster with pagination
func (h *DBServiceHandler) ListServicesByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	params := h.parsePaginationParams(c)
	result, err := h.service.ListByCluster(c.Request.Context(), clusterID, params)
	if err != nil {
		h.logger.Error("failed to list services by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to list services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"services":     result.Data,
		"total_count":  result.TotalCount,
		"page":         result.Page,
		"page_size":    result.PageSize,
		"total_pages":  result.TotalPages,
		"has_next":     result.HasNext,
		"has_previous": result.HasPrevious,
	})
}

// ListServicesByNamespace lists services in a namespace with pagination
func (h *DBServiceHandler) ListServicesByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	params := h.parsePaginationParams(c)
	result, err := h.service.ListByNamespace(c.Request.Context(), clusterID, namespace, params)
	if err != nil {
		h.logger.Error("failed to list services by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to list services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"services":     result.Data,
		"total_count":  result.TotalCount,
		"page":         result.Page,
		"page_size":    result.PageSize,
		"total_pages":  result.TotalPages,
		"has_next":     result.HasNext,
		"has_previous": result.HasPrevious,
	})
}

// UpdateService updates a service
func (h *DBServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	var updates models.ServiceUpdateRequest
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("failed to bind service update JSON", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.service.Update(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("failed to update service", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusInternalServerError, "Failed to update service: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Service updated successfully",
	})
}

// DeleteService deletes a service
func (h *DBServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete service", zap.Error(err), zap.String("id", id))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete service: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Service deleted successfully",
	})
}

// DeleteServicesByCluster deletes all services in a cluster
func (h *DBServiceHandler) DeleteServicesByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	if err := h.service.DeleteByCluster(c.Request.Context(), clusterID); err != nil {
		h.logger.Error("failed to delete services by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "All services in cluster deleted successfully",
	})
}

// DeleteServicesByNamespace deletes all services in a namespace
func (h *DBServiceHandler) DeleteServicesByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	if err := h.service.DeleteByNamespace(c.Request.Context(), clusterID, namespace); err != nil {
		h.logger.Error("failed to delete services by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to delete services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "All services in namespace deleted successfully",
	})
}

// CountServicesByCluster counts services in a cluster
func (h *DBServiceHandler) CountServicesByCluster(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	count, err := h.service.Count(c.Request.Context(), clusterID)
	if err != nil {
		h.logger.Error("failed to count services by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		ResponseError(c, http.StatusInternalServerError, "Failed to count services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"count": count,
	})
}

// CountServicesByNamespace counts services in a namespace
func (h *DBServiceHandler) CountServicesByNamespace(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	namespace := c.Param("namespace")

	if clusterID == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster ID and namespace are required")
		return
	}

	count, err := h.service.CountByNamespace(c.Request.Context(), clusterID, namespace)
	if err != nil {
		h.logger.Error("failed to count services by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		ResponseError(c, http.StatusInternalServerError, "Failed to count services: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"count": count,
	})
}

// parsePaginationParams parses pagination parameters from query string
func (h *DBServiceHandler) parsePaginationParams(c *gin.Context) models.PaginationParams {
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
