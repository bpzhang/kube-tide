package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"kube-tide/internal/core/k8s"
)

// NamespaceHandler 处理命名空间相关的API请求
type NamespaceHandler struct {
	namespaceService *k8s.NamespaceService
}

// NewNamespaceHandler 创建新的命名空间处理器
func NewNamespaceHandler(namespaceService *k8s.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: namespaceService,
	}
}

// ListNamespaces 获取指定集群的命名空间列表
// @Summary 获取命名空间列表
// @Description 获取指定集群中的所有命名空间
// @Tags 命名空间
// @Accept json
// @Produce json
// @Param clusterName path string true "集群名称"
// @Success 200 {object} Response{data=ListNamespacesResponse} "成功获取命名空间列表"
// @Failure 400 {object} Response "请求错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/clusters/{clusterName}/namespaces [get]
func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	namespaces, err := h.namespaceService.ListNamespaces(clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "获取命名空间列表失败: "+err.Error())
		return
	}

	ResponseSuccess(c, ListNamespacesResponse{
		Namespaces: namespaces,
	})
}

// ListNamespacesResponse 获取命名空间列表的响应
type ListNamespacesResponse struct {
	Namespaces []string `json:"namespaces"`
}
