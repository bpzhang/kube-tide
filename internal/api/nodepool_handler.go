package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

// NodePoolHandler get node pool management handler
type NodePoolHandler struct {
	service *k8s.NodePoolService
}

// NewNodePoolHandler create node pool management handler
func NewNodePoolHandler(service *k8s.NodePoolService) *NodePoolHandler {
	return &NodePoolHandler{
		service: service,
	}
}

// ListNodePools get node pool list
func (h *NodePoolHandler) ListNodePools(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}

	pools, err := h.service.ListNodePools(context.Background(), clusterName)
	if err != nil {
		logger.Errorf("Failed to list node pools: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pools": pools,
	})
}

// CreateNodePool create new node pool
func (h *NodePoolHandler) CreateNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}

	var pool k8s.NodePool
	if err := c.ShouldBindJSON(&pool); err != nil {
		ResponseError(c, http.StatusBadRequest, "request parameters are incorrect: "+err.Error())
		return
	}

	if pool.Name == "" {
		ResponseError(c, http.StatusBadRequest, "node pool name cannot be empty")
		return
	}

	err := h.service.CreateNodePool(context.Background(), clusterName, pool)
	if err != nil {
		logger.Errorf("Failed to create node pool: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "node pool created successfully",
	})
}

// UpdateNodePool update node pool
func (h *NodePoolHandler) UpdateNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name or node pool name cannot be empty")
		return
	}

	var pool k8s.NodePool
	if err := c.ShouldBindJSON(&pool); err != nil {
		ResponseError(c, http.StatusBadRequest, "request parameters are incorrect: "+err.Error())
		return
	}

	// ensure path parameters and body names are consistent
	pool.Name = poolName

	err := h.service.UpdateNodePool(context.Background(), clusterName, pool)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "node pool updated successfully",
	})
}

// DeleteNodePool delete node pool
func (h *NodePoolHandler) DeleteNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name or node pool name cannot be empty")
		return
	}

	err := h.service.DeleteNodePool(context.Background(), clusterName, poolName)
	if err != nil {
		logger.Errorf("Failed to delete node pool: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "node pool deleted successfully",
	})
}

// GetNodePool get node pool details
func (h *NodePoolHandler) GetNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name or node pool name cannot be empty")
		return
	}

	pool, err := h.service.GetNodePool(context.Background(), clusterName, poolName)
	if err != nil {
		logger.Errorf("Failed to get node pool: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pool": pool,
	})
}
