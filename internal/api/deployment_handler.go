package api

import (
	"context"
	"fmt"
	"net/http"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

// DeploymentHandler Deployment management handler
type DeploymentHandler struct {
	service *k8s.DeploymentService
}

// NewDeploymentHandler Create Deployment management handler
func NewDeploymentHandler(service *k8s.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		service: service,
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

	ResponseSuccess(c, gin.H{
		"message": "Deployment updated successfully",
	})
}

// CreateDeployment Create new Deployment
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

	// Validate required fields
	if createRequest.Name == "" {
		ResponseError(c, http.StatusBadRequest, "Deployment name cannot be empty")
		return
	}
	if len(createRequest.Containers) == 0 {
		ResponseError(c, http.StatusBadRequest, "At least one container must be defined")
		return
	}

	deployment, err := h.service.CreateDeployment(clusterName, namespace, createRequest)
	if err != nil {
		logger.Error("Failed to create deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message":    "Deployment created successfully",
		"deployment": deployment,
	})
}

// GetAllRelatedEvents Get all events for Deployment and its associated ReplicaSets and Pods
func (h *DeploymentHandler) GetAllRelatedEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Getting all related events for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)

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

	eventMap, err := h.service.GetAllDeploymentEvents(context.Background(), clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to get all related events: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": eventMap,
	})
}

// DeleteDeployment 删除Deployment
func (h *DeploymentHandler) DeleteDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("删除deployment: " + clusterName + ", 命名空间: " + namespace + ", 名称: " + deploymentName)

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

	err := h.service.DeleteDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Errorf("删除Deployment %s 失败: %v", deploymentName, err)
		FailWithError(c, http.StatusInternalServerError, "deployment.deleteFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment deleted successfully",
	})
}

// GetDeploymentRolloutHistory 获取Deployment版本历史
func (h *DeploymentHandler) GetDeploymentRolloutHistory(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("获取Deployment版本历史: " + clusterName + ", 命名空间: " + namespace + ", 名称: " + deploymentName)

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

	revisions, err := h.service.GetDeploymentRolloutHistory(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("获取Deployment版本历史失败: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"revisions": revisions,
	})
}

// GetDeploymentRevisionDetails 获取指定版本详情
func (h *DeploymentHandler) GetDeploymentRevisionDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	revisionStr := c.Param("revision")

	logger.Info("获取Deployment版本详情: " + clusterName + ", 命名空间: " + namespace + ", 名称: " + deploymentName + ", 版本: " + revisionStr)

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
	if revisionStr == "" {
		ResponseError(c, http.StatusBadRequest, "Revision cannot be empty")
		return
	}

	var revision int64
	if _, err := fmt.Sscanf(revisionStr, "%d", &revision); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid revision number")
		return
	}

	revisionDetails, err := h.service.GetDeploymentRevisionDetails(clusterName, namespace, deploymentName, revision)
	if err != nil {
		logger.Error("获取Deployment版本详情失败: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"revision": revisionDetails,
	})
}

// RollbackDeployment 回滚Deployment到指定版本
func (h *DeploymentHandler) RollbackDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	logger.Info("回滚Deployment: " + clusterName + ", 命名空间: " + namespace + ", 名称: " + deploymentName)

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

	var rollbackRequest struct {
		Revision *int64 `json:"revision,omitempty"`
	}

	if err := c.ShouldBindJSON(&rollbackRequest); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	var err error
	if rollbackRequest.Revision != nil {
		// 回滚到指定版本
		err = h.service.RollbackDeployment(clusterName, namespace, deploymentName, *rollbackRequest.Revision)
	} else {
		// 回滚到上一个版本
		err = h.service.RollbackToPreviousRevision(clusterName, namespace, deploymentName)
	}

	if err != nil {
		logger.Error("回滚Deployment失败: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Deployment rollback successfully",
	})
}
