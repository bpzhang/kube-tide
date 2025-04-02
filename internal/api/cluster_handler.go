package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"

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
	// 使用通用操作日志记录
	err := logger.LogOperation("获取集群列表", func() error {
		clusters := h.clientManager.ListClusters()
		ResponseSuccess(c, gin.H{
			"clusters": clusters,
		})
		return nil
	})

	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
	}
}

// AddCluster 添加集群
func (h *ClusterHandler) AddCluster(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		KubeconfigPath string `json:"kubeconfigPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("无效的添加集群请求", "error", err.Error())
		ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	logger.Info("添加集群请求", "name", req.Name, "configPath", req.KubeconfigPath)

	// 使用操作日志记录
	err := logger.LogOperation("添加集群", func() error {
		return h.clientManager.AddCluster(req.Name, req.KubeconfigPath)
	})

	if err != nil {
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

	logger.Info("删除集群", "name", clusterName)

	err := logger.LogOperation("删除集群", func() error {
		h.clientManager.RemoveCluster(clusterName)
		return nil
	})

	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// TestConnection 测试集群连接
func (h *ClusterHandler) TestConnection(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	// 使用通用日志函数
	_, err := logger.LogFunc("测试集群连接", func() (interface{}, error) {
		logger.Info("测试连接到集群", "clusterName", clusterName)
		return nil, h.clientManager.TestConnection(clusterName)
	})

	if err != nil {
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

	// 使用通用日志函数记录操作，并添加上下文
	result, err := logger.LogFuncWithContext(c.Request.Context(), "获取集群详情", func(ctx context.Context) (interface{}, error) {
		logger.Info("获取集群详情", "clusterName", clusterName)

		client, err := h.clientManager.GetClient(clusterName)
		if err != nil {
			return nil, err
		}

		// 获取集群版本信息
		version, err := client.ServerVersion()
		if err != nil {
			return nil, fmt.Errorf("获取集群版本失败: %v", err)
		}

		// 获取命名空间列表
		namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取命名空间列表失败: %v", err)
		}

		// 获取节点列表以统计集群资源
		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取节点列表失败: %v", err)
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

		return gin.H{
			"name":            clusterName,
			"version":         version.String(),
			"totalNodes":      len(nodes.Items),
			"totalNamespaces": len(namespaces.Items),
			"namespaces":      namespaces.Items,
			"totalCPU":        totalCPU,
			"totalMemory":     fmt.Sprintf("%.2f GB", totalMemoryGB),
			"platform":        version.Platform,
		}, nil
	})

	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{"cluster": result})
}

// GetClusterMetrics 获取集群监控指标
func (h *ClusterHandler) GetClusterMetrics(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	// 使用通用日志函数
	metrics, err := logger.LogFunc("获取集群监控指标", func() (interface{}, error) {
		logger.Info("获取集群监控指标", "clusterName", clusterName)

		// 获取集群客户端
		client, err := h.clientManager.GetClient(clusterName)
		if err != nil {
			return nil, err
		}

		// 获取集群监控指标
		return k8s.GetClusterMetrics(client)
	})

	if err != nil {
		ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取集群监控指标失败: %v", err))
		return
	}

	ResponseSuccess(c, gin.H{
		"metrics": metrics,
	})
}

// GetClusterEvents 获取集群事件列表
func (h *ClusterHandler) GetClusterEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	// 使用通用日志函数和上下文
	events, err := logger.LogFuncWithContext(c.Request.Context(), "获取集群事件", func(ctx context.Context) (interface{}, error) {
		logger.Info("获取集群事件", "clusterName", clusterName)

		// 获取集群客户端
		client, err := h.clientManager.GetClient(clusterName)
		if err != nil {
			return nil, err
		}

		// 获取集群范围内的事件
		events, err := client.CoreV1().Events("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取集群事件失败: %v", err)
		}

		// 按时间倒序排序，最新的事件在前面
		sort.Slice(events.Items, func(i, j int) bool {
			return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
		})

		return events.Items, nil
	})

	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": events,
	})
}
