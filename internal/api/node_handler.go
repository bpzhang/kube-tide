package api

import (
	"context"
	"net/http"
	"strconv"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

// NodeHandler 节点管理处理器
type NodeHandler struct {
	service *k8s.NodeService
}

// NewNodeHandler 创建节点管理处理器
func NewNodeHandler(service *k8s.NodeService) *NodeHandler {
	return &NodeHandler{
		service: service,
	}
}

// ListNodes 获取节点列表
func (h *NodeHandler) ListNodes(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	// 解析分页参数
	page := 1
	limit := 10 // 默认每页10条记录

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	var err error
	if page, err = strconv.Atoi(pageStr); err != nil || page < 1 {
		page = 1
	}

	if limit, err = strconv.Atoi(limitStr); err != nil || limit < 1 {
		limit = 10
	}

	// 最大限制，避免一次请求过多数据
	if limit > 100 {
		limit = 100
	}

	// 获取分页节点数据
	nodes, total, err := h.service.GetNodes(context.Background(), clusterName, limit, page)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回分页数据
	ResponseSuccess(c, gin.H{
		"nodes": nodes,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + limit - 1) / limit, // 计算总页数
		},
	})
}

// GetNodeDetails 获取节点详情
func (h *NodeHandler) GetNodeDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "节点名称不能为空")
		return
	}

	node, err := h.service.GetNodeDetails(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"node": node,
	})
}

// GetNodeMetrics 获取节点指标
func (h *NodeHandler) GetNodeMetrics(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
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
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"metrics": metrics,
	})
}

// DrainNode 对节点进行排水操作
func (h *NodeHandler) DrainNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var params struct {
		GracePeriodSeconds int  `json:"gracePeriodSeconds"`
		DeleteLocalData    bool `json:"deleteLocalData"`
		IgnoreDaemonSets   bool `json:"ignoreDaemonSets"`
	}

	if err := c.BindJSON(&params); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	// 设置默认值
	if params.GracePeriodSeconds <= 0 {
		params.GracePeriodSeconds = 300 // 默认5分钟
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
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点排水操作执行成功",
	})
}

// CordonNode 将节点设置为不可调度
func (h *NodeHandler) CordonNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	err := h.service.CordonNode(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点已设置为不可调度",
	})
}

// UncordonNode 将节点设置为可调度
func (h *NodeHandler) UncordonNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	err := h.service.UncordonNode(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点已设置为可调度",
	})
}

// 获取节点污点
func (h *NodeHandler) GetNodeTaints(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
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

// 添加节点污点
func (h *NodeHandler) AddNodeTaint(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var req struct {
		Key    string             `json:"key" binding:"required"`
		Value  string             `json:"value"`
		Effect corev1.TaintEffect `json:"effect" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	taint := corev1.Taint{
		Key:    req.Key,
		Value:  req.Value,
		Effect: req.Effect,
	}

	if err := h.service.AddNodeTaint(context.Background(), clusterName, nodeName, taint); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "污点添加成功",
	})
}

// 删除节点污点
func (h *NodeHandler) RemoveNodeTaint(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var req struct {
		Key    string             `json:"key" binding:"required"`
		Effect corev1.TaintEffect `json:"effect" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	if err := h.service.RemoveNodeTaint(context.Background(), clusterName, nodeName, req.Key, req.Effect); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "污点删除成功",
	})
}

// 获取节点标签
func (h *NodeHandler) GetNodeLabels(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	labels, err := h.service.GetNodeLabels(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"labels": labels,
	})
}

// 添加或更新节点标签
func (h *NodeHandler) AddNodeLabel(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	if err := h.service.AddNodeLabel(context.Background(), clusterName, nodeName, req.Key, req.Value); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "标签添加成功",
	})
}

// 删除节点标签
func (h *NodeHandler) RemoveNodeLabel(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var req struct {
		Key string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	if err := h.service.RemoveNodeLabel(context.Background(), clusterName, nodeName, req.Key); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "标签删除成功",
	})
}

// AddNode 添加新节点
func (h *NodeHandler) AddNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	var nodeConfig k8s.NodeConfig
	if err := c.ShouldBindJSON(&nodeConfig); err != nil {
		ResponseError(c, http.StatusBadRequest, "请求参数不正确: "+err.Error())
		return
	}

	// 设置默认值
	if nodeConfig.SSHPort == 0 {
		nodeConfig.SSHPort = 22
	}
	if nodeConfig.SSHUser == "" {
		nodeConfig.SSHUser = "root"
	}
	if nodeConfig.AuthType == "" {
		nodeConfig.AuthType = "key" // 默认使用密钥方式
	}

	// 验证认证方式
	if nodeConfig.AuthType != "key" && nodeConfig.AuthType != "password" {
		ResponseError(c, http.StatusBadRequest, "认证方式不正确，必须是 'key' 或 'password'")
		return
	}

	// 根据认证方式验证必要参数
	if nodeConfig.AuthType == "key" && nodeConfig.SSHKeyFile == "" {
		ResponseError(c, http.StatusBadRequest, "使用密钥认证时，SSH密钥文件路径不能为空")
		return
	}
	if nodeConfig.AuthType == "password" && nodeConfig.SSHPassword == "" {
		ResponseError(c, http.StatusBadRequest, "使用密码认证时，SSH密码不能为空")
		return
	}

	err := h.service.AddNode(context.Background(), clusterName, nodeConfig)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点添加成功",
	})
}

// RemoveNode 移除节点
func (h *NodeHandler) RemoveNode(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	var params struct {
		Force bool `json:"force"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		params.Force = false // 默认非强制删除
	}

	err := h.service.RemoveNode(context.Background(), clusterName, nodeName, params.Force)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "节点移除成功",
	})
}

// GetNodePods 获取节点上运行的Pod列表
func (h *NodeHandler) GetNodePods(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	if clusterName == "" || nodeName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或节点名称不能为空")
		return
	}

	pods, err := h.service.GetNodePods(context.Background(), clusterName, nodeName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}
