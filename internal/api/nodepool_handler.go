package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// NodePoolHandler 节点池管理处理器
type NodePoolHandler struct {
	service *k8s.NodePoolService
}

// NewNodePoolHandler 创建节点池管理处理器
func NewNodePoolHandler(service *k8s.NodePoolService) *NodePoolHandler {
	return &NodePoolHandler{
		service: service,
	}
}

// ListNodePools 获取节点池列表
func (h *NodePoolHandler) ListNodePools(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	pools, err := h.service.ListNodePools(context.Background(), clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pools": pools,
	})
}

// CreateNodePool 创建新的节点池
func (h *NodePoolHandler) CreateNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	var pool k8s.NodePool
	if err := c.ShouldBindJSON(&pool); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	if pool.Name == "" {
		ResponseError(c, http.StatusBadRequest, "节点池名称不能为空")
		return
	}

	err := h.service.CreateNodePool(context.Background(), clusterName, pool)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点池创建成功",
	})
}

// UpdateNodePool 更新节点池
func (h *NodePoolHandler) UpdateNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点池名称不能为空")
		return
	}

	var pool k8s.NodePool
	if err := c.ShouldBindJSON(&pool); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	// 确保path参数和body中的名称一致
	pool.Name = poolName

	err := h.service.UpdateNodePool(context.Background(), clusterName, pool)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点池更新成功",
	})
}

// DeleteNodePool 删除节点池
func (h *NodePoolHandler) DeleteNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点池名称不能为空")
		return
	}

	err := h.service.DeleteNodePool(context.Background(), clusterName, poolName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点池删除成功",
	})
}

// GetNodePool 获取节点池详情
func (h *NodePoolHandler) GetNodePool(c *gin.Context) {
	clusterName := c.Param("cluster")
	poolName := c.Param("pool")
	if clusterName == "" || poolName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点池名称不能为空")
		return
	}

	pool, err := h.service.GetNodePool(context.Background(), clusterName, poolName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pool": pool,
	})
}
