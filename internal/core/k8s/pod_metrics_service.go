package k8s

import (
	"context"
	"time"

	"kube-tide/internal/utils/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodMetricsService pod指标服务
type PodMetricsService struct {
	clientManager *ClientManager
	metricsCache  *MemoryMetricsCache
}

// NewPodMetricsService 创建一个新的Pod指标服务
func NewPodMetricsService(clientManager *ClientManager) *PodMetricsService {
	// 创建MemoryMetricsCache实例，设置TTL为24小时，最大缓存大小为1000，聚合间隔为1小时
	cache := NewMemoryMetricsCache(24*time.Hour, DefaultMaxCacheSize, DefaultAggregationInterval)
	return &PodMetricsService{
		clientManager: clientManager,
		metricsCache:  cache,
	}
}

// GetPodMetrics 获取Pod的CPU和内存监控指标
func (s *PodMetricsService) GetPodMetrics(ctx context.Context, clusterName, namespace, podName string) (*PodMetrics, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return nil, err
	}

	// 首先尝试从缓存获取指标数据
	if metrics, found := s.metricsCache.GetPodMetrics(namespace, podName); found {
		currentMetrics, err := GetPodMetrics(client, config, namespace, podName)
		if err == nil {
			now := time.Now().Format(time.RFC3339)

			metrics.HistoricalData.CPUUsage = append(metrics.HistoricalData.CPUUsage, MetricDataPoint{
				Timestamp: now,
				Value:     currentMetrics.CPUUsage,
			})
			metrics.HistoricalData.MemoryUsage = append(metrics.HistoricalData.MemoryUsage, MetricDataPoint{
				Timestamp: now,
				Value:     currentMetrics.MemoryUsage,
			})
			metrics.HistoricalData.DiskUsage = append(metrics.HistoricalData.DiskUsage, MetricDataPoint{
				Timestamp: now,
				Value:     float64(currentMetrics.DiskUsedBytes),
			})

			metrics.CPUUsage = currentMetrics.CPUUsage
			metrics.MemoryUsage = currentMetrics.MemoryUsage
			metrics.DiskUsedBytes = currentMetrics.DiskUsedBytes
			metrics.DiskUsed = currentMetrics.DiskUsed
			metrics.Containers = currentMetrics.Containers

			s.metricsCache.SetPodMetrics(namespace, podName, metrics)

			logger.Debug("更新Pod指标并添加历史数据点",
				"namespace", namespace,
				"pod", podName,
				"timestamp", now,
				"cpuUsage", currentMetrics.CPUUsage,
				"memoryUsage", currentMetrics.MemoryUsage,
				"diskUsed", currentMetrics.DiskUsed,
				"historyPoints", len(metrics.HistoricalData.CPUUsage))
		}

		return metrics, nil
	}

	metrics, err := GetPodMetrics(client, config, namespace, podName)
	if err != nil {
		return nil, err
	}

	s.metricsCache.SetPodMetrics(namespace, podName, metrics)

	logger.Debug("首次获取Pod指标数据",
		"namespace", namespace,
		"pod", podName,
		"cpuUsage", metrics.CPUUsage,
		"memoryUsage", metrics.MemoryUsage,
		"diskUsed", metrics.DiskUsed)

	return metrics, nil
}

// StartPeriodicMetricsCollection 开始定期收集指标
func (s *PodMetricsService) StartPeriodicMetricsCollection(ctx context.Context, clusterName string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logger.Info("启动定期Pod指标收集任务", "cluster", clusterName, "interval", interval.String())

		for {
			select {
			case <-ctx.Done():
				logger.Info("停止Pod指标收集任务", "cluster", clusterName)
				return
			case <-ticker.C:
				s.collectAllPodsMetrics(ctx, clusterName)
			}
		}
	}()
}

// collectAllPodsMetrics 收集所有Pod的指标数据
func (s *PodMetricsService) collectAllPodsMetrics(ctx context.Context, clusterName string) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		logger.Error("获取K8s客户端失败", "cluster", clusterName, "error", err)
		return
	}
	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		logger.Error("获取集群配置失败", "cluster", clusterName, "error", err)
		return
	}

	// 记录开始收集指标的时间
	startTime := time.Now()
	logger.Debug("开始收集Pod指标数据", "cluster", clusterName, "time", startTime.Format(time.RFC3339))

	// 获取所有命名空间
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error("获取命名空间列表失败", "cluster", clusterName, "error", err)
		return
	}

	// 统计信息
	totalPods := 0
	successPods := 0

	// 遍历所有命名空间和Pod，收集指标
	for _, ns := range namespaces.Items {
		pods, err := client.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("获取Pod列表失败", "namespace", ns.Name, "error", err)
			continue
		}

		totalPods += len(pods.Items)

		for _, pod := range pods.Items {
			metrics, err := GetPodMetrics(client, config, ns.Name, pod.Name)
			if err != nil {
				logger.Debug("获取Pod指标失败", "namespace", ns.Name, "pod", pod.Name, "error", err)
				continue
			}

			// 获取现有的缓存数据
			existingMetrics, found := s.metricsCache.GetPodMetrics(ns.Name, pod.Name)
			if found {
				// 保留历史数据
				metrics.HistoricalData = existingMetrics.HistoricalData

				// 添加当前数据点到历史数据中
				now := time.Now().Format(time.RFC3339)
				metrics.HistoricalData.CPUUsage = append(metrics.HistoricalData.CPUUsage, MetricDataPoint{
					Timestamp: now,
					Value:     metrics.CPUUsage,
				})

				metrics.HistoricalData.MemoryUsage = append(metrics.HistoricalData.MemoryUsage, MetricDataPoint{
					Timestamp: now,
					Value:     metrics.MemoryUsage,
				})

				metrics.HistoricalData.DiskUsage = append(metrics.HistoricalData.DiskUsage, MetricDataPoint{
					Timestamp: now,
					Value:     float64(metrics.DiskUsedBytes),
				})
			}

			// 更新缓存
			s.metricsCache.SetPodMetrics(ns.Name, pod.Name, metrics)
			successPods++
		}
	}

	// 记录完成收集指标的时间和统计数据
	logger.Debug("完成Pod指标数据收集",
		"cluster", clusterName,
		"duration", time.Since(startTime).String(),
		"totalPods", totalPods,
		"successPods", successPods)
}

// CleanExpiredMetricsCache 清理过期的指标缓存
func (s *PodMetricsService) CleanExpiredMetricsCache() {
	beforeCount := s.metricsCache.GetCacheSize()
	s.metricsCache.CleanExpired()
	afterCount := s.metricsCache.GetCacheSize()

	logger.Info("清理过期的Pod指标缓存",
		"beforeCount", beforeCount,
		"afterCount", afterCount,
		"removed", beforeCount-afterCount)
}
