// Package k8s 提供了与Kubernetes集群交互的功能
package k8s

import (
	"sync"
	"time"
)

// MetricsCache 用于存储Pod指标数据的缓存
type MetricsCache struct {
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
}

// NewMetricsCache 创建一个新的指标缓存
func NewMetricsCache(ttl time.Duration) *MetricsCache {
	return &MetricsCache{
		metricsCache:  make(map[string]*PodMetrics),
		resourceCache: make(map[string]*PodResourceUsage),
		lastUpdated:   make(map[string]time.Time),
		ttl:           ttl,
	}
}

// GetPodMetrics 从缓存中获取Pod指标数据
func (c *MetricsCache) GetPodMetrics(namespace, podName string) (*PodMetrics, bool) {
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

	return metrics, true
}

// SetPodMetrics 将Pod指标数据存入缓存
func (c *MetricsCache) SetPodMetrics(namespace, podName string, metrics *PodMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := namespace + "/" + podName
	c.metricsCache[key] = metrics
	c.lastUpdated[key] = time.Now()
}

// GetPodResourceUsage 从缓存中获取Pod资源使用情况
func (c *MetricsCache) GetPodResourceUsage(namespace, podName string) (*PodResourceUsage, bool) {
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

	return usage, true
}

// SetPodResourceUsage 将Pod资源使用情况存入缓存
func (c *MetricsCache) SetPodResourceUsage(namespace, podName string, usage *PodResourceUsage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := namespace + "/" + podName
	c.resourceCache[key] = usage
	c.lastUpdated[key] = time.Now()
}

// Clear 清除所有缓存数据
func (c *MetricsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metricsCache = make(map[string]*PodMetrics)
	c.resourceCache = make(map[string]*PodResourceUsage)
	c.lastUpdated = make(map[string]time.Time)
}

// CleanExpired 清除过期的缓存数据
func (c *MetricsCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, updated := range c.lastUpdated {
		if now.Sub(updated) > c.ttl {
			delete(c.metricsCache, key)
			delete(c.resourceCache, key)
			delete(c.lastUpdated, key)
		}
	}
}
