package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceHandler Service management handler
type ServiceHandler struct {
	manager *k8s.ServiceManager
}

// NewServiceHandler Create Service management handler
func NewServiceHandler(manager *k8s.ServiceManager) *ServiceHandler {
	return &ServiceHandler{
		manager: manager,
	}
}

// ListServices Get all Services list
func (h *ServiceHandler) ListServices(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name cannot be empty")
		return
	}

	services, err := h.manager.GetServices(context.Background(), clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"services": services,
	})
}

// ListServicesByNamespace Get Services list by namespace
func (h *ServiceHandler) ListServicesByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or namespace cannot be empty")
		return
	}

	services, err := h.manager.GetServicesByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"services": services,
	})
}

// GetServiceDetails Get Service details
func (h *ServiceHandler) GetServiceDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name, namespace or service name cannot be empty")
		return
	}

	service, err := h.manager.GetServiceDetails(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"service": service,
	})
}

// CreateService Create Service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name or namespace cannot be empty")
		return
	}

	var req struct {
		Name        string            `json:"name" binding:"required"`
		Type        string            `json:"type" binding:"required"`
		Ports       []ServicePort     `json:"ports" binding:"required"`
		Selector    map[string]string `json:"selector"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(req.Type),
			Selector: req.Selector,
			Ports:    convertToPorts(req.Ports),
		},
	}

	if err := h.manager.CreateService(context.Background(), clusterName, service); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Service created successfully",
	})
}

// UpdateService Update Service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name, namespace or service name cannot be empty")
		return
	}

	var req struct {
		Ports       []ServicePort     `json:"ports"`
		Selector    map[string]string `json:"selector"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	// Get existing Service
	existingService, err := h.manager.GetServiceDetails(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Update Service fields
	if req.Labels != nil {
		existingService.Labels = req.Labels
	}
	if req.Annotations != nil {
		existingService.Annotations = req.Annotations
	}
	if req.Selector != nil {
		existingService.Spec.Selector = req.Selector
	}
	if req.Ports != nil {
		existingService.Spec.Ports = convertToPorts(req.Ports)
	}

	if err := h.manager.UpdateService(context.Background(), clusterName, existingService); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Service updated successfully",
	})
}

// DeleteService Delete Service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "Cluster name, namespace or service name cannot be empty")
		return
	}

	if err := h.manager.DeleteService(context.Background(), clusterName, namespace, serviceName); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
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
