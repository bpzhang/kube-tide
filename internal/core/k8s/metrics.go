package k8s

import (
	"context"
	"fmt"
	"math"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

// ClusterMetrics 集群监控指标
type ClusterMetrics struct {
	Timestamp                string  `json:"timestamp"`
	CPUUsage                 float64 `json:"cpuUsage"`
	MemoryUsage              float64 `json:"memoryUsage"`
	CPURequestsPercentage    float64 `json:"cpuRequestsPercentage"`
	CPULimitsPercentage      float64 `json:"cpuLimitsPercentage"`
	MemoryRequestsPercentage float64 `json:"memoryRequestsPercentage"`
	MemoryLimitsPercentage   float64 `json:"memoryLimitsPercentage"`
	PodCount                 int     `json:"podCount"`
	NodeCounts               struct {
		Ready    int `json:"ready"`
		NotReady int `json:"notReady"`
	} `json:"nodeCounts"`
	DeploymentReadiness struct {
		Available int `json:"available"`
		Total     int `json:"total"`
	} `json:"deploymentReadiness"`
	// 模拟过去24小时的历史数据（实际项目中应从时序数据库获取）
	HistoricalData struct {
		CPUUsage    []MetricDataPoint `json:"cpuUsage"`
		MemoryUsage []MetricDataPoint `json:"memoryUsage"`
		PodCount    []MetricDataPoint `json:"podCount"`
	} `json:"historicalData"`
}

// MetricDataPoint 指标数据点
type MetricDataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// GetClusterMetrics 获取集群监控指标
func GetClusterMetrics(client *kubernetes.Clientset) (*ClusterMetrics, error) {
	ctx := context.Background()
	metrics := &ClusterMetrics{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 获取节点列表和状态
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %v", err)
	}

	// 获取Pod列表
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %v", err)
	}

	// 获取Deployment列表
	deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Deployment列表失败: %v", err)
	}

	// 统计节点状态
	for _, node := range nodes.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}
		if isReady {
			metrics.NodeCounts.Ready++
		} else {
			metrics.NodeCounts.NotReady++
		}
	}

	// 计算Pod总数
	metrics.PodCount = len(pods.Items)

	// 计算Deployment可用情况
	metrics.DeploymentReadiness.Total = len(deployments.Items)
	for _, deployment := range deployments.Items {
		if deployment.Status.AvailableReplicas == deployment.Status.Replicas {
			metrics.DeploymentReadiness.Available++
		}
	}

	// 计算资源使用率和分配率
	metrics = calculateResourceUsage(client, nodes, metrics)

	// 生成历史数据（模拟数据）
	// metrics.HistoricalData = generateHistoricalData()

	return metrics, nil
}

// calculateResourceUsage 计算资源使用率和分配率
func calculateResourceUsage(client *kubernetes.Clientset, nodes *corev1.NodeList, metrics *ClusterMetrics) *ClusterMetrics {
	// 在实际项目中，应该使用metrics-server获取真实的CPU和内存使用情况
	// 这里为了演示，我们模拟一些合理的数据

	var totalCPUCapacity int64
	var totalMemoryCapacity int64
	var totalCPURequests, totalCPULimits int64
	var totalMemoryRequests, totalMemoryLimits int64

	// 计算集群总容量
	for _, node := range nodes.Items {
		cpu := node.Status.Capacity.Cpu()
		memory := node.Status.Capacity.Memory()
		totalCPUCapacity += cpu.Value()
		totalMemoryCapacity += memory.Value()
	}

	// 获取所有Pod的资源请求和限制
	pods, _ := client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			// CPU请求和限制
			if container.Resources.Requests != nil {
				if cpuRequest, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					totalCPURequests += cpuRequest.MilliValue()
				}
				if memoryRequest, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					totalMemoryRequests += memoryRequest.Value()
				}
			}
			if container.Resources.Limits != nil {
				if cpuLimit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
					totalCPULimits += cpuLimit.MilliValue()
				}
				if memoryLimit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
					totalMemoryLimits += memoryLimit.Value()
				}
			}
		}
	}

	// 计算百分比
	if totalCPUCapacity > 0 {
		metrics.CPURequestsPercentage = float64(totalCPURequests) / float64(totalCPUCapacity*1000) * 100
		metrics.CPULimitsPercentage = float64(totalCPULimits) / float64(totalCPUCapacity*1000) * 100
	}
	if totalMemoryCapacity > 0 {
		metrics.MemoryRequestsPercentage = float64(totalMemoryRequests) / float64(totalMemoryCapacity) * 100
		metrics.MemoryLimitsPercentage = float64(totalMemoryLimits) / float64(totalMemoryCapacity) * 100
	}

	// 尝试获取真实的资源使用率（如果有metrics-server的话）
	// 这里先使用模拟数据
	metricsClient, err := getMetricsClient(client)
	if err == nil {
		// 如果有metrics-server，获取真实的使用率数据
		metrics = getRealMetricsData(metricsClient, metrics, totalCPUCapacity, totalMemoryCapacity)
	} else {
		// 没有metrics-server，使用模拟数据
		metrics.CPUUsage = metrics.CPURequestsPercentage * 0.8
		metrics.MemoryUsage = metrics.MemoryRequestsPercentage * 0.7
	}

	return metrics
}

// getMetricsClient 获取metrics-server客户端
func getMetricsClient(client *kubernetes.Clientset) (v1beta1.MetricsV1beta1Interface, error) {
	// 检查metrics-server是否已安装
	_, err := client.CoreV1().Services("kube-system").Get(context.Background(), "metrics-server", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("metrics-server未安装或无法访问: %v", err)
	}

	// 获取当前使用的config
	config, err := getCurrentClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("获取集群配置失败: %v", err)
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建metrics客户端失败: %v", err)
	}

	return metricsClient.MetricsV1beta1(), nil
}

// getCurrentClusterConfig 获取当前集群的配置
func getCurrentClusterConfig() (*rest.Config, error) {
	// 优先使用环境变量中的配置
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// 如果在集群外运行，则尝试使用当前context的kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	return kubeConfig.ClientConfig()
}

// getRealMetricsData 获取真实的指标数据
func getRealMetricsData(metricsClient v1beta1.MetricsV1beta1Interface, metrics *ClusterMetrics, totalCPUCapacity, totalMemoryCapacity int64) *ClusterMetrics {
	ctx := context.Background()

	nodeMetrics, err := metricsClient.NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		// 如果获取失败，使用模拟数据
		metrics.CPUUsage = metrics.CPURequestsPercentage * 0.8
		metrics.MemoryUsage = metrics.MemoryRequestsPercentage * 0.7
		return metrics
	}

	var totalCPUUsage, totalMemoryUsage int64
	for _, nodeMetric := range nodeMetrics.Items {
		cpuUsage := nodeMetric.Usage.Cpu().MilliValue()
		memoryUsage := nodeMetric.Usage.Memory().Value()
		totalCPUUsage += cpuUsage
		totalMemoryUsage += memoryUsage
	}

	// 计算使用率百分比
	if totalCPUCapacity > 0 {
		metrics.CPUUsage = float64(totalCPUUsage) / float64(totalCPUCapacity*1000) * 100
	}
	if totalMemoryCapacity > 0 {
		metrics.MemoryUsage = float64(totalMemoryUsage) / float64(totalMemoryCapacity) * 100
	}

	return metrics
}

// generateHistoricalData 生成历史数据（模拟数据）
func generateHistoricalData() struct {
	CPUUsage    []MetricDataPoint
	MemoryUsage []MetricDataPoint
	PodCount    []MetricDataPoint
} {
	now := time.Now()
	data := struct {
		CPUUsage    []MetricDataPoint
		MemoryUsage []MetricDataPoint
		PodCount    []MetricDataPoint
	}{
		CPUUsage:    make([]MetricDataPoint, 24),
		MemoryUsage: make([]MetricDataPoint, 24),
		PodCount:    make([]MetricDataPoint, 24),
	}

	// 使用正弦波模拟一天内的资源使用变化
	for i := 0; i < 24; i++ {
		timestamp := now.Add(time.Duration(-23+i) * time.Hour).Format(time.RFC3339)

		// CPU使用率在30%-70%之间波动
		cpuValue := 50 + 20*math.Sin(float64(i)/3.0)
		// 内存使用率在40%-80%之间波动
		memValue := 60 + 20*math.Sin(float64(i)/4.0)
		// Pod数量在100-150之间波动
		podValue := 125 + 25*math.Sin(float64(i)/6.0)

		data.CPUUsage[i] = MetricDataPoint{
			Timestamp: timestamp,
			Value:     math.Round(cpuValue*10) / 10,
		}
		data.MemoryUsage[i] = MetricDataPoint{
			Timestamp: timestamp,
			Value:     math.Round(memValue*10) / 10,
		}
		data.PodCount[i] = MetricDataPoint{
			Timestamp: timestamp,
			Value:     math.Round(podValue),
		}
	}

	return data
}
