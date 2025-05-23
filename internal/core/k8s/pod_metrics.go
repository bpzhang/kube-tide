package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodMetrics Pod指标结构
type PodMetrics struct {
	// 当前CPU使用率（百分比）
	CPUUsage float64 `json:"cpuUsage"`
	// 当前内存使用率（百分比）
	MemoryUsage float64 `json:"memoryUsage"`
	// 当前硬盘使用率（百分比）
	DiskUsage float64 `json:"diskUsage"`
	// CPU请求值（单位：m）
	CPURequests string `json:"cpuRequests"`
	// CPU限制值（单位：m）
	CPULimits string `json:"cpuLimits"`
	// 内存请求值（例如：100Mi）
	MemoryRequests string `json:"memoryRequests"`
	// 内存限制值（例如：200Mi）
	MemoryLimits string `json:"memoryLimits"`
	// 硬盘存储请求值（例如：1Gi）
	DiskRequests string `json:"diskRequests"`
	// 硬盘存储限制值（例如：10Gi）
	DiskLimits string `json:"diskLimits"`
	// 历史数据（24小时内每小时一个数据点）
	HistoricalData struct {
		CPUUsage    []MetricDataPoint `json:"cpuUsage"`
		MemoryUsage []MetricDataPoint `json:"memoryUsage"`
		DiskUsage   []MetricDataPoint `json:"diskUsage"`
	} `json:"historicalData"`
	// 容器指标
	Containers []ContainerMetrics `json:"containers"`
}

// ContainerMetrics 容器指标结构
type ContainerMetrics struct {
	Name           string  `json:"name"`
	CPUUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	DiskUsage      float64 `json:"diskUsage"`
	CPURequests    string  `json:"cpuRequests"`
	CPULimits      string  `json:"cpuLimits"`
	MemoryRequests string  `json:"memoryRequests"`
	MemoryLimits   string  `json:"memoryLimits"`
	DiskRequests   string  `json:"diskRequests"`
	DiskLimits     string  `json:"diskLimits"`
	// 容器级别的历史数据
	HistoricalData struct {
		CPUUsage    []MetricDataPoint `json:"cpuUsage"`
		MemoryUsage []MetricDataPoint `json:"memoryUsage"`
		DiskUsage   []MetricDataPoint `json:"diskUsage"`
	} `json:"historicalData"`
}

// GetPodMetrics 获取Pod监控指标
func GetPodMetrics(client *kubernetes.Clientset, namespace, podName string) (*PodMetrics, error) {
	ctx := context.Background()
	metrics := &PodMetrics{
		HistoricalData: struct {
			CPUUsage    []MetricDataPoint `json:"cpuUsage"`
			MemoryUsage []MetricDataPoint `json:"memoryUsage"`
			DiskUsage   []MetricDataPoint `json:"diskUsage"`
		}{
			CPUUsage:    make([]MetricDataPoint, 0),
			MemoryUsage: make([]MetricDataPoint, 0),
			DiskUsage:   make([]MetricDataPoint, 0),
		},
		Containers: make([]ContainerMetrics, 0),
	}

	// 获取Pod详情
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod详情失败: %v", err)
	}

	// 创建metrics客户端
	// 使用当前集群的配置
	config, err := getCurrentClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("获取集群配置失败: %v", err)
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建metrics客户端失败: %v", err)
	}

	// 获取Pod metrics
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod metrics失败: %v", err)
	}

	// 计算CPU和内存使用率
	// 如果没有资源请求或限制，则使用节点总量作为参考
	nodeCapacity := make(map[string]map[string]int64)

	// 获取Pod所在节点的资源容量
	if pod.Spec.NodeName != "" {
		node, err := client.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err == nil {
			capacity := node.Status.Capacity
			nodeCapacity["cpu"] = map[string]int64{"capacity": capacity.Cpu().MilliValue()}
			nodeCapacity["memory"] = map[string]int64{"capacity": capacity.Memory().Value()}

			// 添加磁盘容量信息
			for _, condition := range node.Status.Conditions {
				if condition.Type == "DiskPressure" && condition.Status == "False" {
					// 如果节点没有磁盘压力，假设有充足的磁盘空间
					// 在实际环境中，应该通过metrics-server或其他方式获取真实的磁盘信息
					nodeCapacity["disk"] = map[string]int64{"capacity": 1000 * 1024 * 1024 * 1024} // 假设1TB
				}
			}
		}
	}

	// 计算Pod的总CPU和内存请求/限制
	totalCPURequests := int64(0)
	totalCPULimits := int64(0)
	totalMemoryRequests := int64(0)
	totalMemoryLimits := int64(0)
	totalDiskRequests := int64(0)
	totalDiskLimits := int64(0)

	for _, container := range pod.Spec.Containers {
		cpuRequests := container.Resources.Requests.Cpu().MilliValue()
		cpuLimits := container.Resources.Limits.Cpu().MilliValue()
		memoryRequests := container.Resources.Requests.Memory().Value()
		memoryLimits := container.Resources.Limits.Memory().Value()

		totalCPURequests += cpuRequests
		totalCPULimits += cpuLimits
		totalMemoryRequests += memoryRequests
		totalMemoryLimits += memoryLimits
	}

	// 获取PVC的磁盘请求量
	pvcs := make(map[string]int64)
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName
			pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
			if err == nil {
				if storage, ok := pvc.Spec.Resources.Requests["storage"]; ok {
					storageValue := storage.Value()
					pvcs[pvcName] = storageValue
					totalDiskRequests += storageValue
				}
			}
		}
	}

	// 如果有磁盘限制策略（如StorageClass的限制），这里可以获取并计算totalDiskLimits
	// 简化处理：如果没有明确的限制，假设限制等于请求的1.2倍
	if totalDiskLimits == 0 && totalDiskRequests > 0 {
		totalDiskLimits = int64(float64(totalDiskRequests) * 1.2)
	}

	metrics.CPURequests = formatCPU(totalCPURequests)
	metrics.CPULimits = formatCPU(totalCPULimits)
	metrics.MemoryRequests = formatMemory(totalMemoryRequests)
	metrics.MemoryLimits = formatMemory(totalMemoryLimits)
	metrics.DiskRequests = FormatStorage(totalDiskRequests)
	metrics.DiskLimits = FormatStorage(totalDiskLimits)

	// 计算容器指标
	totalCPUUsage := int64(0)
	totalMemoryUsage := int64(0)

	// 获取真实的CPU和内存使用情况
	resourceUsage, err := GetPodResourceUsage(client, config, namespace, podName)
	if err == nil && resourceUsage != nil {
		// 使用真实的资源使用数据
		totalCPUUsage = int64(resourceUsage.TotalCPUUsage * 1000) // 转换为毫核
		totalMemoryUsage = resourceUsage.TotalMemoryUsage

		// 使用容器级别的真实数据
		for _, containerMetrics := range resourceUsage.Containers {
			// 找到容器的资源请求和限制
			var cpuRequests, cpuLimits int64
			var memoryRequests, memoryLimits int64

			for _, container := range pod.Spec.Containers {
				if container.Name == containerMetrics.Name {
					cpuRequests = container.Resources.Requests.Cpu().MilliValue()
					cpuLimits = container.Resources.Limits.Cpu().MilliValue()
					memoryRequests = container.Resources.Requests.Memory().Value()
					memoryLimits = container.Resources.Limits.Memory().Value()
					break
				}
			}

			// 计算容器CPU使用率
			cpuUsage := int64(containerMetrics.CPUUsage * 1000) // 转换为毫核
			cpuUsagePercentage := float64(0)
			if cpuLimits > 0 {
				cpuUsagePercentage = float64(cpuUsage) / float64(cpuLimits) * 100
			} else if cpuRequests > 0 {
				cpuUsagePercentage = float64(cpuUsage) / float64(cpuRequests) * 100
			} else if capacity, ok := nodeCapacity["cpu"]; ok {
				cpuUsagePercentage = float64(cpuUsage) / float64(capacity["capacity"]) * 100
			}

			// 计算容器内存使用率
			memoryUsagePercentage := float64(0)
			if memoryLimits > 0 {
				memoryUsagePercentage = float64(containerMetrics.MemoryUsage) / float64(memoryLimits) * 100
			} else if memoryRequests > 0 {
				memoryUsagePercentage = float64(containerMetrics.MemoryUsage) / float64(memoryRequests) * 100
			} else if capacity, ok := nodeCapacity["memory"]; ok {
				memoryUsagePercentage = float64(containerMetrics.MemoryUsage) / float64(capacity["capacity"]) * 100
			} // 容器的磁盘使用率
			// 现在使用真实数据而非模拟数据
			diskUsagePercentage := float64(0)
			diskRequests := int64(0)
			diskLimits := int64(0)

			// 计算容器使用的存储资源
			// 根据容器资源比例分配磁盘使用
			if totalMemoryRequests > 0 && memoryRequests > 0 && totalDiskRequests > 0 {
				ratio := float64(memoryRequests) / float64(totalMemoryRequests)
				diskRequests = int64(ratio * float64(totalDiskRequests))
				if totalDiskLimits > 0 {
					diskLimits = int64(ratio * float64(totalDiskLimits))
				}
			}

			// 获取单个容器的磁盘使用率
			containerDiskUsage := int64(0)

			// 先检查我们是否已经获取了Pod的磁盘使用情况
			podDiskUsage, err := GetPodDiskUsage(client, config, namespace, podName)
			if err == nil && len(podDiskUsage) > 0 {
				// 查找与此容器相关的卷挂载
				var containerVolumes []string
				for _, container := range pod.Spec.Containers {
					if container.Name == containerMetrics.Name {
						for _, volumeMount := range container.VolumeMounts {
							containerVolumes = append(containerVolumes, volumeMount.MountPath)
						}
						break
					}
				}

				// 累加此容器使用的磁盘空间
				for path, usage := range podDiskUsage {
					for _, volumePath := range containerVolumes {
						if strings.HasPrefix(path, volumePath) {
							containerDiskUsage += usage
							break
						}
					}
				}

				// 如果未找到与容器关联的卷，按内存使用比例分配总磁盘使用量
				if containerDiskUsage == 0 {
					totalUsedBytes := int64(0)
					for _, usage := range podDiskUsage {
						totalUsedBytes += usage
					}

					// 按内存使用比例分配
					if totalMemoryUsage > 0 && containerMetrics.MemoryUsage > 0 {
						ratio := float64(containerMetrics.MemoryUsage) / float64(totalMemoryUsage)
						containerDiskUsage = int64(ratio * float64(totalUsedBytes))
					}
				}

				// 计算使用率
				if diskLimits > 0 {
					diskUsagePercentage = float64(containerDiskUsage) / float64(diskLimits) * 100
				} else if diskRequests > 0 {
					diskUsagePercentage = float64(containerDiskUsage) / float64(diskRequests) * 100
				} else if _, ok := nodeCapacity["disk"]; ok {
					diskUsagePercentage = float64(containerDiskUsage) / float64(nodeCapacity["disk"]["capacity"]) * 100
				}
			} else {
				// 如果无法获取真实磁盘使用情况，回退到基于内存使用率的估算
				if diskLimits > 0 {
					diskUsagePercentage = memoryUsagePercentage * 0.8
				} else if diskRequests > 0 {
					diskUsagePercentage = memoryUsagePercentage * 0.9
				} else if _, ok := nodeCapacity["disk"]; ok {
					diskUsagePercentage = memoryUsagePercentage * 0.5
				}
			}

			// 确保使用率不超过100%
			if diskUsagePercentage > 100 {
				diskUsagePercentage = 100
			} // 初始化容器的历史数据结构
			containerHistorical := struct {
				CPUUsage    []MetricDataPoint `json:"cpuUsage"`
				MemoryUsage []MetricDataPoint `json:"memoryUsage"`
				DiskUsage   []MetricDataPoint `json:"diskUsage"`
			}{
				CPUUsage:    make([]MetricDataPoint, 0),
				MemoryUsage: make([]MetricDataPoint, 0),
				DiskUsage:   make([]MetricDataPoint, 0),
			}

			// 添加当前的资源使用数据作为历史数据的起点
			currentTime := time.Now().Format(time.RFC3339)
			containerHistorical.CPUUsage = append(containerHistorical.CPUUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     cpuUsagePercentage,
			})

			containerHistorical.MemoryUsage = append(containerHistorical.MemoryUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     memoryUsagePercentage,
			})

			containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     diskUsagePercentage,
			})

			// 从缓存中获取容器的历史数据
			if containerCache, err := GetPodResourceUsage(client, config, namespace, podName); err == nil && containerCache != nil {
				// 从资源使用缓存中查找容器的历史数据
				for _, container := range containerCache.Containers {
					if container.Name == containerMetrics.Name {
						// 获取容器的CPU历史数据
						if cpuData, ok := containerCache.Historical["cpu"]; ok && len(cpuData) > 0 {
							for _, point := range cpuData {
								// 计算CPU使用率百分比
								cpuPercentage := float64(0)
								if cpuLimits > 0 {
									cpuPercentage = point.Value * 1000 / float64(cpuLimits) * 100
								} else if cpuRequests > 0 {
									cpuPercentage = point.Value * 1000 / float64(cpuRequests) * 100
								} else if capacity, ok := nodeCapacity["cpu"]; ok {
									cpuPercentage = point.Value * 1000 / float64(capacity["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									cpuPercentage = cpuUsagePercentage
								}

								// 确保值在有效范围内
								if cpuPercentage < 0 {
									cpuPercentage = 0
								} else if cpuPercentage > 100 {
									cpuPercentage = 100
								}

								containerHistorical.CPUUsage = append(containerHistorical.CPUUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     cpuPercentage,
								})
							}
						}

						// 获取容器的内存历史数据
						if memoryData, ok := containerCache.Historical["memory"]; ok && len(memoryData) > 0 {
							for _, point := range memoryData {
								// 计算内存使用率百分比
								memoryPercentage := float64(0)
								if memoryLimits > 0 {
									memoryPercentage = point.Value / float64(memoryLimits) * 100
								} else if memoryRequests > 0 {
									memoryPercentage = point.Value / float64(memoryRequests) * 100
								} else if capacity, ok := nodeCapacity["memory"]; ok {
									memoryPercentage = point.Value / float64(capacity["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									memoryPercentage = memoryUsagePercentage
								}

								// 确保值在有效范围内
								if memoryPercentage < 0 {
									memoryPercentage = 0
								} else if memoryPercentage > 100 {
									memoryPercentage = 100
								}

								containerHistorical.MemoryUsage = append(containerHistorical.MemoryUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     memoryPercentage,
								})
							}
						}

						// 获取容器的磁盘历史数据
						if diskData, ok := containerCache.Historical["disk"]; ok && len(diskData) > 0 {
							for _, point := range diskData {
								// 计算磁盘使用率百分比
								diskPercentage := float64(0)
								if diskLimits > 0 {
									diskPercentage = point.Value / float64(diskLimits) * 100
								} else if diskRequests > 0 {
									diskPercentage = point.Value / float64(diskRequests) * 100
								} else if _, ok := nodeCapacity["disk"]; ok {
									diskPercentage = point.Value / float64(nodeCapacity["disk"]["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									diskPercentage = diskUsagePercentage
								}

								// 确保值在有效范围内
								if diskPercentage < 0 {
									diskPercentage = 0
								} else if diskPercentage > 100 {
									diskPercentage = 100
								}

								containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     diskPercentage,
								})
							}
						}
						break
					}
				}
			} // 添加到容器指标列表
			metrics.Containers = append(metrics.Containers, ContainerMetrics{
				Name:           containerMetrics.Name,
				CPUUsage:       cpuUsagePercentage,
				MemoryUsage:    memoryUsagePercentage,
				DiskUsage:      diskUsagePercentage,
				CPURequests:    formatCPU(cpuRequests),
				CPULimits:      formatCPU(cpuLimits),
				MemoryRequests: formatMemory(memoryRequests),
				MemoryLimits:   formatMemory(memoryLimits),
				DiskRequests:   FormatStorage(diskRequests),
				DiskLimits:     FormatStorage(diskLimits),
				HistoricalData: containerHistorical,
			})
		}
	} else {
		// 如果无法获取真实资源使用情况，回退到使用metrics API的数据
		for _, containerMetrics := range podMetrics.Containers {
			cpuUsage := containerMetrics.Usage.Cpu().MilliValue()
			memoryUsage := containerMetrics.Usage.Memory().Value()

			totalCPUUsage += cpuUsage
			totalMemoryUsage += memoryUsage

			// 找到容器的资源请求和限制
			var cpuRequests, cpuLimits int64
			var memoryRequests, memoryLimits int64

			for _, container := range pod.Spec.Containers {
				if container.Name == containerMetrics.Name {
					cpuRequests = container.Resources.Requests.Cpu().MilliValue()
					cpuLimits = container.Resources.Limits.Cpu().MilliValue()
					memoryRequests = container.Resources.Requests.Memory().Value()
					memoryLimits = container.Resources.Limits.Memory().Value()
					break
				}
			}

			// 计算容器CPU使用率
			cpuUsagePercentage := float64(0)
			if cpuLimits > 0 {
				cpuUsagePercentage = float64(cpuUsage) / float64(cpuLimits) * 100
			} else if cpuRequests > 0 {
				cpuUsagePercentage = float64(cpuUsage) / float64(cpuRequests) * 100
			} else if capacity, ok := nodeCapacity["cpu"]; ok {
				cpuUsagePercentage = float64(cpuUsage) / float64(capacity["capacity"]) * 100
			}

			// 计算容器内存使用率
			memoryUsagePercentage := float64(0)
			if memoryLimits > 0 {
				memoryUsagePercentage = float64(memoryUsage) / float64(memoryLimits) * 100
			} else if memoryRequests > 0 {
				memoryUsagePercentage = float64(memoryUsage) / float64(memoryRequests) * 100
			} else if capacity, ok := nodeCapacity["memory"]; ok {
				memoryUsagePercentage = float64(memoryUsage) / float64(capacity["capacity"]) * 100
			}

			// 容器的磁盘使用率（简化处理：平均分配总体磁盘使用率）
			diskUsagePercentage := float64(0)
			diskRequests := int64(0)
			diskLimits := int64(0)

			// 简化处理：根据容器资源比例分配磁盘使用
			if totalMemoryRequests > 0 && memoryRequests > 0 && totalDiskRequests > 0 {
				ratio := float64(memoryRequests) / float64(totalMemoryRequests)
				diskRequests = int64(ratio * float64(totalDiskRequests))
				if totalDiskLimits > 0 {
					diskLimits = int64(ratio * float64(totalDiskLimits))
				}
			}

			// 回退到基于内存使用率的磁盘使用估算
			if diskLimits > 0 {
				diskUsagePercentage = memoryUsagePercentage * 0.8
			} else if diskRequests > 0 {
				diskUsagePercentage = memoryUsagePercentage * 0.9
			} else if _, ok := nodeCapacity["disk"]; ok {
				diskUsagePercentage = memoryUsagePercentage * 0.5
			}

			// 确保使用率不超过100%
			if diskUsagePercentage > 100 {
				diskUsagePercentage = 100
			}

			// 初始化容器的历史数据结构
			containerHistorical := struct {
				CPUUsage    []MetricDataPoint `json:"cpuUsage"`
				MemoryUsage []MetricDataPoint `json:"memoryUsage"`
				DiskUsage   []MetricDataPoint `json:"diskUsage"`
			}{
				CPUUsage:    make([]MetricDataPoint, 0),
				MemoryUsage: make([]MetricDataPoint, 0),
				DiskUsage:   make([]MetricDataPoint, 0),
			}

			// 添加当前的资源使用数据作为历史数据的起点
			currentTime := time.Now().Format(time.RFC3339)
			containerHistorical.CPUUsage = append(containerHistorical.CPUUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     cpuUsagePercentage,
			})

			containerHistorical.MemoryUsage = append(containerHistorical.MemoryUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     memoryUsagePercentage,
			})

			containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
				Timestamp: currentTime,
				Value:     diskUsagePercentage,
			})

			// 从缓存中获取容器的历史数据
			if containerCache, err := GetPodResourceUsage(client, config, namespace, podName); err == nil && containerCache != nil {
				// 从资源使用缓存中查找容器的历史数据
				for _, container := range containerCache.Containers {
					if container.Name == containerMetrics.Name {
						// 获取容器的CPU历史数据
						if cpuData, ok := containerCache.Historical["cpu"]; ok && len(cpuData) > 0 {
							for _, point := range cpuData {
								// 计算CPU使用率百分比
								cpuPercentage := float64(0)
								if cpuLimits > 0 {
									cpuPercentage = point.Value * 1000 / float64(cpuLimits) * 100
								} else if cpuRequests > 0 {
									cpuPercentage = point.Value * 1000 / float64(cpuRequests) * 100
								} else if capacity, ok := nodeCapacity["cpu"]; ok {
									cpuPercentage = point.Value * 1000 / float64(capacity["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									cpuPercentage = cpuUsagePercentage
								}

								// 确保值在有效范围内
								if cpuPercentage < 0 {
									cpuPercentage = 0
								} else if cpuPercentage > 100 {
									cpuPercentage = 100
								}

								containerHistorical.CPUUsage = append(containerHistorical.CPUUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     cpuPercentage,
								})
							}
						}

						// 获取容器的内存历史数据
						if memoryData, ok := containerCache.Historical["memory"]; ok && len(memoryData) > 0 {
							for _, point := range memoryData {
								// 计算内存使用率百分比
								memoryPercentage := float64(0)
								if memoryLimits > 0 {
									memoryPercentage = point.Value / float64(memoryLimits) * 100
								} else if memoryRequests > 0 {
									memoryPercentage = point.Value / float64(memoryRequests) * 100
								} else if capacity, ok := nodeCapacity["memory"]; ok {
									memoryPercentage = point.Value / float64(capacity["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									memoryPercentage = memoryUsagePercentage
								}

								// 确保值在有效范围内
								if memoryPercentage < 0 {
									memoryPercentage = 0
								} else if memoryPercentage > 100 {
									memoryPercentage = 100
								}

								containerHistorical.MemoryUsage = append(containerHistorical.MemoryUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     memoryPercentage,
								})
							}
						}

						// 获取容器的磁盘历史数据
						if diskData, ok := containerCache.Historical["disk"]; ok && len(diskData) > 0 {
							for _, point := range diskData {
								// 计算磁盘使用率百分比
								diskPercentage := float64(0)
								if diskLimits > 0 {
									diskPercentage = point.Value / float64(diskLimits) * 100
								} else if diskRequests > 0 {
									diskPercentage = point.Value / float64(diskRequests) * 100
								} else if _, ok := nodeCapacity["disk"]; ok {
									diskPercentage = point.Value / float64(nodeCapacity["disk"]["capacity"]) * 100
								} else {
									// 如果没有限制或请求，使用当前实际值
									diskPercentage = diskUsagePercentage
								}

								// 确保值在有效范围内
								if diskPercentage < 0 {
									diskPercentage = 0
								} else if diskPercentage > 100 {
									diskPercentage = 100
								}

								containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     diskPercentage,
								})
							}
						}
						break
					}
				}
			}

			// 添加到容器指标列表
			metrics.Containers = append(metrics.Containers, ContainerMetrics{
				Name:           containerMetrics.Name,
				CPUUsage:       cpuUsagePercentage,
				MemoryUsage:    memoryUsagePercentage,
				DiskUsage:      diskUsagePercentage,
				CPURequests:    formatCPU(cpuRequests),
				CPULimits:      formatCPU(cpuLimits),
				MemoryRequests: formatMemory(memoryRequests),
				MemoryLimits:   formatMemory(memoryLimits),
				DiskRequests:   FormatStorage(diskRequests),
				DiskLimits:     FormatStorage(diskLimits),
				HistoricalData: containerHistorical,
			})
		}
	}

	// 计算Pod总体CPU使用率
	if totalCPULimits > 0 {
		metrics.CPUUsage = float64(totalCPUUsage) / float64(totalCPULimits) * 100
	} else if totalCPURequests > 0 {
		metrics.CPUUsage = float64(totalCPUUsage) / float64(totalCPURequests) * 100
	} else if capacity, ok := nodeCapacity["cpu"]; ok {
		metrics.CPUUsage = float64(totalCPUUsage) / float64(capacity["capacity"]) * 100
	}

	// 计算Pod总体内存使用率
	if totalMemoryLimits > 0 {
		metrics.MemoryUsage = float64(totalMemoryUsage) / float64(totalMemoryLimits) * 100
	} else if totalMemoryRequests > 0 {
		metrics.MemoryUsage = float64(totalMemoryUsage) / float64(totalMemoryRequests) * 100
	} else if capacity, ok := nodeCapacity["memory"]; ok {
		metrics.MemoryUsage = float64(totalMemoryUsage) / float64(capacity["capacity"]) * 100
	}

	// 计算Pod总体硬盘使用率 - 通过直接在Pod中执行命令获取真实的磁盘使用情况
	// 获取当前的集群配置
	diskUsage, err := GetPodDiskUsage(client, config, namespace, podName)
	if err == nil && len(diskUsage) > 0 {
		// 计算所有卷的总使用量
		totalUsedBytes := int64(0)
		for _, usedBytes := range diskUsage {
			totalUsedBytes += usedBytes
		}

		// 计算使用率
		if totalDiskLimits > 0 {
			metrics.DiskUsage = float64(totalUsedBytes) / float64(totalDiskLimits) * 100
		} else if totalDiskRequests > 0 {
			metrics.DiskUsage = float64(totalUsedBytes) / float64(totalDiskRequests) * 100
		} else if _, ok := nodeCapacity["disk"]; ok {
			metrics.DiskUsage = float64(totalUsedBytes) / float64(nodeCapacity["disk"]["capacity"]) * 100
		}
	} else {
		// 如果无法获取真实的磁盘使用情况，回退到估算
		if totalDiskLimits > 0 {
			metrics.DiskUsage = metrics.MemoryUsage * 0.8
		} else if totalDiskRequests > 0 {
			metrics.DiskUsage = metrics.MemoryUsage * 0.9
		} else if _, ok := nodeCapacity["disk"]; ok {
			metrics.DiskUsage = metrics.MemoryUsage * 0.5
		}
	}

	// 确保硬盘使用率不超过100%
	if metrics.DiskUsage > 100 {
		metrics.DiskUsage = 100
	}

	// 使用真实数据来替代模拟数据
	// 首先尝试通过 pod_resource_usage.go 中的实现获取真实的历史数据
	resourceUsage, err = GetPodResourceUsage(client, config, namespace, podName)

	// 检查是否成功获取了真实历史数据
	if err == nil && resourceUsage != nil && len(resourceUsage.Historical) > 0 {
		// 如果有CPU历史数据，使用它
		if cpuHistory, ok := resourceUsage.Historical["cpu"]; ok && len(cpuHistory) > 0 {
			// 清空现有的模拟数据
			metrics.HistoricalData.CPUUsage = []MetricDataPoint{}

			// 添加真实的CPU历史数据
			for _, point := range cpuHistory {
				// 计算CPU使用率百分比
				cpuPercentage := float64(0)
				if totalCPULimits > 0 {
					cpuPercentage = point.Value * 1000 / float64(totalCPULimits) * 100 // 转换为毫核再计算百分比
				} else if totalCPURequests > 0 {
					cpuPercentage = point.Value * 1000 / float64(totalCPURequests) * 100
				} else if capacity, ok := nodeCapacity["cpu"]; ok {
					cpuPercentage = point.Value * 1000 / float64(capacity["capacity"]) * 100
				} else {
					// 如果没有限制或请求，使用当前实际值
					cpuPercentage = metrics.CPUUsage
				}

				// 确保值在有效范围内
				if cpuPercentage < 0 {
					cpuPercentage = 0
				} else if cpuPercentage > 100 {
					cpuPercentage = 100
				}

				metrics.HistoricalData.CPUUsage = append(metrics.HistoricalData.CPUUsage, MetricDataPoint{
					Timestamp: point.Timestamp.Format(time.RFC3339),
					Value:     cpuPercentage,
				})
			}
		}

		// 如果有内存历史数据，使用它
		if memoryHistory, ok := resourceUsage.Historical["memory"]; ok && len(memoryHistory) > 0 {
			// 清空现有的模拟数据
			metrics.HistoricalData.MemoryUsage = []MetricDataPoint{}

			// 添加真实的内存历史数据
			for _, point := range memoryHistory {
				// 计算内存使用率百分比
				memoryPercentage := float64(0)
				if totalMemoryLimits > 0 {
					memoryPercentage = point.Value / float64(totalMemoryLimits) * 100
				} else if totalMemoryRequests > 0 {
					memoryPercentage = point.Value / float64(totalMemoryRequests) * 100
				} else if capacity, ok := nodeCapacity["memory"]; ok {
					memoryPercentage = point.Value / float64(capacity["capacity"]) * 100
				} else {
					// 如果没有限制或请求，使用当前实际值
					memoryPercentage = metrics.MemoryUsage
				}

				// 确保值在有效范围内
				if memoryPercentage < 0 {
					memoryPercentage = 0
				} else if memoryPercentage > 100 {
					memoryPercentage = 100
				}

				metrics.HistoricalData.MemoryUsage = append(metrics.HistoricalData.MemoryUsage, MetricDataPoint{
					Timestamp: point.Timestamp.Format(time.RFC3339),
					Value:     memoryPercentage,
				})
			}
		}
	}

	// 如果没有获取到历史数据或获取历史数据失败，添加当前的实际值作为唯一数据点
	now := time.Now()

	// 如果CPU历史数据为空，添加当前值
	if len(metrics.HistoricalData.CPUUsage) == 0 {
		metrics.HistoricalData.CPUUsage = append(metrics.HistoricalData.CPUUsage, MetricDataPoint{
			Timestamp: now.Format(time.RFC3339),
			Value:     metrics.CPUUsage,
		})
	}

	// 如果内存历史数据为空，添加当前值
	if len(metrics.HistoricalData.MemoryUsage) == 0 {
		metrics.HistoricalData.MemoryUsage = append(metrics.HistoricalData.MemoryUsage, MetricDataPoint{
			Timestamp: now.Format(time.RFC3339),
			Value:     metrics.MemoryUsage,
		})
	}

	// 如果磁盘历史数据为空，添加当前值
	if len(metrics.HistoricalData.DiskUsage) == 0 {
		metrics.HistoricalData.DiskUsage = append(metrics.HistoricalData.DiskUsage, MetricDataPoint{
			Timestamp: now.Format(time.RFC3339),
			Value:     metrics.DiskUsage,
		})
	}

	return metrics, nil
}

// 格式化CPU值
func formatCPU(milliValue int64) string {
	if milliValue == 0 {
		return "0m"
	}

	if milliValue >= 1000 {
		return fmt.Sprintf("%d", milliValue/1000)
	}
	return fmt.Sprintf("%dm", milliValue)
}

// 格式化内存值
func formatMemory(bytes int64) string {
	if bytes == 0 {
		return "0Mi"
	}

	const (
		kilobyte = 1024
		megabyte = 1024 * kilobyte
		gigabyte = 1024 * megabyte
	)

	if bytes >= gigabyte {
		return fmt.Sprintf("%.2fGi", float64(bytes)/float64(gigabyte))
	} else if bytes >= megabyte {
		return fmt.Sprintf("%.0fMi", float64(bytes)/float64(megabyte))
	} else {
		return fmt.Sprintf("%.0fKi", float64(bytes)/float64(kilobyte))
	}
}
