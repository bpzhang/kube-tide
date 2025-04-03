package api

import (
	"context"
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

// ScaleDeployment Adjust Deployment replica count
func (h *DeploymentHandler) ScaleDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Scaling deployment for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)
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

	var requestBody struct {
		Replicas int32 `json:"replicas" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err := h.service.ScaleDeployment(clusterName, namespace, deploymentName, requestBody.Replicas)
	if err != nil {
		logger.Error("Failed to scale deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// RestartDeployment Restart Deployment
func (h *DeploymentHandler) RestartDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")
	logger.Info("Restarting deployment for cluster: " + clusterName + ", namespace: " + namespace + ", deployment: " + deploymentName)

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

	err := h.service.RestartDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		logger.Error("Failed to restart deployment: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
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
