// Package k8s 提供了与Kubernetes集群交互的功能
package k8s

import (
	"container/list"
	"sync"
	"time"
)

// MaxCacheSize 默认的最大缓存条目数
const DefaultMaxCacheSize = 1000

// DataAggregationInterval 数据聚合的时间间隔
const DefaultAggregationInterval = 1 * time.Hour

// MemoryMetricsCache 是一个基于内存的Pod指标缓存实现
// 支持LRU淘汰策略、内存大小限制和数据聚合
type MemoryMetricsCache struct {
	// 使用互斥锁保护缓存数据的并发访问
	mu sync.RWMutex

	// 存储Pod指标数据的映射：key为"namespace/podName"，value为PodMetrics
	metricsCache map[string]*PodMetrics

	// 存储资源使用数据的映射：key为"namespace/podName"，value为PodResourceUsage
	resourceCache map[string]*PodResourceUsage

	// 缓存有效期
	ttl time.Duration

	// 最后更新时间的映射：key为"namespace/podName"，value为更新时间
	lastUpdated map[string]time.Time

	// 最后访问时间的映射：key为"namespace/podName"，value为访问时间
	lastAccessed map[string]time.Time

	// LRU列表，用于实现最近最少使用淘汰策略
	lruList *list.List

	// LRU映射，用于快速查找列表中的元素
	lruMap map[string]*list.Element

	// 最大缓存条目数
	maxCacheSize int

	// 数据聚合时间间隔
	aggregationInterval time.Duration

	// 最后一次聚合的时间
	lastAggregation time.Time
}

// LRUCacheItem LRU缓存项目
type LRUCacheItem struct {
	Key string
}

// NewMemoryMetricsCache 创建一个新的基于内存的指标缓存
func NewMemoryMetricsCache(ttl time.Duration, maxCacheSize int, aggregationInterval time.Duration) *MemoryMetricsCache {
	if maxCacheSize <= 0 {
		maxCacheSize = DefaultMaxCacheSize
	}

	if aggregationInterval <= 0 {
		aggregationInterval = DefaultAggregationInterval
	}

	return &MemoryMetricsCache{
		metricsCache:        make(map[string]*PodMetrics),
		resourceCache:       make(map[string]*PodResourceUsage),
		lastUpdated:         make(map[string]time.Time),
		lastAccessed:        make(map[string]time.Time),
		lruList:             list.New(),
		lruMap:              make(map[string]*list.Element),
		ttl:                 ttl,
		maxCacheSize:        maxCacheSize,
		aggregationInterval: aggregationInterval,
		lastAggregation:     time.Now(),
	}
}

// GetPodMetrics 从缓存中获取Pod指标数据
func (c *MemoryMetricsCache) GetPodMetrics(namespace, podName string) (*PodMetrics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := namespace + "/" + podName
	metrics, exists := c.metricsCache[key]
	if !exists {
		return nil, false
	}

	// 检查缓存是否过期
	if time.Since(c.lastUpdated[key]) > c.ttl {
		return nil, false
	}

	// 更新访问时间并调整LRU位置
	c.updateAccess(key)

	return metrics, true
}

// SetPodMetrics 将Pod指标数据存入缓存
func (c *MemoryMetricsCache) SetPodMetrics(namespace, podName string, metrics *PodMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := namespace + "/" + podName

	// 检查是否需要进行数据聚合
	c.aggregateDataIfNeeded()

	// 如果达到最大缓存大小且该键不存在于缓存中，则淘汰最久未使用的条目
	if len(c.metricsCache) >= c.maxCacheSize && c.metricsCache[key] == nil {
		c.evictLRU()
	}

	// 将当前数据点添加到历史数据中
	now := time.Now().Format(time.RFC3339)

	// 获取现有的指标数据（如果有）
	existingMetrics, exists := c.metricsCache[key]

	if exists && existingMetrics != nil {
		// 将现有历史数据复制到新的metrics对象中
		metrics.HistoricalData.CPUUsage = append([]MetricDataPoint{}, existingMetrics.HistoricalData.CPUUsage...)
		metrics.HistoricalData.MemoryUsage = append([]MetricDataPoint{}, existingMetrics.HistoricalData.MemoryUsage...)
		metrics.HistoricalData.DiskUsage = append([]MetricDataPoint{}, existingMetrics.HistoricalData.DiskUsage...)
	}

	// 添加当前数据点到历史数据中
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
		Value:     metrics.DiskUsage,
	})

	// 更新缓存
	c.metricsCache[key] = metrics
	c.lastUpdated[key] = time.Now()

	// 更新LRU信息
	c.updateLRU(key)
}

// GetPodResourceUsage 从缓存中获取Pod资源使用情况
func (c *MemoryMetricsCache) GetPodResourceUsage(namespace, podName string) (*PodResourceUsage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := namespace + "/" + podName
	usage, exists := c.resourceCache[key]
	if !exists {
		return nil, false
	}

	// 检查缓存是否过期
	if time.Since(c.lastUpdated[key]) > c.ttl {
		return nil, false
	}

	// 更新访问时间并调整LRU位置
	c.updateAccess(key)

	return usage, true
}

// SetPodResourceUsage 将Pod资源使用情况存入缓存
func (c *MemoryMetricsCache) SetPodResourceUsage(namespace, podName string, usage *PodResourceUsage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := namespace + "/" + podName

	// 检查是否需要进行数据聚合
	c.aggregateDataIfNeeded()

	// 如果达到最大缓存大小且该键不存在于缓存中，则淘汰最久未使用的条目
	if len(c.resourceCache) >= c.maxCacheSize && c.resourceCache[key] == nil {
		c.evictLRU()
	}

	// 更新缓存
	c.resourceCache[key] = usage
	c.lastUpdated[key] = time.Now()

	// 更新LRU信息
	c.updateLRU(key)
}

// Clear 清除所有缓存数据
func (c *MemoryMetricsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metricsCache = make(map[string]*PodMetrics)
	c.resourceCache = make(map[string]*PodResourceUsage)
	c.lastUpdated = make(map[string]time.Time)
	c.lastAccessed = make(map[string]time.Time)
	c.lruList = list.New()
	c.lruMap = make(map[string]*list.Element)
}

// CleanExpired 清除过期的缓存数据
func (c *MemoryMetricsCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, updatedAt := range c.lastUpdated {
		if now.Sub(updatedAt) > c.ttl {
			// 删除过期的缓存条目
			delete(c.metricsCache, key)
			delete(c.resourceCache, key)
			delete(c.lastUpdated, key)
			delete(c.lastAccessed, key)

			// 从LRU列表中删除
			if element, exists := c.lruMap[key]; exists {
				c.lruList.Remove(element)
				delete(c.lruMap, key)
			}
		}
	}
}

// updateLRU 更新LRU列表
func (c *MemoryMetricsCache) updateLRU(key string) {
	// 已经存在，则移到列表前端
	if element, exists := c.lruMap[key]; exists {
		c.lruList.MoveToFront(element)
		c.lastAccessed[key] = time.Now()
		return
	}

	// 不存在，则添加到列表前端
	element := c.lruList.PushFront(&LRUCacheItem{Key: key})
	c.lruMap[key] = element
	c.lastAccessed[key] = time.Now()
}

// updateAccess 更新访问时间（在读锁保护下调用，只是记录逻辑调用时间）
func (c *MemoryMetricsCache) updateAccess(key string) {
	// 这里不需要更新LRU位置，因为GetXXX方法已经在读锁中
	// 我们只需要在下一次写操作时更新LRU
	c.lastAccessed[key] = time.Now()
}

// evictLRU 淘汰最近最少使用的缓存条目
func (c *MemoryMetricsCache) evictLRU() {
	if c.lruList.Len() == 0 {
		return
	}

	// 获取最后一个元素（最久未使用的）
	element := c.lruList.Back()
	if element == nil {
		return
	}

	// 从LRU列表中删除
	c.lruList.Remove(element)

	// 获取键并从缓存中删除
	item := element.Value.(*LRUCacheItem)
	key := item.Key

	delete(c.metricsCache, key)
	delete(c.resourceCache, key)
	delete(c.lastUpdated, key)
	delete(c.lastAccessed, key)
	delete(c.lruMap, key)
}

// aggregateDataIfNeeded 检查是否需要进行数据聚合，如果需要则执行聚合
func (c *MemoryMetricsCache) aggregateDataIfNeeded() {
	// 检查是否到达聚合间隔
	if time.Since(c.lastAggregation) < c.aggregationInterval {
		return
	}

	// 更新聚合时间
	c.lastAggregation = time.Now()

	// 聚合所有缓存的指标数据
	for key, metrics := range c.metricsCache {
		c.aggregateMetricsData(key, metrics)
	}

	// 聚合所有缓存的资源使用数据
	for key, usage := range c.resourceCache {
		c.aggregateResourceUsageData(key, usage)
	}
}

// aggregateMetricsData 聚合Pod指标数据
func (c *MemoryMetricsCache) aggregateMetricsData(key string, metrics *PodMetrics) {
	// 跳过没有历史数据的项
	if metrics == nil {
		return
	}

	// 聚合CPU历史数据
	if len(metrics.HistoricalData.CPUUsage) > 0 {
		metrics.HistoricalData.CPUUsage = c.aggregateTimeSeriesData(metrics.HistoricalData.CPUUsage)
	}

	// 聚合内存历史数据
	if len(metrics.HistoricalData.MemoryUsage) > 0 {
		metrics.HistoricalData.MemoryUsage = c.aggregateTimeSeriesData(metrics.HistoricalData.MemoryUsage)
	}

	// 聚合磁盘历史数据
	if len(metrics.HistoricalData.DiskUsage) > 0 {
		metrics.HistoricalData.DiskUsage = c.aggregateTimeSeriesData(metrics.HistoricalData.DiskUsage)
	}
}

// aggregateResourceUsageData 聚合Pod资源使用数据
func (c *MemoryMetricsCache) aggregateResourceUsageData(key string, usage *PodResourceUsage) {
	// 跳过没有历史数据的项
	if usage == nil || len(usage.Historical) == 0 {
		return
	}

	// 遍历所有指标类型
	for metricType, dataPoints := range usage.Historical {
		if len(dataPoints) > 0 {
			// 聚合数据点
			usage.Historical[metricType] = c.aggregateResourceDataPoints(dataPoints)
		}
	}
}

// aggregateTimeSeriesData 聚合时间序列数据
// 聚合策略：
// 1. 保留最近24小时的数据点不变
// 2. 24小时-7天的数据按小时聚合
// 3. 7天以上的数据按天聚合
func (c *MemoryMetricsCache) aggregateTimeSeriesData(dataPoints []MetricDataPoint) []MetricDataPoint {
	// 如果数据点太少，不需要聚合
	if len(dataPoints) <= 24 {
		return dataPoints
	}

	now := time.Now()
	var recentPoints []MetricDataPoint                  // 最近24小时的数据点
	hourlyBuckets := make(map[string][]MetricDataPoint) // 24小时-7天按小时分组的数据点
	dailyBuckets := make(map[string][]MetricDataPoint)  // 7天以上按天分组的数据点

	// 遍历所有数据点
	for _, point := range dataPoints {
		// 解析时间戳
		timestamp, err := time.Parse(time.RFC3339, point.Timestamp)
		if err != nil {
			// 如果解析失败，保留原始数据点
			recentPoints = append(recentPoints, point)
			continue
		}

		// 计算时间差
		duration := now.Sub(timestamp)

		// 保留最近24小时的数据点
		if duration <= 24*time.Hour {
			recentPoints = append(recentPoints, point)
		} else if duration <= 7*24*time.Hour {
			// 对24小时-7天的数据点按小时分组
			hourKey := timestamp.Format("2006-01-02-15") // 按小时分组
			hourlyBuckets[hourKey] = append(hourlyBuckets[hourKey], point)
		} else {
			// 对7天以上的数据点按天分组
			dayKey := timestamp.Format("2006-01-02") // 按天分组
			dailyBuckets[dayKey] = append(dailyBuckets[dayKey], point)
		}
	}

	// 聚合24小时-7天的数据点（按小时）
	var hourlyAggregatedPoints []MetricDataPoint
	for _, points := range hourlyBuckets {
		if len(points) == 0 {
			continue
		}

		// 计算平均值
		var sum float64
		for _, point := range points {
			sum += point.Value
		}
		avg := sum / float64(len(points))

		// 使用该小时第一个数据点的时间戳
		timestamp := points[0].Timestamp

		// 添加聚合后的数据点
		hourlyAggregatedPoints = append(hourlyAggregatedPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     avg,
		})
	}

	// 聚合7天以上的数据点（按天）
	var dailyAggregatedPoints []MetricDataPoint
	for _, points := range dailyBuckets {
		if len(points) == 0 {
			continue
		}

		// 计算平均值
		var sum float64
		for _, point := range points {
			sum += point.Value
		}
		avg := sum / float64(len(points))

		// 使用该天第一个数据点的时间戳
		timestamp := points[0].Timestamp

		// 添加聚合后的数据点
		dailyAggregatedPoints = append(dailyAggregatedPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     avg,
		})
	}

	// 合并所有聚合后的数据点和最近24小时的数据点
	result := append(dailyAggregatedPoints, hourlyAggregatedPoints...)
	result = append(result, recentPoints...)

	return result
}

// aggregateResourceDataPoints 聚合资源使用数据点
func (c *MemoryMetricsCache) aggregateResourceDataPoints(dataPoints []ResourceDataPoint) []ResourceDataPoint {
	// 如果数据点太少，不需要聚合
	if len(dataPoints) <= 24 {
		return dataPoints
	}

	now := time.Now()
	var recentPoints []ResourceDataPoint                  // 最近24小时的数据点
	hourlyBuckets := make(map[string][]ResourceDataPoint) // 按小时分组的数据点

	// 遍历所有数据点
	for _, point := range dataPoints {
		// 计算时间差
		duration := now.Sub(point.Timestamp)

		// 保留最近24小时的数据点
		if duration <= 24*time.Hour {
			recentPoints = append(recentPoints, point)
		} else {
			// 对24小时前的数据点按小时分组
			hourKey := point.Timestamp.Format("2006-01-02-15") // 按 order by
			hourlyBuckets[hourKey] = append(hourlyBuckets[hourKey], point)
		}
	}

	// 聚合24小时前的数据点
	var aggregatedPoints []ResourceDataPoint
	for _, points := range hourlyBuckets {
		if len(points) == 0 {
			continue
		}

		// 计算平均值
		var sum float64
		for _, point := range points {
			sum += point.Value
		}
		avg := sum / float64(len(points))

		// 使用该小时第一个数据点的时间戳
		timestamp := points[0].Timestamp

		// 添加聚合后的数据点
		aggregatedPoints = append(aggregatedPoints, ResourceDataPoint{
			Timestamp: timestamp,
			Value:     avg,
		})
	}

	// 合并聚合后的数据点和最近24小时的数据点
	result := append(aggregatedPoints, recentPoints...)

	return result
}

// GetCacheSize 获取当前缓存大小
func (c *MemoryMetricsCache) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.metricsCache)
}

// GetMaxCacheSize 获取最大缓存大小
func (c *MemoryMetricsCache) GetMaxCacheSize() int {
	return c.maxCacheSize
}

// SetMaxCacheSize 设置最大缓存大小
func (c *MemoryMetricsCache) SetMaxCacheSize(size int) {
	if size <= 0 {
		size = DefaultMaxCacheSize
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	oldSize := c.maxCacheSize
	c.maxCacheSize = size

	// 如果新的大小小于旧的大小，需要淘汰多余的条目
	if size < oldSize && len(c.metricsCache) > size {
		// 需要淘汰的条目数
		toEvict := len(c.metricsCache) - size
		for i := 0; i < toEvict; i++ {
			c.evictLRU()
		}
	}
}

// SetAggregationInterval 设置数据聚合间隔
func (c *MemoryMetricsCache) SetAggregationInterval(interval time.Duration) {
	if interval <= 0 {
		interval = DefaultAggregationInterval
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.aggregationInterval = interval
}

// GetAggregationInterval 获取数据聚合间隔
func (c *MemoryMetricsCache) GetAggregationInterval() time.Duration {
	return c.aggregationInterval
}

// SaveToStorage 将缓存数据保存到持久存储（接口预留，暂不实现）
func (c *MemoryMetricsCache) SaveToStorage(storagePath string) error {
	// 这是一个预留的接口，用于将内存数据持久化到磁盘
	// 暂时返回nil，表示不执行任何操作
	return nil
}

// LoadFromStorage 从持久存储加载缓存数据（接口预留，暂不实现）
func (c *MemoryMetricsCache) LoadFromStorage(storagePath string) error {
	// 这是一个预留的接口，用于从磁盘加载持久化的数据
	// 暂时返回nil，表示不执行任何操作
	return nil
}
