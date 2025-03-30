package api

import (
	"fmt"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterHandler 集群管理处理器
type ClusterHandler struct {
	clientManager *k8s.ClientManager
}

// NewClusterHandler 创建集群管理处理器
func NewClusterHandler(clientManager *k8s.ClientManager) *ClusterHandler {
	return &ClusterHandler{
		clientManager: clientManager,
	}
}

// ListClusters 获取集群列表
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	clusters := h.clientManager.ListClusters()
	ResponseSuccess(c, gin.H{
		"clusters": clusters,
	})
}

// AddCluster 添加集群
func (h *ClusterHandler) AddCluster(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		KubeconfigPath string `json:"kubeconfigPath" binding:"required"`
	}
	// print log to console
	fmt.Println("AddCluster request:", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := h.clientManager.AddCluster(req.Name, req.KubeconfigPath); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// RemoveCluster 删除集群
func (h *ClusterHandler) RemoveCluster(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	h.clientManager.RemoveCluster(clusterName)
	ResponseSuccess(c, nil)
}

// TestConnection 测试集群连接
func (h *ClusterHandler) TestConnection(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	if err := h.clientManager.TestConnection(clusterName); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"status": "connected",
	})
}

// GetClusterDetails 获取集群详细信息
func (h *ClusterHandler) GetClusterDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	client, err := h.clientManager.GetClient(clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 获取集群版本信息
	version, err := client.ServerVersion()
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取集群版本失败: %v", err))
		return
	}

	// 获取命名空间列表
	namespaces, err := client.CoreV1().Namespaces().List(c, metav1.ListOptions{})
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取命名空间列表失败: %v", err))
		return
	}

	// 获取节点列表以统计集群资源
	nodes, err := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取节点列表失败: %v", err))
		return
	}

	// 统计集群总资源
	var totalCPU, totalMemory int64
	for _, node := range nodes.Items {
		cpu := node.Status.Capacity.Cpu()
		memory := node.Status.Capacity.Memory()
		totalCPU += cpu.Value()
		totalMemory += memory.Value()
	}

	// 转换内存单位为GB
	totalMemoryGB := float64(totalMemory) / (1024 * 1024 * 1024)

	ResponseSuccess(c, gin.H{
		"cluster": gin.H{
			"name":            clusterName,
			"version":         version.String(),
			"totalNodes":      len(nodes.Items),
			"totalNamespaces": len(namespaces.Items),
			"namespaces":      namespaces.Items,
			"totalCPU":        totalCPU,
			"totalMemory":     fmt.Sprintf("%.2f GB", totalMemoryGB),
			"platform":        version.Platform,
		},
	})
}

// GetClusterMetrics 获取集群监控指标
func (h *ClusterHandler) GetClusterMetrics(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	// 获取集群客户端
	client, err := h.clientManager.GetClient(clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 获取集群监控指标
	metrics, err := k8s.GetClusterMetrics(client)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取集群监控指标失败: %v", err))
		return
	}

	ResponseSuccess(c, gin.H{
		"metrics": metrics,
	})
}
