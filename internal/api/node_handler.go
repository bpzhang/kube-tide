package api

import (
	"context"
	"net/http"
	"strconv"
	"kube-tide/internal/utils/logger"
	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

// NodeHandler node management handler
type NodeHandler struct {
	service *k8s.NodeService
}

// NewNodeHandler create a new NodeHandler
func NewNodeHandler(service *k8s.NodeService) *NodeHandler {
	return &NodeHandler{
		service: service,
	}
}

// ListNodes get all nodes in the specified cluster
func (h *NodeHandler) ListNodes(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}

	page := 1
	limit := 10 

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	logger.Infof("page: %s, limit: %s", pageStr, limitStr)
	var err error
	if page, err = strconv.Atoi(pageStr); err != nil || page < 1 {
		page = 1
	}

	if limit, err = strconv.Atoi(limitStr); err != nil || limit < 1 {
		limit = 10
	}

	// max limit
	if limit > 100 {
		limit = 100
	}

	// max limit for pagination nodes data
	nodes, total, err := h.service.GetNodes(context.Background(), clusterName, limit, page)
	if err != nil {
		logger.Errorf("Failed to get nodes: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, "Failed to retrieve nodes: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"nodes": nodes,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + limit - 1) / limit, 
		},
	})
}

// GetNodeDetails get node details
func (h *NodeHandler) GetNodeDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Getting node details for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Node name cannot be empty")
		return
	}

	node, err := h.service.GetNodeDetails(context.Background(), clusterName, nodeName)
	if err != nil {
		logger.Errorf("Failed to get node details: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"node": node,
	})
}

// GetNodeMetrics get node metrics
func (h *NodeHandler) GetNodeMetrics(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Getting node metrics for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "节点名称不能为空")
		return
	}

	metrics, err := h.service.GetNodeMetrics(context.Background(), clusterName, nodeName)
	if err != nil {
		logger.Errorf("Failed to get node metrics: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"metrics": metrics,
	})
}

// DrainNode node drain
func (h *NodeHandler) DrainNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Draining node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name or node name cannot be empty")
		return
	}

	var params struct {
		GracePeriodSeconds int  `json:"gracePeriodSeconds"`
		DeleteLocalData    bool `json:"deleteLocalData"`
		IgnoreDaemonSets   bool `json:"ignoreDaemonSets"`
	}

	if err := c.BindJSON(&params); err != nil {
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	// Set default values
	if params.GracePeriodSeconds <= 0 {
		params.GracePeriodSeconds = 300 // Default 5 minutes
	}

	err := h.service.DrainNode(
		context.Background(),
		clusterName,
		nodeName,
		params.GracePeriodSeconds,
		params.DeleteLocalData,
		params.IgnoreDaemonSets,
	)
	if err != nil {
		logger.Errorf("Failed to drain node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Node drained successfully",
	})
}

// CordonNode set node to unschedulable
func (h *NodeHandler) CordonNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Cordoning node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	err := h.service.CordonNode(context.Background(), clusterName, nodeName)
	if err != nil {
		logger.Errorf("Failed to cordon node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Node cordoned successfully",
	})
}

// UncordonNode set node to schedulable
func (h *NodeHandler) UncordonNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	err := h.service.UncordonNode(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Node uncordoned successfully",
	})
}

// get node taints
func (h *NodeHandler) GetNodeTaints(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	taints, err := h.service.GetNodeTaints(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"taints": taints,
	})
}

// add node taint
func (h *NodeHandler) AddNodeTaint(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Adding taint to node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	var req struct {
		Key    string             `json:"key" binding:"required"`
		Value  string             `json:"value"`
		Effect corev1.TaintEffect `json:"effect" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	taint := corev1.Taint{
		Key:    req.Key,
		Value:  req.Value,
		Effect: req.Effect,
	}

	if err := h.service.AddNodeTaint(context.Background(), clusterName, nodeName, taint); err != nil {
		logger.Errorf("Failed to add taint to node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Taint added successfully",
	})
}

// delete node taint
func (h *NodeHandler) RemoveNodeTaint(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Removing taint from node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	var req struct {
		Key    string             `json:"key" binding:"required"`
		Effect corev1.TaintEffect `json:"effect" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	if err := h.service.RemoveNodeTaint(context.Background(), clusterName, nodeName, req.Key, req.Effect); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Taint removed successfully",
	})
}

// get node labels
func (h *NodeHandler) GetNodeLabels(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Getting labels for node in cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	labels, err := h.service.GetNodeLabels(context.Background(), clusterName, nodeName)
	if err != nil {
		logger.Errorf("Failed to get node labels: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"labels": labels,
	})
}

// add or update node label
func (h *NodeHandler) AddNodeLabel(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Adding label to node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Failed to bind JSON: %s", err.Error())
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	if err := h.service.AddNodeLabel(context.Background(), clusterName, nodeName, req.Key, req.Value); err != nil {
		logger.Errorf("Failed to add label to node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Label added successfully",
	})
}

// delete node label
func (h *NodeHandler) RemoveNodeLabel(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	logger.Infof("Removing label from node for cluster: %s, node: %s", clusterName, nodeName)
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name or node name cannot be empty")
		return
	}

	var req struct {
		Key string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	if err := h.service.RemoveNodeLabel(context.Background(), clusterName, nodeName, req.Key); err != nil {
		logger.Errorf("Failed to remove label from node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Label removed successfully",
	})
}

// AddNode 
func (h *NodeHandler) AddNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}

	var nodeConfig k8s.NodeConfig
	if err := c.ShouldBindJSON(&nodeConfig); err != nil {
		ResponseError(c, http.StatusBadRequest, "Request parameters are incorrect: "+err.Error())
		return
	}

	
	if nodeConfig.SSHPort == 0 {
		nodeConfig.SSHPort = 22
	}
	if nodeConfig.SSHUser == "" {
		nodeConfig.SSHUser = "root"
	}
	if nodeConfig.AuthType == "" {
		nodeConfig.AuthType = "key" 
	}

	// validate authentication method
	if nodeConfig.AuthType != "key" && nodeConfig.AuthType != "password" {
		ResponseError(c, http.StatusBadRequest, "Authentication method is incorrect, must be 'key' or 'password'")
		return
	}

	// validate necessary parameters based on authentication method
	if nodeConfig.AuthType == "key" && nodeConfig.SSHKeyFile == "" {
		ResponseError(c, http.StatusBadRequest, "SSH key file path cannot be empty when using key authentication")
		return
	}
	if nodeConfig.AuthType == "password" && nodeConfig.SSHPassword == "" {
		ResponseError(c, http.StatusBadRequest, "SSH password cannot be empty when using password authentication")
		return
	}

	err := h.service.AddNode(context.Background(), clusterName, nodeConfig)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Node added successfully",
	})
}

// RemoveNode 
func (h *NodeHandler) RemoveNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	var params struct {
		Force bool `json:"force"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		params.Force = false // Default to non-force delete
	}

	err := h.service.RemoveNode(context.Background(), clusterName, nodeName, params.Force)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Node removed successfully",
	})
}

// GetNodePods get pods on the node
func (h *NodeHandler) GetNodePods(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or node name cannot be empty")
		return
	}

	pods, err := h.service.GetNodePods(context.Background(), clusterName, nodeName)
	if err != nil {
		logger.Errorf("Failed to get pods on node: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}
