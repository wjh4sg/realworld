package cache

import (
	"sync/atomic"
	"time"
)

// CacheMetricsCollector 缓存监控指标收集器
type CacheMetricsCollector struct {
	// L1 缓存指标
	l1HitCount   atomic.Int64
	l1MissCount  atomic.Int64
	
	// L2 缓存指标
	l2HitCount   atomic.Int64
	l2MissCount  atomic.Int64
	
	// 数据库指标
	dbQueryCount atomic.Int64
	
	// 延迟指标（简单实现，记录总延迟和次数）
	totalLatencyNs atomic.Int64
	latencyCount   atomic.Int64
}

// NewCacheMetricsCollector 创建缓存监控指标收集器
func NewCacheMetricsCollector() *CacheMetricsCollector {
	return &CacheMetricsCollector{}
}

// RecordL1Hit 记录 L1 缓存命中
func (c *CacheMetricsCollector) RecordL1Hit() {
	c.l1HitCount.Add(1)
}

// RecordL1Miss 记录 L1 缓存未命中
func (c *CacheMetricsCollector) RecordL1Miss() {
	c.l1MissCount.Add(1)
}

// RecordL2Hit 记录 L2 缓存命中
func (c *CacheMetricsCollector) RecordL2Hit() {
	c.l2HitCount.Add(1)
}

// RecordL2Miss 记录 L2 缓存未命中
func (c *CacheMetricsCollector) RecordL2Miss() {
	c.l2MissCount.Add(1)
}

// RecordDBQuery 记录数据库查询
func (c *CacheMetricsCollector) RecordDBQuery() {
	c.dbQueryCount.Add(1)
}

// RecordLatency 记录请求延迟（纳秒）
func (c *CacheMetricsCollector) RecordLatency(duration time.Duration) {
	c.totalLatencyNs.Add(int64(duration))
	c.latencyCount.Add(1)
}

// GetHitRate 获取总体缓存命中率
func (c *CacheMetricsCollector) GetHitRate() float64 {
	totalHits := c.l1HitCount.Load() + c.l2HitCount.Load()
	totalMisses := c.l1MissCount.Load() + c.l2MissCount.Load()
	total := totalHits + totalMisses
	
	if total == 0 {
		return 0
	}
	
	return float64(totalHits) / float64(total) * 100
}

// GetL1HitRate 获取 L1 缓存命中率
func (c *CacheMetricsCollector) GetL1HitRate() float64 {
	hits := c.l1HitCount.Load()
	misses := c.l1MissCount.Load()
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total) * 100
}

// GetL2HitRate 获取 L2 缓存命中率
func (c *CacheMetricsCollector) GetL2HitRate() float64 {
	hits := c.l2HitCount.Load()
	misses := c.l2MissCount.Load()
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total) * 100
}

// GetAvgLatency 获取平均延迟（毫秒）
func (c *CacheMetricsCollector) GetAvgLatency() float64 {
	totalNs := c.totalLatencyNs.Load()
	count := c.latencyCount.Load()
	
	if count == 0 {
		return 0
	}
	
	// 转换为毫秒
	return float64(totalNs) / float64(count) / 1e6
}

// ExportMetrics 导出所有指标
func (c *CacheMetricsCollector) ExportMetrics() map[string]interface{} {
	return map[string]interface{}{
		"l1_hit_count":    c.l1HitCount.Load(),
		"l1_miss_count":   c.l1MissCount.Load(),
		"l1_hit_rate":     c.GetL1HitRate(),
		"l2_hit_count":    c.l2HitCount.Load(),
		"l2_miss_count":   c.l2MissCount.Load(),
		"l2_hit_rate":     c.GetL2HitRate(),
		"total_hit_rate":  c.GetHitRate(),
		"db_query_count":  c.dbQueryCount.Load(),
		"avg_latency_ms":  c.GetAvgLatency(),
	}
}

// Reset 重置所有指标（用于测试）
func (c *CacheMetricsCollector) Reset() {
	c.l1HitCount.Store(0)
	c.l1MissCount.Store(0)
	c.l2HitCount.Store(0)
	c.l2MissCount.Store(0)
	c.dbQueryCount.Store(0)
	c.totalLatencyNs.Store(0)
	c.latencyCount.Store(0)
}
