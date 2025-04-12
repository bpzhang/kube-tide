package k8s

import (
	"context"
	"fmt"
	"kube-tide/internal/utils/logger"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetService 提供与Kubernetes StatefulSets交互的服务
type StatefulSetService struct {
	clientManager *ClientManager
}

// NewStatefulSetService 创建一个新的StatefulSetService实例
func NewStatefulSetService(clientManager *ClientManager) *StatefulSetService {
	return &StatefulSetService{
		clientManager: clientManager,
	}
}

// StatefulSetInfo 包含StatefulSet的基本信息
type StatefulSetInfo struct {
	Name                 string            `json:"name"`
	Namespace            string            `json:"namespace"`
	Replicas             int32             `json:"replicas"`
	ReadyReplicas        int32             `json:"readyReplicas"`
	ServiceName          string            `json:"serviceName"`
	CreationTime         time.Time         `json:"creationTime"`
	Labels               map[string]string `json:"labels"`
	Selector             map[string]string `json:"selector"`
	ContainerCount       int               `json:"containerCount"`
	Images               []string          `json:"images"`
	UpdateStrategy       string            `json:"updateStrategy"`
	VolumeClaimTemplates []string          `json:"volumeClaimTemplates"`
}

// StatefulSetDetails 包含StatefulSet的详细信息
type StatefulSetDetails struct {
	StatefulSetInfo
	Annotations          map[string]string       `json:"annotations"`
	Containers           []ContainerInfo         `json:"containers"`
	Conditions           []StatefulSetCondition  `json:"conditions"`
	RevisionHistoryLimit *int32                  `json:"revisionHistoryLimit,omitempty"`
	PodManagementPolicy  string                  `json:"podManagementPolicy"`
	MinReadySeconds      int32                   `json:"minReadySeconds"`
	UpdateStrategy       string                  `json:"updateStrategy"`
	Paused               bool                    `json:"paused"`
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates"`
}

// StatefulSetCondition 表示StatefulSet的状态条件
type StatefulSetCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
}

// PersistentVolumeClaim 表示PVC模板信息
type PersistentVolumeClaim struct {
	Name             string            `json:"name"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	AccessModes      []string          `json:"accessModes"`
	Storage          string            `json:"storage"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// GetStatefulSets 获取指定命名空间的StatefulSet列表
func (s *StatefulSetService) GetStatefulSets(ctx context.Context, clusterName, namespace string) ([]StatefulSetInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	var statefulsetList *appsv1.StatefulSetList
	if namespace == "all" || namespace == "" {
		statefulsetList, err = client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	} else {
		statefulsetList, err = client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	var statefulsetInfos []StatefulSetInfo
	for _, sts := range statefulsetList.Items {
		containers := sts.Spec.Template.Spec.Containers
		images := make([]string, len(containers))
		for i, container := range containers {
			images[i] = container.Image
		}

		volClaims := make([]string, len(sts.Spec.VolumeClaimTemplates))
		for i, pvc := range sts.Spec.VolumeClaimTemplates {
			volClaims[i] = pvc.Name
		}

		stsUpdateStrategy := string(sts.Spec.UpdateStrategy.Type)

		statefulsetInfo := StatefulSetInfo{
			Name:                 sts.Name,
			Namespace:            sts.Namespace,
			Replicas:             *sts.Spec.Replicas,
			ReadyReplicas:        sts.Status.ReadyReplicas,
			ServiceName:          sts.Spec.ServiceName,
			CreationTime:         sts.CreationTimestamp.Time,
			Labels:               sts.Labels,
			Selector:             sts.Spec.Selector.MatchLabels,
			ContainerCount:       len(containers),
			Images:               images,
			UpdateStrategy:       string(sts.Spec.UpdateStrategy.Type),
			VolumeClaimTemplates: volClaims,
		}
		statefulsetInfos = append(statefulsetInfos, statefulsetInfo)
	}

	// 按创建时间排序，最新的在前面
	sort.Slice(statefulsetInfos, func(i, j int) bool {
		return statefulsetInfos[i].CreationTime.After(statefulsetInfos[j].CreationTime)
	})

	return statefulsetInfos, nil
}

// GetStatefulSetDetails 获取StatefulSet详情
func (s *StatefulSetService) GetStatefulSetDetails(ctx context.Context, clusterName, namespace, name string) (*StatefulSetDetails, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := sts.Spec.Template.Spec.Containers
	containerInfos := make([]ContainerInfo, len(containers))
	images := make([]string, len(containers))

	for i, container := range containers {
		resources := ResourceRequirements{}
		if container.Resources.Requests != nil {
			resources.Requests = make(map[string]string)
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				resources.Requests["cpu"] = cpu.String()
			}
			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				resources.Requests["memory"] = memory.String()
			}
		}
		if container.Resources.Limits != nil {
			resources.Limits = make(map[string]string)
			if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				resources.Limits["cpu"] = cpu.String()
			}
			if memory, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				resources.Limits["memory"] = memory.String()
			}
		}

		// 处理环境变量
		envVars := make([]EnvVar, len(container.Env))
		for j, env := range container.Env {
			envVar := EnvVar{
				Name:  env.Name,
				Value: env.Value,
			}
			if env.ValueFrom != nil {
				envVar.ValueFrom = &EnvVarSource{}
				if env.ValueFrom.ConfigMapKeyRef != nil {
					envVar.ValueFrom.ConfigMapKeyRef = &ConfigMapKeySelector{
						Name: env.ValueFrom.ConfigMapKeyRef.Name,
						Key:  env.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
				if env.ValueFrom.SecretKeyRef != nil {
					envVar.ValueFrom.SecretKeyRef = &SecretKeySelector{
						Name: env.ValueFrom.SecretKeyRef.Name,
						Key:  env.ValueFrom.SecretKeyRef.Key,
					}
				}
			}
			envVars[j] = envVar
		}

		// 处理端口
		ports := make([]ContainerPort, len(container.Ports))
		for j, port := range container.Ports {
			ports[j] = ContainerPort{
				Name:          port.Name,
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
			}
		}

		// 转换自定义类型的资源需求到K8s API类型
		k8sResources := convertResourceRequirementsToK8s(resources)
		// 转换自定义类型的环境变量到K8s API类型
		k8sEnvVars := convertEnvVarsToK8s(envVars)
		// 转换自定义类型的容器端口到K8s API类型
		k8sPorts := convertContainerPortsToK8s(ports)

		containerInfos[i] = ContainerInfo{
			Name:      container.Name,
			Image:     container.Image,
			Resources: k8sResources,
			Env:       k8sEnvVars,
			Ports:     k8sPorts,
		}
		images[i] = container.Image

		// 添加健康检查探针
		if container.LivenessProbe != nil {
			containerInfos[i].LivenessProbe = convertK8sProbeToCustomProbe(container.LivenessProbe)
		}
		if container.ReadinessProbe != nil {
			containerInfos[i].ReadinessProbe = convertK8sProbeToCustomProbe(container.ReadinessProbe)
		}
		if container.StartupProbe != nil {
			containerInfos[i].StartupProbe = convertK8sProbeToCustomProbe(container.StartupProbe)
		}
	}

	// 处理PVC模板
	pvcTemplates := make([]PersistentVolumeClaim, len(sts.Spec.VolumeClaimTemplates))
	for i, pvc := range sts.Spec.VolumeClaimTemplates {
		accessModes := make([]string, len(pvc.Spec.AccessModes))
		for j, mode := range pvc.Spec.AccessModes {
			accessModes[j] = string(mode)
		}

		storageClassName := ""
		if pvc.Spec.StorageClassName != nil {
			storageClassName = *pvc.Spec.StorageClassName
		}

		storage := ""
		if pvc.Spec.Resources.Requests != nil {
			if storageResource, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				storage = storageResource.String()
			}
		}

		pvcTemplates[i] = PersistentVolumeClaim{
			Name:             pvc.Name,
			StorageClassName: storageClassName,
			AccessModes:      accessModes,
			Storage:          storage,
			Labels:           pvc.Labels,
		}
	}

	// 处理条件
	conditions := make([]StatefulSetCondition, len(sts.Status.Conditions))
	for i, condition := range sts.Status.Conditions {
		conditions[i] = StatefulSetCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastTransitionTime: condition.LastTransitionTime.Format(time.RFC3339),
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
	}

	var updateStrategy string
	if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType {
		updateStrategy = "RollingUpdate"
	} else {
		updateStrategy = "OnDelete"
	}

	paused := false
	if sts.Spec.UpdateStrategy.RollingUpdate != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		// 如果partition等于副本数，视为暂停
		paused = (*sts.Spec.UpdateStrategy.RollingUpdate.Partition == *sts.Spec.Replicas)
	}

	details := &StatefulSetDetails{
		StatefulSetInfo: StatefulSetInfo{
			Name:                 sts.Name,
			Namespace:            sts.Namespace,
			Replicas:             *sts.Spec.Replicas,
			ReadyReplicas:        sts.Status.ReadyReplicas,
			ServiceName:          sts.Spec.ServiceName,
			CreationTime:         sts.CreationTimestamp.Time,
			Labels:               sts.Labels,
			Selector:             sts.Spec.Selector.MatchLabels,
			ContainerCount:       len(containers),
			Images:               images,
			UpdateStrategy:       string(sts.Spec.UpdateStrategy.Type),
			VolumeClaimTemplates: make([]string, len(sts.Spec.VolumeClaimTemplates)),
		},
		Annotations:          sts.Annotations,
		Containers:           containerInfos,
		Conditions:           conditions,
		PodManagementPolicy:  string(sts.Spec.PodManagementPolicy),
		MinReadySeconds:      sts.Spec.MinReadySeconds,
		UpdateStrategy:       updateStrategy,
		Paused:               paused,
		VolumeClaimTemplates: pvcTemplates,
	}

	if sts.Spec.RevisionHistoryLimit != nil {
		details.RevisionHistoryLimit = sts.Spec.RevisionHistoryLimit
	}

	// 处理VolumeClaimTemplates
	for i, pvc := range sts.Spec.VolumeClaimTemplates {
		details.StatefulSetInfo.VolumeClaimTemplates[i] = pvc.Name
	}

	return details, nil
}

// CreateStatefulSet 创建新的StatefulSet
func (s *StatefulSetService) CreateStatefulSet(ctx context.Context, clusterName string, statefulset *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().StatefulSets(statefulset.Namespace).Create(ctx, statefulset, metav1.CreateOptions{})
}

// UpdateStatefulSet 更新StatefulSet
func (s *StatefulSetService) UpdateStatefulSet(ctx context.Context, clusterName, namespace, name string, updateData map[string]interface{}) (*appsv1.StatefulSet, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取现有StatefulSet
	currentSts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 应用更新
	if replicas, ok := updateData["replicas"].(int32); ok {
		currentSts.Spec.Replicas = &replicas
	}

	// 更新容器镜像
	if images, ok := updateData["image"].(map[string]string); ok {
		for i, container := range currentSts.Spec.Template.Spec.Containers {
			if newImage, ok := images[container.Name]; ok {
				currentSts.Spec.Template.Spec.Containers[i].Image = newImage
			}
		}
	}

	// 更新资源请求/限制
	if resources, ok := updateData["resources"].(map[string]map[string]map[string]string); ok {
		for i, container := range currentSts.Spec.Template.Spec.Containers {
			if containerResources, ok := resources[container.Name]; ok {
				if requests, ok := containerResources["requests"]; ok {
					if currentSts.Spec.Template.Spec.Containers[i].Resources.Requests == nil {
						currentSts.Spec.Template.Spec.Containers[i].Resources.Requests = corev1.ResourceList{}
					}
					for resource, value := range requests {
						parsedQuantity, err := corev1.ParseQuantity(value)
						if err != nil {
							return nil, fmt.Errorf("解析资源请求失败: %w", err)
						}
						currentSts.Spec.Template.Spec.Containers[i].Resources.Requests[corev1.ResourceName(resource)] = parsedQuantity
					}
				}
				if limits, ok := containerResources["limits"]; ok {
					if currentSts.Spec.Template.Spec.Containers[i].Resources.Limits == nil {
						currentSts.Spec.Template.Spec.Containers[i].Resources.Limits = corev1.ResourceList{}
					}
					for resource, value := range limits {
						parsedQuantity, err := corev1.ParseQuantity(value)
						if err != nil {
							return nil, fmt.Errorf("解析资源限制失败: %w", err)
						}
						currentSts.Spec.Template.Spec.Containers[i].Resources.Limits[corev1.ResourceName(resource)] = parsedQuantity
					}
				}
			}
		}
	}

	// 更新环境变量
	if envs, ok := updateData["env"].(map[string][]map[string]interface{}); ok {
		for containerName, envVars := range envs {
			for i, container := range currentSts.Spec.Template.Spec.Containers {
				if container.Name == containerName {
					newEnvVars := make([]corev1.EnvVar, len(envVars))
					for j, env := range envVars {
						envVar := corev1.EnvVar{
							Name:  env["name"].(string),
							Value: env["value"].(string),
						}
						if valueFrom, ok := env["valueFrom"]; ok {
							valueFromMap := valueFrom.(map[string]interface{})
							envVar.ValueFrom = &corev1.EnvVarSource{}
							if configMapRef, ok := valueFromMap["configMapKeyRef"]; ok {
								configMapRefMap := configMapRef.(map[string]string)
								envVar.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapRefMap["name"],
									},
									Key: configMapRefMap["key"],
								}
							} else if secretRef, ok := valueFromMap["secretKeyRef"]; ok {
								secretRefMap := secretRef.(map[string]string)
								envVar.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretRefMap["name"],
									},
									Key: secretRefMap["key"],
								}
							}
						}
						newEnvVars[j] = envVar
					}
					currentSts.Spec.Template.Spec.Containers[i].Env = newEnvVars
					break
				}
			}
		}
	}

	// 更新标签
	if labels, ok := updateData["labels"].(map[string]string); ok {
		if currentSts.Labels == nil {
			currentSts.Labels = make(map[string]string)
		}
		for k, v := range labels {
			currentSts.Labels[k] = v
		}
		// 同时更新Pod模板标签
		if currentSts.Spec.Template.Labels == nil {
			currentSts.Spec.Template.Labels = make(map[string]string)
		}
		for k, v := range labels {
			currentSts.Spec.Template.Labels[k] = v
		}
	}

	// 更新注解
	if annotations, ok := updateData["annotations"].(map[string]string); ok {
		if currentSts.Annotations == nil {
			currentSts.Annotations = make(map[string]string)
		}
		for k, v := range annotations {
			currentSts.Annotations[k] = v
		}
	}

	// 更新策略
	if strategy, ok := updateData["strategy"].(map[string]interface{}); ok {
		if strategyType, ok := strategy["type"].(string); ok {
			currentSts.Spec.UpdateStrategy.Type = appsv1.StatefulSetUpdateStrategyType(strategyType)
			if strategyType == "RollingUpdate" && strategy["rollingUpdate"] != nil {
				rollingUpdate := strategy["rollingUpdate"].(map[string]interface{})
				if currentSts.Spec.UpdateStrategy.RollingUpdate == nil {
					currentSts.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{}
				}
				if partition, ok := rollingUpdate["partition"].(int32); ok {
					currentSts.Spec.UpdateStrategy.RollingUpdate.Partition = &partition
				}
			}
		}
	}

	// 更新暂停状态
	if paused, ok := updateData["paused"].(bool); ok {
		if paused {
			// 如果要暂停，设置partition为replicas数量
			if currentSts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType {
				if currentSts.Spec.UpdateStrategy.RollingUpdate == nil {
					currentSts.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{}
				}
				replicas := *currentSts.Spec.Replicas
				currentSts.Spec.UpdateStrategy.RollingUpdate.Partition = &replicas
			}
		} else {
			// 如果要恢复，设置partition为0
			if currentSts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType {
				if currentSts.Spec.UpdateStrategy.RollingUpdate == nil {
					currentSts.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{}
				}
				var zero int32 = 0
				currentSts.Spec.UpdateStrategy.RollingUpdate.Partition = &zero
			}
		}
	}

	// 更新其他字段...

	// 执行更新
	return client.AppsV1().StatefulSets(namespace).Update(ctx, currentSts, metav1.UpdateOptions{})
}

// DeleteStatefulSet 删除StatefulSet
func (s *StatefulSetService) DeleteStatefulSet(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	return client.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ScaleStatefulSet 扩缩容StatefulSet
func (s *StatefulSetService) ScaleStatefulSet(ctx context.Context, clusterName, namespace, name string, replicas int32) (*appsv1.StatefulSet, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取现有StatefulSet
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 更新副本数
	sts.Spec.Replicas = &replicas

	// 应用更新
	return client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
}

// RestartStatefulSet 重启StatefulSet（通过添加重启注解）
func (s *StatefulSetService) RestartStatefulSet(ctx context.Context, clusterName, namespace, name string) (*appsv1.StatefulSet, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取现有StatefulSet
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 添加重启注解
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 应用更新
	return client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
}

// GetStatefulSetPods 获取StatefulSet关联的所有Pod
func (s *StatefulSetService) GetStatefulSetPods(ctx context.Context, clusterName, namespace, statefulsetName string) ([]corev1.Pod, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取StatefulSet
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 使用标签选择器查找关联的Pod
	labelSelector := metav1.FormatLabelSelector(sts.Spec.Selector)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// GetStatefulSetEvents 获取与StatefulSet相关的事件
func (s *StatefulSetService) GetStatefulSetEvents(ctx context.Context, clusterName, namespace, statefulsetName string) ([]corev1.Event, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=StatefulSet", statefulsetName, namespace)
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	// 按时间降序排序
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})

	return events.Items, nil
}

// GetAllStatefulSetEvents 获取StatefulSet及其关联的Pod的所有事件
func (s *StatefulSetService) GetAllStatefulSetEvents(ctx context.Context, clusterName, namespace, statefulsetName string) (map[string][]corev1.Event, error) {
	result := map[string][]corev1.Event{
		"statefulset": {},
		"pod":         {},
	}

	// 获取StatefulSet事件
	stsEvents, err := s.GetStatefulSetEvents(ctx, clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet事件失败: %v", err)
	} else {
		result["statefulset"] = stsEvents
	}

	// 获取关联的Pod
	pods, err := s.GetStatefulSetPods(ctx, clusterName, namespace, statefulsetName)
	if err != nil {
		logger.Errorf("获取StatefulSet的Pod失败: %v", err)
		return result, nil // 返回已获取的事件，不中断
	}

	// 获取每个Pod的事件
	var podEvents []corev1.Event
	k8sClient, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		logger.Errorf("获取K8s客户端失败: %v", err)
		return result, nil
	}

	for _, pod := range pods {
		fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", pod.Name, namespace)
		events, err := k8sClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		if err != nil {
			logger.Errorf("获取Pod %s 事件失败: %v", pod.Name, err)
			continue
		}
		podEvents = append(podEvents, events.Items...)
	}

	// 按时间降序排序
	sort.Slice(podEvents, func(i, j int) bool {
		return podEvents[i].LastTimestamp.After(podEvents[j].LastTimestamp.Time)
	})

	result["pod"] = podEvents
	return result, nil
}
