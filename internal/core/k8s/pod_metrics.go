package k8s

import (
	"context"
	"fmt"
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
	// CPU请求值（单位：m）
	CPURequests string `json:"cpuRequests"`
	// CPU限制值（单位：m）
	CPULimits string `json:"cpuLimits"`
	// 内存请求值（例如：100Mi）
	MemoryRequests string `json:"memoryRequests"`
	// 内存限制值（例如：200Mi）
	MemoryLimits string `json:"memoryLimits"`
	// 历史数据（24小时内每小时一个数据点）
	HistoricalData struct {
		CPUUsage    []MetricDataPoint `json:"cpuUsage"`
		MemoryUsage []MetricDataPoint `json:"memoryUsage"`
	} `json:"historicalData"`
	// 容器指标
	Containers []ContainerMetrics `json:"containers"`
}

// ContainerMetrics 容器指标结构
type ContainerMetrics struct {
	Name           string  `json:"name"`
	CPUUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	CPURequests    string  `json:"cpuRequests"`
	CPULimits      string  `json:"cpuLimits"`
	MemoryRequests string  `json:"memoryRequests"`
	MemoryLimits   string  `json:"memoryLimits"`
}

// GetPodMetrics 获取Pod监控指标
func GetPodMetrics(client *kubernetes.Clientset, namespace, podName string) (*PodMetrics, error) {
	ctx := context.Background()
	metrics := &PodMetrics{
		HistoricalData: struct {
			CPUUsage    []MetricDataPoint `json:"cpuUsage"`
			MemoryUsage []MetricDataPoint `json:"memoryUsage"`
		}{
			CPUUsage:    make([]MetricDataPoint, 0),
			MemoryUsage: make([]MetricDataPoint, 0),
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
		}
	}

	// 计算Pod的总CPU和内存请求/限制
	totalCPURequests := int64(0)
	totalCPULimits := int64(0)
	totalMemoryRequests := int64(0)
	totalMemoryLimits := int64(0)

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

	metrics.CPURequests = formatCPU(totalCPURequests)
	metrics.CPULimits = formatCPU(totalCPULimits)
	metrics.MemoryRequests = formatMemory(totalMemoryRequests)
	metrics.MemoryLimits = formatMemory(totalMemoryLimits)

	// 计算容器指标
	totalCPUUsage := int64(0)
	totalMemoryUsage := int64(0)

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

		// 添加到容器指标列表
		metrics.Containers = append(metrics.Containers, ContainerMetrics{
			Name:           containerMetrics.Name,
			CPUUsage:       cpuUsagePercentage,
			MemoryUsage:    memoryUsagePercentage,
			CPURequests:    formatCPU(cpuRequests),
			CPULimits:      formatCPU(cpuLimits),
			MemoryRequests: formatMemory(memoryRequests),
			MemoryLimits:   formatMemory(memoryLimits),
		})
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

	// 生成模拟的历史数据
	now := time.Now()
	for i := 23; i >= 0; i-- {
		t := now.Add(time.Duration(-i) * time.Hour)
		// 简单模拟一些波动的数据
		cpuVariation := float64(i%5) * 0.8
		memoryVariation := float64(i%3) * 0.5

		// 确保CPU和内存使用率不小于0且不超过100%
		cpuValue := metrics.CPUUsage + cpuVariation
		if cpuValue < 0 {
			cpuValue = 0
		} else if cpuValue > 100 {
			cpuValue = 100
		}

		memoryValue := metrics.MemoryUsage + memoryVariation
		if memoryValue < 0 {
			memoryValue = 0
		} else if memoryValue > 100 {
			memoryValue = 100
		}

		metrics.HistoricalData.CPUUsage = append(metrics.HistoricalData.CPUUsage, MetricDataPoint{
			Timestamp: t.Format(time.RFC3339),
			Value:     cpuValue,
		})

		metrics.HistoricalData.MemoryUsage = append(metrics.HistoricalData.MemoryUsage, MetricDataPoint{
			Timestamp: t.Format(time.RFC3339),
			Value:     memoryValue,
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
