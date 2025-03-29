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

// ServiceHandler Service管理处理器
type ServiceHandler struct {
	manager *k8s.ServiceManager
}

// NewServiceHandler 创建Service管理处理器
func NewServiceHandler(manager *k8s.ServiceManager) *ServiceHandler {
	return &ServiceHandler{
		manager: manager,
	}
}

// ListServices 获取所有Service列表
func (h *ServiceHandler) ListServices(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
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

// ListServicesByNamespace 获取指定命名空间的Service列表
func (h *ServiceHandler) ListServicesByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或命名空间不能为空")
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

// GetServiceDetails 获取Service详情
func (h *ServiceHandler) GetServiceDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称、命名空间或服务名称不能为空")
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

// CreateService 创建Service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" || namespace == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称或命名空间不能为空")
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
		ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
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
		"message": "服务创建成功",
	})
}

// UpdateService 更新Service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称、命名空间或服务名称不能为空")
		return
	}

	var req struct {
		Ports       []ServicePort     `json:"ports"`
		Selector    map[string]string `json:"selector"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	// 获取现有的Service
	existingService, err := h.manager.GetServiceDetails(context.Background(), clusterName, namespace, serviceName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 更新Service字段
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
		"message": "服务更新成功",
	})
}

// DeleteService 删除Service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	serviceName := c.Param("service")
	if clusterName == "" || namespace == "" || serviceName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称、命名空间或服务名称不能为空")
		return
	}

	if err := h.manager.DeleteService(context.Background(), clusterName, namespace, serviceName); err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "服务删除成功",
	})
}

// ServicePort Service端口配置
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol,omitempty"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// convertToPorts 转换端口配置
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
