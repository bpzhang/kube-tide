package k8s

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workloadHealthHealthy  = "healthy"
	workloadHealthWarning  = "warning"
	workloadHealthCritical = "critical"
	workloadHealthUnknown  = "unknown"

	defaultCPUWarning     = 70.0
	defaultCPUCritical    = 85.0
	defaultMemoryWarning  = 70.0
	defaultMemoryCritical = 85.0
)

// WorkloadMetrics 无状态应用（Deployment）级别监控汇总
type WorkloadMetrics struct {
	WorkloadType   string                     `json:"workloadType"`
	Name           string                     `json:"name"`
	Namespace      string                     `json:"namespace"`
	Summary        WorkloadMetricsSummary     `json:"summary"`
	Strategy       WorkloadMonitoringStrategy `json:"monitoringStrategy"`
	Pods           []WorkloadPodMetrics       `json:"pods"`
	ContainerGroups []ContainerGroupMetrics   `json:"containerGroups"`
	HistoricalData struct {
		CPUUsage    []MetricDataPoint `json:"cpuUsage"`
		MemoryUsage []MetricDataPoint `json:"memoryUsage"`
		DiskUsage   []MetricDataPoint `json:"diskUsage"`
	} `json:"historicalData"`
}

// WorkloadMetricsSummary 应用级汇总指标
type WorkloadMetricsSummary struct {
	Replicas          int32   `json:"replicas"`
	ReadyReplicas     int32   `json:"readyReplicas"`
	AvailableReplicas int32   `json:"availableReplicas"`
	PodCount          int     `json:"podCount"`
	RunningPods       int     `json:"runningPods"`
	MetricsPodCount   int     `json:"metricsPodCount"`
	AvgCPUUsage       float64 `json:"avgCpuUsage"`
	MaxCPUUsage       float64 `json:"maxCpuUsage"`
	AvgMemoryUsage    float64 `json:"avgMemoryUsage"`
	MaxMemoryUsage    float64 `json:"maxMemoryUsage"`
	AvgDiskUsed       string  `json:"avgDiskUsed"`
	MaxDiskUsed       string  `json:"maxDiskUsed"`
	TotalDiskUsed     string  `json:"totalDiskUsed"`
	AvgDiskUsedBytes  int64   `json:"avgDiskUsedBytes"`
	MaxDiskUsedBytes  int64   `json:"maxDiskUsedBytes"`
	TotalDiskUsedBytes int64  `json:"totalDiskUsedBytes"`
	CPURequests       string  `json:"cpuRequests"`
	CPULimits         string  `json:"cpuLimits"`
	MemoryRequests    string  `json:"memoryRequests"`
	MemoryLimits      string  `json:"memoryLimits"`
	DiskRequests      string  `json:"diskRequests"`
	DiskLimits        string  `json:"diskLimits"`
	HealthStatus      string  `json:"healthStatus"`
	Alerts            []WorkloadAlert `json:"alerts"`
}

// WorkloadMonitoringStrategy 容器组监控策略说明
type WorkloadMonitoringStrategy struct {
	Policy         string               `json:"policy"`
	Description    string               `json:"description"`
	Thresholds     MonitoringThresholds `json:"thresholds"`
	PodCoverage    string               `json:"podCoverage"`
	Recommendation string               `json:"recommendation"`
}

// MonitoringThresholds 监控阈值
type MonitoringThresholds struct {
	CPUWarning     float64 `json:"cpuWarning"`
	CPUCritical    float64 `json:"cpuCritical"`
	MemoryWarning  float64 `json:"memoryWarning"`
	MemoryCritical float64 `json:"memoryCritical"`
}

// WorkloadPodMetrics Pod 级监控摘要
type WorkloadPodMetrics struct {
	Name         string  `json:"name"`
	Phase        string  `json:"phase"`
	Ready        bool    `json:"ready"`
	CPUUsage     float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	DiskUsed       string  `json:"diskUsed"`
	DiskUsedBytes  int64   `json:"diskUsedBytes"`
	HealthStatus   string  `json:"healthStatus"`
	Restarts     int32   `json:"restarts"`
}

// ContainerGroupMetrics 按容器名汇总的容器组指标（跨 Pod 聚合）
type ContainerGroupMetrics struct {
	Name           string  `json:"name"`
	PodCount       int     `json:"podCount"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	MaxCPUUsage    float64 `json:"maxCpuUsage"`
	MinCPUUsage    float64 `json:"minCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	MaxMemoryUsage float64 `json:"maxMemoryUsage"`
	MinMemoryUsage float64 `json:"minMemoryUsage"`
	AvgDiskUsed    string  `json:"avgDiskUsed"`
	MaxDiskUsed    string  `json:"maxDiskUsed"`
	MinDiskUsed    string  `json:"minDiskUsed"`
	AvgDiskUsedBytes int64 `json:"avgDiskUsedBytes"`
	MaxDiskUsedBytes int64 `json:"maxDiskUsedBytes"`
	MinDiskUsedBytes int64 `json:"minDiskUsedBytes"`
	CPURequests    string  `json:"cpuRequests"`
	CPULimits      string  `json:"cpuLimits"`
	MemoryRequests string  `json:"memoryRequests"`
	MemoryLimits   string  `json:"memoryLimits"`
	DiskRequests   string  `json:"diskRequests"`
	DiskLimits     string  `json:"diskLimits"`
	HealthStatus   string  `json:"healthStatus"`
}

// WorkloadAlert 监控告警项
type WorkloadAlert struct {
	Level   string  `json:"level"`
	Source  string  `json:"source"`
	Metric  string  `json:"metric"`
	Value   float64 `json:"value"`
	Message string  `json:"message"`
}

// GetDeploymentMetrics 获取 Deployment 应用级监控汇总
func (s *PodMetricsService) GetDeploymentMetrics(ctx context.Context, clusterName, namespace, deploymentName string) (*WorkloadMetrics, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析标签选择器失败: %w", err)
	}

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("查询 Pod 列表失败: %w", err)
	}

	thresholds := MonitoringThresholds{
		CPUWarning:     defaultCPUWarning,
		CPUCritical:    defaultCPUCritical,
		MemoryWarning:  defaultMemoryWarning,
		MemoryCritical: defaultMemoryCritical,
	}

	result := &WorkloadMetrics{
		WorkloadType: "deployment",
		Name:         deploymentName,
		Namespace:    namespace,
		Strategy: WorkloadMonitoringStrategy{
			Policy:      "container-group-aggregate",
			Description: "按 Deployment 容器组名称跨 Pod 聚合 CPU/内存使用率与磁盘占用量，并结合副本就绪率评估应用健康度",
			Thresholds:  thresholds,
		},
		Pods:            make([]WorkloadPodMetrics, 0),
		ContainerGroups: make([]ContainerGroupMetrics, 0),
	}
	result.HistoricalData.CPUUsage = make([]MetricDataPoint, 0)
	result.HistoricalData.MemoryUsage = make([]MetricDataPoint, 0)
	result.HistoricalData.DiskUsage = make([]MetricDataPoint, 0)

	result.Summary.Replicas = derefInt32(deployment.Spec.Replicas)
	result.Summary.ReadyReplicas = deployment.Status.ReadyReplicas
	result.Summary.AvailableReplicas = deployment.Status.AvailableReplicas
	result.Summary.PodCount = len(pods.Items)

	containerGroupMap := make(map[string]*containerGroupAccumulator)
	var alerts []WorkloadAlert
	var cpuUsages, memUsages []float64
	var diskUsedBytesList []int64
	var totalCPUReqMilli, totalCPULimitMilli, totalMemReq, totalMemLimit int64
	var totalDiskReq, totalDiskLimit int64
	metricsSuccess := 0
	historyBuckets := make(map[string]*historyBucket)

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			result.Summary.RunningPods++
		}

		podSummary := WorkloadPodMetrics{
			Name:     pod.Name,
			Phase:    string(pod.Status.Phase),
			Ready:    isPodReady(&pod),
			Restarts: countPodRestarts(&pod),
		}

		podMetrics, err := s.GetPodMetrics(ctx, clusterName, namespace, pod.Name)
		if err != nil {
			podSummary.HealthStatus = workloadHealthUnknown
			result.Pods = append(result.Pods, podSummary)
			continue
		}

		metricsSuccess++
		podSummary.CPUUsage = podMetrics.CPUUsage
		podSummary.MemoryUsage = podMetrics.MemoryUsage
		podSummary.DiskUsedBytes = podMetrics.DiskUsedBytes
		podSummary.DiskUsed = podMetrics.DiskUsed
		podSummary.HealthStatus = evaluateWorkloadHealth(
			podMetrics.CPUUsage, podMetrics.MemoryUsage, thresholds,
		)
		cpuUsages = append(cpuUsages, podMetrics.CPUUsage)
		memUsages = append(memUsages, podMetrics.MemoryUsage)
		diskUsedBytesList = append(diskUsedBytesList, podMetrics.DiskUsedBytes)

		totalCPUReqMilli += quantityToMilliCPU(podMetrics.CPURequests)
		totalCPULimitMilli += quantityToMilliCPU(podMetrics.CPULimits)
		totalMemReq += quantityToBytes(podMetrics.MemoryRequests)
		totalMemLimit += quantityToBytes(podMetrics.MemoryLimits)
		totalDiskReq += quantityToBytes(podMetrics.DiskRequests)
		totalDiskLimit += quantityToBytes(podMetrics.DiskLimits)

		alerts = append(alerts, buildUsageAlerts(
			pod.Name, podMetrics.CPUUsage, podMetrics.MemoryUsage, thresholds,
		)...)

		for _, container := range podMetrics.Containers {
			acc, ok := containerGroupMap[container.Name]
			if !ok {
				acc = &containerGroupAccumulator{
					name:           container.Name,
					cpuRequests:    container.CPURequests,
					cpuLimits:      container.CPULimits,
					memoryRequests: container.MemoryRequests,
					memoryLimits:   container.MemoryLimits,
					diskRequests:   container.DiskRequests,
					diskLimits:     container.DiskLimits,
				}
				containerGroupMap[container.Name] = acc
			}
			acc.add(container.CPUUsage, container.MemoryUsage, container.DiskUsedBytes)
			alerts = append(alerts, buildUsageAlerts(
				fmt.Sprintf("%s/%s", pod.Name, container.Name),
				container.CPUUsage,
				container.MemoryUsage,
				thresholds,
			)...)
		}

		mergeHistorical(
			historyBuckets,
			podMetrics.HistoricalData.CPUUsage,
			podMetrics.HistoricalData.MemoryUsage,
			podMetrics.HistoricalData.DiskUsage,
		)
		result.Pods = append(result.Pods, podSummary)
	}

	// 补充 Deployment 模板中定义但当前无运行实例的容器组
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if _, ok := containerGroupMap[c.Name]; !ok {
			containerGroupMap[c.Name] = &containerGroupAccumulator{
				name:           c.Name,
				cpuRequests:    formatCPU(c.Resources.Requests.Cpu().MilliValue()),
				cpuLimits:      formatCPU(c.Resources.Limits.Cpu().MilliValue()),
				memoryRequests: formatMemory(c.Resources.Requests.Memory().Value()),
				memoryLimits:   formatMemory(c.Resources.Limits.Memory().Value()),
			}
		}
	}

	result.Summary.MetricsPodCount = metricsSuccess
	result.Summary.CPURequests = formatCPU(totalCPUReqMilli)
	result.Summary.CPULimits = formatCPU(totalCPULimitMilli)
	result.Summary.MemoryRequests = formatMemory(totalMemReq)
	result.Summary.MemoryLimits = formatMemory(totalMemLimit)
	result.Summary.DiskRequests = FormatStorage(totalDiskReq)
	result.Summary.DiskLimits = FormatStorage(totalDiskLimit)
	result.Summary.AvgCPUUsage = average(cpuUsages)
	result.Summary.MaxCPUUsage = maxFloat(cpuUsages)
	result.Summary.AvgMemoryUsage = average(memUsages)
	result.Summary.MaxMemoryUsage = maxFloat(memUsages)
	result.Summary.TotalDiskUsedBytes = sumInt64(diskUsedBytesList)
	result.Summary.AvgDiskUsedBytes = averageInt64(diskUsedBytesList)
	result.Summary.MaxDiskUsedBytes = maxInt64(diskUsedBytesList)
	result.Summary.TotalDiskUsed = FormatStorage(result.Summary.TotalDiskUsedBytes)
	result.Summary.AvgDiskUsed = FormatStorage(result.Summary.AvgDiskUsedBytes)
	result.Summary.MaxDiskUsed = FormatStorage(result.Summary.MaxDiskUsedBytes)
	result.Summary.Alerts = dedupeAlerts(alerts)
	result.Summary.HealthStatus = evaluateWorkloadHealth(
		result.Summary.MaxCPUUsage, result.Summary.MaxMemoryUsage, thresholds,
	)

	if result.Summary.ReadyReplicas < result.Summary.Replicas {
		result.Summary.Alerts = append(result.Summary.Alerts, WorkloadAlert{
			Level:   workloadHealthWarning,
			Source:  deploymentName,
			Metric:  "replicas",
			Message: fmt.Sprintf("就绪副本 %d/%d", result.Summary.ReadyReplicas, result.Summary.Replicas),
		})
		if result.Summary.ReadyReplicas == 0 && result.Summary.Replicas > 0 {
			result.Summary.HealthStatus = workloadHealthCritical
		} else if result.Summary.HealthStatus == workloadHealthHealthy {
			result.Summary.HealthStatus = workloadHealthWarning
		}
	}

	for _, acc := range containerGroupMap {
		result.ContainerGroups = append(result.ContainerGroups, acc.toMetrics(thresholds))
	}
	sort.Slice(result.ContainerGroups, func(i, j int) bool {
		return result.ContainerGroups[i].Name < result.ContainerGroups[j].Name
	})
	sort.Slice(result.Pods, func(i, j int) bool {
		return result.Pods[i].Name < result.Pods[j].Name
	})

	result.Strategy.PodCoverage = fmt.Sprintf("%d/%d Pods 上报指标", metricsSuccess, len(pods.Items))
	result.Strategy.Recommendation = buildMonitoringRecommendation(result.Summary, thresholds)

	result.HistoricalData.CPUUsage = finalizeHistorical(historyBuckets, "cpu")
	result.HistoricalData.MemoryUsage = finalizeHistorical(historyBuckets, "memory")
	result.HistoricalData.DiskUsage = finalizeHistorical(historyBuckets, "disk")

	return result, nil
}

type containerGroupAccumulator struct {
	name           string
	podCount       int
	cpuValues      []float64
	memoryValues   []float64
	diskBytes      []int64
	cpuRequests    string
	cpuLimits      string
	memoryRequests string
	memoryLimits   string
	diskRequests   string
	diskLimits     string
}

func (a *containerGroupAccumulator) add(cpu, memory float64, diskBytes int64) {
	a.podCount++
	a.cpuValues = append(a.cpuValues, cpu)
	a.memoryValues = append(a.memoryValues, memory)
	a.diskBytes = append(a.diskBytes, diskBytes)
}

func (a *containerGroupAccumulator) toMetrics(thresholds MonitoringThresholds) ContainerGroupMetrics {
	maxCPU := maxFloat(a.cpuValues)
	maxMem := maxFloat(a.memoryValues)
	avgDisk := averageInt64(a.diskBytes)
	maxDisk := maxInt64(a.diskBytes)
	minDisk := minInt64(a.diskBytes)
	return ContainerGroupMetrics{
		Name:           a.name,
		PodCount:       a.podCount,
		AvgCPUUsage:    average(a.cpuValues),
		MaxCPUUsage:    maxCPU,
		MinCPUUsage:    minFloat(a.cpuValues),
		AvgMemoryUsage: average(a.memoryValues),
		MaxMemoryUsage: maxMem,
		MinMemoryUsage: minFloat(a.memoryValues),
		AvgDiskUsed:       FormatStorage(avgDisk),
		MaxDiskUsed:       FormatStorage(maxDisk),
		MinDiskUsed:       FormatStorage(minDisk),
		AvgDiskUsedBytes:  avgDisk,
		MaxDiskUsedBytes:  maxDisk,
		MinDiskUsedBytes:  minDisk,
		CPURequests:    a.cpuRequests,
		CPULimits:      a.cpuLimits,
		MemoryRequests: a.memoryRequests,
		MemoryLimits:   a.memoryLimits,
		DiskRequests:   a.diskRequests,
		DiskLimits:     a.diskLimits,
		HealthStatus:   evaluateWorkloadHealth(maxCPU, maxMem, thresholds),
	}
}

type historyBucket struct {
	cpuSum   float64
	cpuCount int
	memSum   float64
	memCount int
	diskSum  float64
	diskCount int
}

func mergeHistorical(buckets map[string]*historyBucket, cpuHistory, memHistory, diskHistory []MetricDataPoint) {
	for _, point := range cpuHistory {
		key := normalizeHistoryKey(point.Timestamp)
		b, ok := buckets[key]
		if !ok {
			b = &historyBucket{}
			buckets[key] = b
		}
		b.cpuSum += point.Value
		b.cpuCount++
	}
	for _, point := range memHistory {
		key := normalizeHistoryKey(point.Timestamp)
		b, ok := buckets[key]
		if !ok {
			b = &historyBucket{}
			buckets[key] = b
		}
		b.memSum += point.Value
		b.memCount++
	}
	for _, point := range diskHistory {
		key := normalizeHistoryKey(point.Timestamp)
		b, ok := buckets[key]
		if !ok {
			b = &historyBucket{}
			buckets[key] = b
		}
		b.diskSum += point.Value
		b.diskCount++
	}
}

func finalizeHistorical(buckets map[string]*historyBucket, metric string) []MetricDataPoint {
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	points := make([]MetricDataPoint, 0, len(keys))
	for _, key := range keys {
		b := buckets[key]
		var value float64
		switch metric {
		case "memory":
			if b.memCount == 0 {
				continue
			}
			value = b.memSum / float64(b.memCount)
		case "disk":
			if b.diskCount == 0 {
				continue
			}
			value = b.diskSum / float64(b.diskCount)
		default:
			if b.cpuCount == 0 {
				continue
			}
			value = b.cpuSum / float64(b.cpuCount)
		}
		points = append(points, MetricDataPoint{
			Timestamp: key,
			Value:     round2(value),
		})
	}
	return points
}

func normalizeHistoryKey(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.UTC().Format(time.RFC3339)
}

func evaluateWorkloadHealth(cpuUsage, memoryUsage float64, thresholds MonitoringThresholds) string {
	if cpuUsage == 0 && memoryUsage == 0 {
		return workloadHealthUnknown
	}
	if cpuUsage >= thresholds.CPUCritical || memoryUsage >= thresholds.MemoryCritical {
		return workloadHealthCritical
	}
	if cpuUsage >= thresholds.CPUWarning || memoryUsage >= thresholds.MemoryWarning {
		return workloadHealthWarning
	}
	return workloadHealthHealthy
}

func buildUsageAlerts(source string, cpu, memory float64, thresholds MonitoringThresholds) []WorkloadAlert {
	var alerts []WorkloadAlert
	if cpu >= thresholds.CPUCritical {
		alerts = append(alerts, WorkloadAlert{
			Level: workloadHealthCritical, Source: source, Metric: "cpu", Value: cpu,
			Message: fmt.Sprintf("CPU 使用率 %.1f%% 超过严重阈值 %.0f%%", cpu, thresholds.CPUCritical),
		})
	} else if cpu >= thresholds.CPUWarning {
		alerts = append(alerts, WorkloadAlert{
			Level: workloadHealthWarning, Source: source, Metric: "cpu", Value: cpu,
			Message: fmt.Sprintf("CPU 使用率 %.1f%% 超过预警阈值 %.0f%%", cpu, thresholds.CPUWarning),
		})
	}
	if memory >= thresholds.MemoryCritical {
		alerts = append(alerts, WorkloadAlert{
			Level: workloadHealthCritical, Source: source, Metric: "memory", Value: memory,
			Message: fmt.Sprintf("内存使用率 %.1f%% 超过严重阈值 %.0f%%", memory, thresholds.MemoryCritical),
		})
	} else if memory >= thresholds.MemoryWarning {
		alerts = append(alerts, WorkloadAlert{
			Level: workloadHealthWarning, Source: source, Metric: "memory", Value: memory,
			Message: fmt.Sprintf("内存使用率 %.1f%% 超过预警阈值 %.0f%%", memory, thresholds.MemoryWarning),
		})
	}
	return alerts
}

func buildMonitoringRecommendation(summary WorkloadMetricsSummary, thresholds MonitoringThresholds) string {
	if summary.MetricsPodCount == 0 {
		return "未获取到 Pod 指标，请确认集群已安装 metrics-server"
	}
	if summary.MaxCPUUsage >= thresholds.CPUCritical ||
		summary.MaxMemoryUsage >= thresholds.MemoryCritical {
		return "应用存在资源瓶颈，建议检查容器 limits/requests 或考虑水平扩容"
	}
	if summary.ReadyReplicas < summary.Replicas {
		return "副本未全部就绪，建议结合 Pod 事件与探针配置排查"
	}
	if summary.MaxCPUUsage >= thresholds.CPUWarning ||
		summary.MaxMemoryUsage >= thresholds.MemoryWarning {
		return "部分容器组资源使用率偏高，建议持续观察或调整资源配额"
	}
	return "应用资源使用正常，可继续按当前监控策略观察"
}

func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func countPodRestarts(pod *corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func quantityToMilliCPU(s string) int64 {
	if s == "" {
		return 0
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return 0
	}
	return q.MilliValue()
}

func quantityToBytes(s string) int64 {
	if s == "" {
		return 0
	}
	q, err := resource.ParseQuantity(s)
	if err != nil {
		return 0
	}
	return q.Value()
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return round2(sum / float64(len(values)))
}

func maxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return round2(max)
}

func minFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return round2(min)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func sumInt64(values []int64) int64 {
	var sum int64
	for _, v := range values {
		sum += v
	}
	return sum
}

func averageInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	return sumInt64(values) / int64(len(values))
}

func maxInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func minInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func dedupeAlerts(alerts []WorkloadAlert) []WorkloadAlert {
	if len(alerts) == 0 {
		return alerts
	}
	seen := make(map[string]struct{})
	result := make([]WorkloadAlert, 0, len(alerts))
	for _, alert := range alerts {
		key := fmt.Sprintf("%s|%s|%s|%.1f", alert.Level, alert.Source, alert.Metric, alert.Value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, alert)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Level != result[j].Level {
			return result[i].Level == workloadHealthCritical
		}
		return result[i].Source < result[j].Source
	})
	return result
}
