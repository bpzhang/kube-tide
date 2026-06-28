package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodMetrics Pod指标结构
type PodMetrics struct {
	// 当前CPU使用率（百分比）
	CPUUsage float64 `json:"cpuUsage"`
	// 当前内存使用率（百分比）
	MemoryUsage float64 `json:"memoryUsage"`
	// 当前磁盘占用（字节）
	DiskUsedBytes int64 `json:"diskUsedBytes"`
	// 当前磁盘占用（可读格式）
	DiskUsed string `json:"diskUsed"`
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
	DiskUsedBytes  int64   `json:"diskUsedBytes"`
	DiskUsed       string  `json:"diskUsed"`
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
func GetPodMetrics(client *kubernetes.Clientset, config *rest.Config, namespace, podName string) (*PodMetrics, error) {
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

	// 创建 metrics 客户端（使用已注册集群的 kubeconfig）
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

	// 汇总 ephemeral-storage 配额与 PVC 容量
	totalDiskRequests, totalDiskLimits = sumEphemeralStorageResources(pod, client, namespace)

	metrics.CPURequests = formatCPU(totalCPURequests)
	metrics.CPULimits = formatCPU(totalCPULimits)
	metrics.MemoryRequests = formatMemory(totalMemoryRequests)
	metrics.MemoryLimits = formatMemory(totalMemoryLimits)
	metrics.DiskRequests = FormatStorage(totalDiskRequests)
	metrics.DiskLimits = FormatStorage(totalDiskLimits)

	diskStats := GetPodDiskStats(client, config, pod, podMetrics)

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
			} // 容器的磁盘占用
			diskRequests, diskLimits := containerStorageQuota(pod, containerMetrics.Name, client, namespace)

			containerDiskUsage := diskStats.containerBytes(containerMetrics.Name)
			if containerDiskUsage == 0 && diskStats != nil && diskStats.PodUsedBytes > 0 && totalMemoryUsage > 0 && containerMetrics.MemoryUsage > 0 {
				ratio := float64(containerMetrics.MemoryUsage) / float64(totalMemoryUsage)
				containerDiskUsage = int64(ratio * float64(diskStats.PodUsedBytes))
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
				Value:     float64(containerDiskUsage),
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
								containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     point.Value,
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
				DiskUsedBytes:  containerDiskUsage,
				DiskUsed:       FormatStorage(containerDiskUsage),
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

			containerDiskUsage := diskStats.containerBytes(containerMetrics.Name)
			if containerDiskUsage == 0 && diskStats != nil && diskStats.PodUsedBytes > 0 && totalMemoryUsage > 0 && memoryUsage > 0 {
				ratio := float64(memoryUsage) / float64(totalMemoryUsage)
				containerDiskUsage = int64(ratio * float64(diskStats.PodUsedBytes))
			}

			diskRequests, diskLimits := containerStorageQuota(pod, containerMetrics.Name, client, namespace)

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
				Value:     float64(containerDiskUsage),
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
								containerHistorical.DiskUsage = append(containerHistorical.DiskUsage, MetricDataPoint{
									Timestamp: point.Timestamp.Format(time.RFC3339),
									Value:     point.Value,
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
				DiskUsedBytes:  containerDiskUsage,
				DiskUsed:       FormatStorage(containerDiskUsage),
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

	// Pod 总体磁盘占用（kubelet / metrics-server / exec）
	if diskStats != nil && diskStats.PodUsedBytes > 0 {
		metrics.DiskUsedBytes = diskStats.PodUsedBytes
		metrics.DiskUsed = FormatStorage(diskStats.PodUsedBytes)
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

	// 如果磁盘历史数据为空，添加当前值（字节）
	if len(metrics.HistoricalData.DiskUsage) == 0 {
		metrics.HistoricalData.DiskUsage = append(metrics.HistoricalData.DiskUsage, MetricDataPoint{
			Timestamp: now.Format(time.RFC3339),
			Value:     float64(metrics.DiskUsedBytes),
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
