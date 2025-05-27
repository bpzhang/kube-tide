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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceHandler Service management handler
type ServiceHandler struct {
	service   *k8s.ServiceManager
	dbService core.ServiceService
}

// NewServiceHandler Create Service management handler
func NewServiceHandler(service *k8s.ServiceManager, dbService core.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		service:   service,
		dbService: dbService,
	}
}

// syncServiceToDB 同步服务信息到数据库
func (h *ServiceHandler) syncServiceToDB(clusterName, namespace, serviceName string) {
	if h.dbService == nil {
		return // 如果没有数据库服务，跳过同步
	}

	// 获取 Kubernetes 中的服务详情
	k8sService, err := h.service.GetServiceDetails(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		logger.Error("Failed to get service details for sync: " + err.Error())
		return
	}

	// 序列化复杂字段为 JSON 字符串
	portsJSON := ""
	if k8sService.Spec.Ports != nil {
		if portsData, err := json.Marshal(k8sService.Spec.Ports); err == nil {
			portsJSON = string(portsData)
		}
	}

	selectorJSON := ""
	if k8sService.Spec.Selector != nil {
		if selectorData, err := json.Marshal(k8sService.Spec.Selector); err == nil {
			selectorJSON = string(selectorData)
		}
	}

	labelsJSON := ""
	if k8sService.Labels != nil {
		if labelsData, err := json.Marshal(k8sService.Labels); err == nil {
			labelsJSON = string(labelsData)
		}
	}

	annotationsJSON := ""
	if k8sService.Annotations != nil {
		if annotationsData, err := json.Marshal(k8sService.Annotations); err == nil {
			annotationsJSON = string(annotationsData)
		}
	}

	externalIPsJSON := ""
	if k8sService.Spec.ExternalIPs != nil {
		if externalIPsData, err := json.Marshal(k8sService.Spec.ExternalIPs); err == nil {
			externalIPsJSON = string(externalIPsData)
		}
	}

	// 转换为数据库模型
	dbService := &models.Service{
		ID:              uuid.New().String(),
		ClusterID:       clusterName,
		Namespace:       namespace,
		Name:            serviceName,
		Type:            string(k8sService.Spec.Type),
		ClusterIP:       k8sService.Spec.ClusterIP,
		ExternalIPs:     externalIPsJSON,
		Ports:           portsJSON,
		Selector:        selectorJSON,
		SessionAffinity: string(k8sService.Spec.SessionAffinity),
		Labels:          labelsJSON,
		Annotations:     annotationsJSON,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 检查是否已存在
	existing, err := h.dbService.GetByClusterNamespaceAndName(context.Background(), clusterName, namespace, serviceName)
	if err == nil && existing != nil {
		// 更新现有记录
		updates := models.ServiceUpdateRequest{
			Type:        &dbService.Type,
			ClusterIP:   &dbService.ClusterIP,
			ExternalIPs: &dbService.ExternalIPs,
			Ports:       &dbService.Ports,
			Selector:    &dbService.Selector,
			Labels:      &dbService.Labels,
			Annotations: &dbService.Annotations,
		}
		if err := h.dbService.Update(context.Background(), existing.ID, updates); err != nil {
			logger.Error("Failed to update service in database: " + err.Error())
		}
	} else {
		// 创建新记录
		if err := h.dbService.Create(context.Background(), dbService); err != nil {
			logger.Error("Failed to create service in database: " + err.Error())
		}
	}
}

// ListServices Get all Services list
func (h *ServiceHandler) ListServices(c *gin.Context) {
	clusterName := c.Param("cluster")
	logger.Info("Listing services for cluster: " + clusterName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}

	services, err := h.service.GetServices(context.Background(), clusterName)
	if err != nil {
		logger.Error("Failed to list services: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go func() {
		for _, service := range services {
			h.syncServiceToDB(clusterName, service.Namespace, service.Name)
		}
	}()

	ResponseSuccess(c, gin.H{
		"services": services,
	})
}

// ListServicesByNamespace Get Services list for specified namespace
func (h *ServiceHandler) ListServicesByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	logger.Info("Listing services for cluster: " + clusterName + ", namespace: " + namespace)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}

	services, err := h.service.GetServicesByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		logger.Error("Failed to list services: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go func() {
		for _, service := range services {
			h.syncServiceToDB(clusterName, namespace, service.Name)
		}
	}()

	ResponseSuccess(c, gin.H{
		"services": services,
	})
}

// GetServiceDetails Get Service details
func (h *ServiceHandler) GetServiceDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	logger.Info("Getting service details for cluster: " + clusterName + ", namespace: " + namespace + ", service: " + serviceName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Service name cannot be empty")
		return
	}

	service, err := h.service.GetServiceDetails(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		logger.Error("Failed to get service details: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncServiceToDB(clusterName, namespace, serviceName)

	ResponseSuccess(c, gin.H{
		"service": service,
	})
}

// CreateService Create a new Service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	logger.Info("Creating service for cluster: " + clusterName + ", namespace: " + namespace)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}

	var serviceSpec corev1.Service
	if err := c.ShouldBindJSON(&serviceSpec); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	// 确保命名空间正确设置
	serviceSpec.Namespace = namespace

	err := h.service.CreateService(context.Background(), clusterName, &serviceSpec)
	if err != nil {
		logger.Error("Failed to create service: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncServiceToDB(clusterName, namespace, serviceSpec.Name)

	ResponseSuccess(c, gin.H{
		"message": "Service created successfully",
	})
}

// UpdateService Update Service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	logger.Info("Updating service for cluster: " + clusterName + ", namespace: " + namespace + ", service: " + serviceName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Service name cannot be empty")
		return
	}

	var serviceSpec corev1.Service
	if err := c.ShouldBindJSON(&serviceSpec); err != nil {
		logger.Error("Failed to bind JSON: " + err.Error())
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	// 确保命名空间和名称正确设置
	serviceSpec.Namespace = namespace
	serviceSpec.Name = serviceName

	err := h.service.UpdateService(context.Background(), clusterName, &serviceSpec)
	if err != nil {
		logger.Error("Failed to update service: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 异步同步到数据库
	go h.syncServiceToDB(clusterName, namespace, serviceName)

	ResponseSuccess(c, gin.H{
		"message": "Service updated successfully",
	})
}

// DeleteService Delete a Service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	logger.Info("Deleting service for cluster: " + clusterName + ", namespace: " + namespace + ", service: " + serviceName)

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Namespace cannot be empty")
		return
	}
	if serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Service name cannot be empty")
		return
	}

	err := h.service.DeleteService(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		logger.Error("Failed to delete service: " + err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 从数据库中删除记录
	if h.dbService != nil {
		go func() {
			existing, err := h.dbService.GetByClusterNamespaceAndName(context.Background(), clusterName, namespace, serviceName)
			if err == nil && existing != nil {
				if err := h.dbService.Delete(context.Background(), existing.ID); err != nil {
					logger.Error("Failed to delete service from database: " + err.Error())
				}
			}
		}()
	}

	ResponseSuccess(c, gin.H{
		"message": "Service deleted successfully",
	})
}

// ServicePort Service port configuration
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol,omitempty"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// convertToPorts Convert port configuration
func convertToPorts(ports []ServicePort) []corev1.ServicePort {
	result := make([]corev1.ServicePort, len(ports))
	for i, port := range ports {
		result[i] = corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
			Protocol:   corev1.Protocol(port.Protocol),
			NodePort:   port.NodePort,
		}
	}
	return result
}
