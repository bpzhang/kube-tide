package api

import (
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// DeploymentHandler Deployment管理处理器
type DeploymentHandler struct {
	service *k8s.DeploymentService
}

// NewDeploymentHandler 创建Deployment管理处理器
func NewDeploymentHandler(service *k8s.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		service: service,
	}
}

// ListDeployments 获取所有Deployment列表
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	deployments, err := h.service.ListDeployments(clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployments": deployments,
	})
}

// ListDeploymentsByNamespace 获取指定命名空间的Deployment列表
func (h *DeploymentHandler) ListDeploymentsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}

	deployments, err := h.service.ListDeploymentsByNamespace(clusterName, namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployments": deployments,
	})
}

// GetDeploymentDetails 获取Deployment详情
func (h *DeploymentHandler) GetDeploymentDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment名称不能为空")
		return
	}

	deployment, err := h.service.GetDeploymentDetails(clusterName, namespace, deploymentName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"deployment": deployment,
	})
}

// ScaleDeployment 调整Deployment副本数
func (h *DeploymentHandler) ScaleDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment名称不能为空")
		return
	}

	var requestBody struct {
		Replicas int32 `json:"replicas" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	err := h.service.ScaleDeployment(clusterName, namespace, deploymentName, requestBody.Replicas)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// RestartDeployment 重启Deployment
func (h *DeploymentHandler) RestartDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment名称不能为空")
		return
	}

	err := h.service.RestartDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// UpdateDeployment 更新Deployment配置
func (h *DeploymentHandler) UpdateDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment名称不能为空")
		return
	}

	var updateRequest k8s.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	err := h.service.UpdateDeployment(clusterName, namespace, deploymentName, updateRequest)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment更新成功",
	})
}

// CreateDeployment 创建新的Deployment
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}

	var createRequest k8s.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	// 验证必填字段
	if createRequest.Name == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment名称不能为空")
		return
	}
	if len(createRequest.Containers) == 0 {
		ResponseError(c, http.StatusBadRequest, "至少需要定义一个容器")
		return
	}

	deployment, err := h.service.CreateDeployment(clusterName, namespace, createRequest)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message":    "Deployment创建成功",
		"deployment": deployment,
	})
}
