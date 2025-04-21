package api

import (
	"context"
	"fmt"
	"net/http"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetHandler 处理StatefulSet相关的API请求
type StatefulSetHandler struct {
	service *k8s.StatefulSetService
}

// NewStatefulSetHandler 创建StatefulSet处理器
func NewStatefulSetHandler(service *k8s.StatefulSetService) *StatefulSetHandler {
	return &StatefulSetHandler{
		service: service,
	}
}

// ListStatefulSets 获取所有StatefulSet列表
func (h *StatefulSetHandler) ListStatefulSets(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Query("namespace")

	if namespace == "" {
		namespace = "default"
	}

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}

	statefulsets, err := h.service.GetStatefulSets(context.Background(), clusterName, namespace)
	if err != nil {
		logger.Errorf("获取StatefulSet列表失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.listFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"statefulsets": statefulsets,
	})
}

// GetStatefulSetDetails 获取StatefulSet详情
func (h *StatefulSetHandler) GetStatefulSetDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	statefulset, err := h.service.GetStatefulSetDetails(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet详情失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.detailsFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"statefulset": statefulset,
	})
}

// CreateStatefulSet 创建新的StatefulSet
func (h *StatefulSetHandler) CreateStatefulSet(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}

	var request struct {
		Name                 string            `json:"name" binding:"required"`
		Namespace            string            `json:"namespace" binding:"required"`
		Replicas             int32             `json:"replicas"`
		ServiceName          string            `json:"serviceName" binding:"required"`
		Labels               map[string]string `json:"labels"`
		Annotations          map[string]string `json:"annotations"`
		Containers           []map[string]any  `json:"containers" binding:"required"`
		PodManagementPolicy  string            `json:"podManagementPolicy"`
		UpdateStrategy       string            `json:"updateStrategy"`
		VolumeClaimTemplates []map[string]any  `json:"volumeClaimTemplates"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		ResponseError(c, http.StatusBadRequest, "statefulsets.invalidRequestFormat", err.Error())
		return
	}

	// 校验基本参数
	if request.Name == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}
	if request.Namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if request.ServiceName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.serviceNameEmpty")
		return
	}
	if len(request.Containers) == 0 {
		ResponseError(c, http.StatusBadRequest, "statefulsets.containersEmpty")
		return
	}

	// 设置默认值
	if request.Replicas <= 0 {
		request.Replicas = 1
	}

	// 创建StatefulSet对象
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        request.Name,
			Namespace:   request.Namespace,
			Labels:      request.Labels,
			Annotations: request.Annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &request.Replicas,
			ServiceName: request.ServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: request.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      request.Labels,
					Annotations: request.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
		},
	}

	// 设置PodManagementPolicy
	if request.PodManagementPolicy != "" {
		sts.Spec.PodManagementPolicy = appsv1.PodManagementPolicyType(request.PodManagementPolicy)
	}

	// 设置UpdateStrategy
	if request.UpdateStrategy == "OnDelete" {
		sts.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.OnDeleteStatefulSetStrategyType,
		}
	} else {
		sts.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
				Partition: nil, // 默认为0
			},
		}
	}

	// 处理容器
	for _, containerInfo := range request.Containers {
		container, err := parseContainer(containerInfo)
		if err != nil {
			ResponseError(c, http.StatusBadRequest, "statefulsets.invalidContainer", err.Error())
			return
		}
		sts.Spec.Template.Spec.Containers = append(sts.Spec.Template.Spec.Containers, *container)
	}

	// 处理卷声明模板
	if len(request.VolumeClaimTemplates) > 0 {
		sts.Spec.VolumeClaimTemplates = make([]corev1.PersistentVolumeClaim, 0, len(request.VolumeClaimTemplates))
		for _, pvcTemplate := range request.VolumeClaimTemplates {
			name, ok := pvcTemplate["name"].(string)
			if !ok || name == "" {
				ResponseError(c, http.StatusBadRequest, "statefulsets.invalidPvcName")
				return
			}

			accessModes := []corev1.PersistentVolumeAccessMode{}
			if modes, ok := pvcTemplate["accessModes"].([]interface{}); ok {
				for _, mode := range modes {
					if modeStr, ok := mode.(string); ok {
						accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(modeStr))
					}
				}
			}
			if len(accessModes) == 0 {
				accessModes = append(accessModes, corev1.ReadWriteOnce)
			}

			storage := "1Gi" // 默认值
			if storageStr, ok := pvcTemplate["storage"].(string); ok && storageStr != "" {
				storage = storageStr
			}

			storageClassName := ""
			if scName, ok := pvcTemplate["storageClassName"].(string); ok {
				storageClassName = scName
			}

			pvc := corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: accessModes,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(storage),
						},
					},
				},
			}

			if storageClassName != "" {
				pvc.Spec.StorageClassName = &storageClassName
			}

			sts.Spec.VolumeClaimTemplates = append(sts.Spec.VolumeClaimTemplates, pvc)
		}
	}

	// 创建StatefulSet
	result, err := h.service.CreateStatefulSet(context.Background(), clusterName, sts)
	if err != nil {
		logger.Errorf("创建StatefulSet失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.createFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "StatefulSet created successfully",
		"statefulset": map[string]interface{}{
			"name":      result.Name,
			"namespace": result.Namespace,
		},
	})
}

// UpdateStatefulSet 更新StatefulSet
func (h *StatefulSetHandler) UpdateStatefulSet(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		ResponseError(c, http.StatusBadRequest, "statefulsets.invalidRequestFormat", err.Error())
		return
	}

	// 执行更新
	result, err := h.service.UpdateStatefulSet(context.Background(), clusterName, namespace, statefulsetName, updateData)
	if err != nil {
		logger.Errorf("更新StatefulSet失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.updateFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "StatefulSet updated successfully",
		"statefulset": map[string]interface{}{
			"name":      result.Name,
			"namespace": result.Namespace,
		},
	})
}

// DeleteStatefulSet 删除StatefulSet
func (h *StatefulSetHandler) DeleteStatefulSet(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	err := h.service.DeleteStatefulSet(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("删除StatefulSet失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.deleteFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "StatefulSet deleted successfully",
	})
}

// ScaleStatefulSet 扩缩容StatefulSet
func (h *StatefulSetHandler) ScaleStatefulSet(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	var request struct {
		Replicas int32 `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		ResponseError(c, http.StatusBadRequest, "statefulsets.invalidReplicas", err.Error())
		return
	}

	if request.Replicas < 0 {
		ResponseError(c, http.StatusBadRequest, "statefulsets.invalidReplicas", "replicas must be >= 0")
		return
	}

	result, err := h.service.ScaleStatefulSet(context.Background(), clusterName, namespace, statefulsetName, request.Replicas)
	if err != nil {
		logger.Errorf("扩缩容StatefulSet失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.scaleFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message":  "StatefulSet scaled successfully",
		"replicas": *result.Spec.Replicas,
	})
}

// RestartStatefulSet 重启StatefulSet
func (h *StatefulSetHandler) RestartStatefulSet(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	result, err := h.service.RestartStatefulSet(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("重启StatefulSet失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.restartFailed", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "StatefulSet restart initiated successfully",
		"statefulset": map[string]interface{}{
			"name":      result.Name,
			"namespace": result.Namespace,
		},
	})
}

// GetStatefulSetPods 获取StatefulSet相关的Pod
func (h *StatefulSetHandler) GetStatefulSetPods(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	pods, err := h.service.GetStatefulSetPods(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet相关Pod失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.getPodsFaileds", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// GetStatefulSetEvents 获取StatefulSet相关事件
func (h *StatefulSetHandler) GetStatefulSetEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	events, err := h.service.GetStatefulSetEvents(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet事件失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.getEventsFaileds", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": events,
	})
}

// GetAllStatefulSetEvents 获取StatefulSet及相关Pod的所有事件
func (h *StatefulSetHandler) GetAllStatefulSetEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	statefulsetName := c.Param("statefulset")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.namespaceEmpty")
		return
	}
	if statefulsetName == "" {
		ResponseError(c, http.StatusBadRequest, "statefulsets.nameEmpty")
		return
	}

	eventMap, err := h.service.GetAllStatefulSetEvents(context.Background(), clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet相关所有事件失败: %v", err)
		ResponseError(c, http.StatusInternalServerError, "statefulsets.getAllEventsFaileds", err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": eventMap,
	})
}

// parseContainer 解析容器配置信息
func parseContainer(containerInfo map[string]interface{}) (*corev1.Container, error) {
	name, ok := containerInfo["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("容器名称不能为空")
	}

	image, ok := containerInfo["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("容器镜像不能为空")
	}

	container := &corev1.Container{
		Name:  name,
		Image: image,
	}

	// 处理资源限制
	if resources, ok := containerInfo["resources"].(map[string]interface{}); ok {
		container.Resources = parseResourceRequirements(resources)
	}

	// 处理环境变量
	if envVars, ok := containerInfo["env"].([]interface{}); ok {
		container.Env = parseEnvVars(envVars)
	}

	// 处理端口
	if ports, ok := containerInfo["ports"].([]interface{}); ok {
		container.Ports = parsePorts(ports)
	}

	return container, nil
}

// parseResourceRequirements 解析资源需求
func parseResourceRequirements(resources map[string]interface{}) corev1.ResourceRequirements {
	result := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if requests, ok := resources["requests"].(map[string]interface{}); ok {
		for key, value := range requests {
			if strValue, ok := value.(string); ok && strValue != "" {
				result.Requests[corev1.ResourceName(key)] = resource.MustParse(strValue)
			}
		}
	}

	if limits, ok := resources["limits"].(map[string]interface{}); ok {
		for key, value := range limits {
			if strValue, ok := value.(string); ok && strValue != "" {
				result.Limits[corev1.ResourceName(key)] = resource.MustParse(strValue)
			}
		}
	}

	return result
}

// parseEnvVars 解析环境变量
func parseEnvVars(envVars []interface{}) []corev1.EnvVar {
	result := []corev1.EnvVar{}

	for _, item := range envVars {
		if env, ok := item.(map[string]interface{}); ok {
			name, nameOk := env["name"].(string)
			value, valueOk := env["value"].(string)

			if nameOk && name != "" {
				envVar := corev1.EnvVar{
					Name: name,
				}
				if valueOk {
					envVar.Value = value
				}
				result = append(result, envVar)
			}
		}
	}

	return result
}

// parsePorts 解析端口配置
func parsePorts(ports []interface{}) []corev1.ContainerPort {
	result := []corev1.ContainerPort{}

	for _, item := range ports {
		if port, ok := item.(map[string]interface{}); ok {
			containerPort, portOk := port["containerPort"].(float64)
			protocol, protocolOk := port["protocol"].(string)
			name, nameOk := port["name"].(string)

			if portOk && containerPort > 0 {
				p := corev1.ContainerPort{
					ContainerPort: int32(containerPort),
				}

				if nameOk && name != "" {
					p.Name = name
				}

				if protocolOk && protocol != "" {
					p.Protocol = corev1.Protocol(protocol)
				} else {
					p.Protocol = corev1.ProtocolTCP
				}

				result = append(result, p)
			}
		}
	}

	return result
}
