package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kube-tide/internal/core"
	"kube-tide/internal/core/k8s"
	"kube-tide/internal/database/models"
	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeploymentHandler Deployment management handler
type DeploymentHandler struct {
	service   *k8s.DeploymentService
	dbService core.DeploymentService // 新增数据库服务
}

// NewDeploymentHandler Create Deployment management handler
func NewDeploymentHandler(service *k8s.DeploymentService, dbService core.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		service:   service,
		dbService: dbService,
	}
}

// syncDeploymentToDB 同步部署信息到数据库
func (h *DeploymentHandler) syncDeploymentToDB(clusterName, namespace, deploymentName string) {
	if h.dbService == nil {
		return // 如果没有数据库服务，跳过同步
	}

	// 获取 Kubernetes 中的部署详情
	k8sDeployment, err := h.service.GetDeploymentDetails(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to get deployment details for sync: " + err.Error())
		return
	}

	// 序列化复杂字段为 JSON 字符串
	labelsJSON := ""
	if k8sDeployment.Labels != nil {
		if labelsData, err := json.Marshal(k8sDeployment.Labels); err == nil {
			labelsJSON = string(labelsData)
		}
	}

	annotationsJSON := ""
	if k8sDeployment.Annotations != nil {
		if annotationsData, err := json.Marshal(k8sDeployment.Annotations); err == nil {
			annotationsJSON = string(annotationsData)
		}
	}

	selectorJSON := ""
	if k8sDeployment.Selector != nil {
		if selectorData, err := json.Marshal(k8sDeployment.Selector); err == nil {
			selectorJSON = string(selectorData)
		}
	}

	templateJSON := ""
	if k8sDeployment.Containers != nil {
		templateData := map[string]interface{}{
			"containers": k8sDeployment.Containers,
		}
		if templateBytes, err := json.Marshal(templateData); err == nil {
			templateJSON = string(templateBytes)
		}
	}

	// 转换为数据库模型
	dbDeployment := &models.Deployment{
		ID:                  uuid.New().String(),
		ClusterID:           clusterName,
		Namespace:           namespace,
		Name:                deploymentName,
		Replicas:            int(k8sDeployment.Replicas),
		ReadyReplicas:       int(k8sDeployment.ReadyReplicas),
		AvailableReplicas:   int(k8sDeployment.ReadyReplicas), // 使用 ReadyReplicas 作为 AvailableReplicas
		UnavailableReplicas: int(k8sDeployment.Replicas - k8sDeployment.ReadyReplicas),
		UpdatedReplicas:     int(k8sDeployment.ReadyReplicas),
		StrategyType:        k8sDeployment.Strategy,
		Labels:              labelsJSON,
		Annotations:         annotationsJSON,
		Selector:            selectorJSON,
		Template:            templateJSON,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// 检查是否已存在
	existing, err := h.dbService.GetByClusterNamespaceAndName(context.Background(), clusterName, namespace, deploymentName)
	if err == nil && existing != nil {
		// 更新现有记录
		updates := models.DeploymentUpdateRequest{
			Replicas:            &dbDeployment.Replicas,
			ReadyReplicas:       &dbDeployment.ReadyReplicas,
			AvailableReplicas:   &dbDeployment.AvailableReplicas,
			UnavailableReplicas: &dbDeployment.UnavailableReplicas,
			UpdatedReplicas:     &dbDeployment.UpdatedReplicas,
			StrategyType:        &dbDeployment.StrategyType,
			Labels:              &dbDeployment.Labels,
			Annotations:         &dbDeployment.Annotations,
			Selector:            &dbDeployment.Selector,
			Template:            &dbDeployment.Template,
		}
		if err := h.dbService.Update(context.Background(), existing.ID, updates); err != nil {
			logger.Error("Failed to update deployment in database: " + err.Error())
		}
	} else {
		// 创建新记录
		if err := h.dbService.Create(context.Background(), dbDeployment); err != nil {
			logger.Error("Failed to create deployment in database: " + err.Error())
		}
	}
}

// ListDeployments Get all Deployments list
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	clusterName := c.Param("cluster")
	logger.Info("Listing deployments for cluster: " + clusterName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}

	deployments, err := h.service.ListDeployments(clusterName)
	if err != nil {
		logger.Error("Failed to list deployments: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go func() {
		for _, deployment := range deployments {
			h.syncDeploymentToDB(clusterName, deployment.Namespace, deployment.Name)
		}
	}()

	ResponseSuccess(c, gin.H{
		"deployments": deployments,
	})
}

// ListDeploymentsByNamespace Get Deployments list for specified namespace
func (h *DeploymentHandler) ListDeploymentsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	logger.Info("Listing deployments for cluster: " + clusterName + ", namespace: " + namespace)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}

	deployments, err := h.service.ListDeploymentsByNamespace(clusterName, namespace)
	if err != nil {
		logger.Error("Failed to list deployments: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go func() {
		for _, deployment := range deployments {
			h.syncDeploymentToDB(clusterName, namespace, deployment.Name)
		}
	}()

	ResponseSuccess(c, gin.H{
		"deployments": deployments,
	})
}

// GetDeploymentDetails Get Deployment details
func (h *DeploymentHandler) GetDeploymentDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Getting deployment details for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment name cannot be empty")
		return
	}

	deployment, err := h.service.GetDeploymentDetails(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to get deployment details: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncDeploymentToDB(clusterName, namespace, deploymentName)

	ResponseSuccess(c, gin.H{
		"deployment": deployment,
	})
}

// ScaleDeployment scales a deployment
func (h *DeploymentHandler) ScaleDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "deployment.deploymentNameEmpty")
		return
	}

	// Parse request body
	var req struct {
		Replicas int32 `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "api.invalidJSON")
		return
	}

	// Scale the deployment
	err := h.service.ScaleDeployment(clusterName, namespace, deploymentName, req.Replicas)
	if err != nil {
		logger.Errorf("Failed to scale deployment %s: %v", deploymentName, err)
		FailWithError(c, http.StatusInternalServerError, "deployment.scaleFailed", err)
		return
	}

	// 异步同步到数据库
	go h.syncDeploymentToDB(clusterName, namespace, deploymentName)

	ResponseSuccess(c, nil)
}

// RestartDeployment restarts a deployment by patching its template annotations
func (h *DeploymentHandler) RestartDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "deployment.deploymentNameEmpty")
		return
	}

	// Restart the deployment
	err := h.service.RestartDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Errorf("Failed to restart deployment %s: %v", deploymentName, err)
		FailWithError(c, http.StatusInternalServerError, "deployment.restartFailed", err)
		return
	}

	// 异步同步到数据库
	go h.syncDeploymentToDB(clusterName, namespace, deploymentName)

	ResponseSuccess(c, nil)
}

// UpdateDeployment Update Deployment configuration
func (h *DeploymentHandler) UpdateDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Updating deployment for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment name cannot be empty")
		return
	}

	var updateRequest k8s.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err := h.service.UpdateDeployment(clusterName, namespace, deploymentName, updateRequest)
	if err != nil {
		logger.Error("Failed to update deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncDeploymentToDB(clusterName, namespace, deploymentName)

	ResponseSuccess(c, gin.H{
		"message": "Deployment updated successfully",
	})
}

// CreateDeployment Create a new Deployment
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	logger.Info("Creating deployment for cluster: " + clusterName + ", namespace: " + namespace)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}

	var createRequest k8s.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	_, err := h.service.CreateDeployment(clusterName, namespace, createRequest)
	if err != nil {
		logger.Error("Failed to create deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncDeploymentToDB(clusterName, namespace, createRequest.Name)

	ResponseSuccess(c, gin.H{
		"message": "Deployment created successfully",
	})
}

// GetAllRelatedEvents Get all related events for a deployment
func (h *DeploymentHandler) GetAllRelatedEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Getting all related events for deployment: " + clusterName + "/" + namespace + "/" + deploymentName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment name cannot be empty")
		return
	}

	events, err := h.service.GetAllDeploymentEvents(context.Background(), clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to get deployment events: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": events,
	})
}

// DeleteDeployment Delete a Deployment
func (h *DeploymentHandler) DeleteDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Deleting deployment for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if deploymentName == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment name cannot be empty")
		return
	}

	err := h.service.DeleteDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to delete deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 从数据库中删除记录
	if h.dbService != nil {
		go func() {
			existing, err := h.dbService.GetByClusterNamespaceAndName(context.Background(), clusterName, namespace, deploymentName)
			if err == nil && existing != nil {
				if err := h.dbService.Delete(context.Background(), existing.ID); err != nil {
					logger.Error("Failed to delete deployment from database: " + err.Error())
				}
			}
		}()
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment deleted successfully",
	})
}
