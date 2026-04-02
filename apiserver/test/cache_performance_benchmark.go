package test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/onexstack/realworld/apiserver/cache"
)

// PerformanceResult 性能测试结果
type PerformanceResult struct {
	totalTime    time.Duration
	avgLatency   time.Duration
	p99Latency   time.Duration
	qps          float64
	successCount int
}

// runRedisOnlyTest 仅使用 Redis 缓存的测试
func runRedisOnlyTest(hitRate float64, numRequests, numConcurrent int) PerformanceResult {
	ctx := context.Background()

	// 使用内存模拟 Redis（避免依赖外部 Redis）
	mockRedis := newMockRedis()

	// 预热数据
	prewarmData(mockRedis, 100)

	// 开始测试
	start := time.Now()
	latencies := make([]time.Duration, 0, numRequests)
	latencyMutex := sync.Mutex{}

	// 并发控制
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, numConcurrent)

	successCount := 0

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			key := fmt.Sprintf("key:%d", idx%100)
			isHit := rand.Float64() < hitRate

			reqStart := time.Now()

			if isHit {
				// 缓存命中
				_, _ = mockRedis.Get(ctx, key).Result()
			} else {
				// 缓存未命中，模拟数据库查询
				time.Sleep(10 * time.Millisecond)
			}

			latency := time.Since(reqStart)
			latencyMutex.Lock()
			latencies = append(latencies, latency)
			successCount++
			latencyMutex.Unlock()
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	return calculateResult(latencies, totalTime, successCount)
}

// runLocalCacheTest 使用本地缓存 + Redis 缓存的测试
func runLocalCacheTest(hitRate float64, numRequests, numConcurrent int) PerformanceResult {
	ctx := context.Background()

	// 创建本地缓存
	ristrettoCache, _ := cache.NewRistrettoCache()
	mockRedis := newMockRedis()

	// 预热数据
	prewarmData(mockRedis, 100)

	// 开始测试
	start := time.Now()
	latencies := make([]time.Duration, 0, numRequests)
	latencyMutex := sync.Mutex{}

	// 并发控制
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, numConcurrent)

	successCount := 0

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			key := fmt.Sprintf("key:%d", idx%100)
			isHit := rand.Float64() < hitRate

			reqStart := time.Now()

			if isHit {
				// 先查本地缓存
				if _, ok := ristrettoCache.Get(key); ok {
					// 本地缓存命中
				} else {
					// 查 Redis
					val, _ := mockRedis.Get(ctx, key).Result()
					// 回填本地缓存
					ristrettoCache.Set(key, val, 1)
				}
			} else {
				// 缓存未命中，模拟数据库查询
				time.Sleep(10 * time.Millisecond)
			}

			latency := time.Since(reqStart)
			latencyMutex.Lock()
			latencies = append(latencies, latency)
			successCount++
			latencyMutex.Unlock()
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	return calculateResult(latencies, totalTime, successCount)
}

// calculateResult 计算性能结果
func calculateResult(latencies []time.Duration, totalTime time.Duration, successCount int) PerformanceResult {
	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	// 计算 P99 延迟（先排序）
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	p99Index := int(float64(len(latencies)) * 0.99)
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}
	p99Latency := latencies[p99Index]

	qps := float64(successCount) / totalTime.Seconds()

	return PerformanceResult{
		totalTime:    totalTime,
		avgLatency:   avgLatency,
		p99Latency:   p99Latency,
		qps:          qps,
		successCount: successCount,
	}
}

// printPerformanceComparison 打印性能对比
func printPerformanceComparison(redisResult, localResult PerformanceResult) {
	fmt.Println("┌─────────────────┬──────────────────┬──────────────────┬──────────────────┐")
	fmt.Println("│     指标        │   仅 Redis 缓存  │ 本地+Redis 缓存  │     性能提升     │")
	fmt.Println("├─────────────────┼──────────────────┼──────────────────┼──────────────────┤")

	// QPS
	redisQPS := fmt.Sprintf("%.0f", redisResult.qps)
	localQPS := fmt.Sprintf("%.0f", localResult.qps)
	qpsImprovement := 0.0
	if redisResult.qps > 0 {
		qpsImprovement = (localResult.qps / redisResult.qps) * 100
	}
	qpsImprovementStr := fmt.Sprintf("%.1f%%", qpsImprovement)
	fmt.Printf("│      QPS        │ %16s │ %16s │ %16s │\n", redisQPS, localQPS, qpsImprovementStr)

	// 平均延迟
	redisAvg := redisResult.avgLatency.String()
	localAvg := localResult.avgLatency.String()
	latencyImprovement := 0.0
	if localResult.avgLatency > 0 {
		latencyImprovement = (redisResult.avgLatency.Seconds() / localResult.avgLatency.Seconds()) * 100
	}
	latencyImprovementStr := fmt.Sprintf("%.1f%%", latencyImprovement)
	fmt.Printf("│   平均延迟      │ %16s │ %16s │ %16s │\n", redisAvg, localAvg, latencyImprovementStr)

	// P99 延迟
	redisP99 := redisResult.p99Latency.String()
	localP99 := localResult.p99Latency.String()
	p99Improvement := 0.0
	if localResult.p99Latency > 0 {
		p99Improvement = (redisResult.p99Latency.Seconds() / localResult.p99Latency.Seconds()) * 100
	}
	p99ImprovementStr := fmt.Sprintf("%.1f%%", p99Improvement)
	fmt.Printf("│   P99 延迟      │ %16s │ %16s │ %16s │\n", redisP99, localP99, p99ImprovementStr)

	// 总耗时
	redisTotal := redisResult.totalTime.String()
	localTotal := localResult.totalTime.String()
	fmt.Printf("│   总耗时        │ %16s │ %16s │                  │\n", redisTotal, localTotal)

	fmt.Println("└─────────────────┴──────────────────┴──────────────────┴──────────────────┘")
}

// mockRedis 模拟 Redis 客户端
type mockRedis struct {
	data map[string]string
	mu   sync.RWMutex
}

func newMockRedis() *mockRedis {
	return &mockRedis{
		data: make(map[string]string),
	}
}

func (m *mockRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		cmd.SetVal(val)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (m *mockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = fmt.Sprintf("%v", value)
	cmd.SetVal("OK")
	return cmd
}

func (m *mockRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			count++
		}
	}
	cmd.SetVal(int64(count))
	return cmd
}

// prewarmData 预热数据
func prewarmData(redis *mockRedis, count int) {
	ctx := context.Background()
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := fmt.Sprintf("value:%d", i)
		redis.Set(ctx, key, value, 0)
	}
}

// main 主函数
func main() {
	fmt.Println("=== 缓存性能对比测试 ===")
	fmt.Println()

	// 测试配置
	testCases := []struct {
		name          string
		cacheHitRate  float64
		numRequests   int
		numConcurrent int
	}{
		{
			name:          "100% 缓存命中 - 1000 请求",
			cacheHitRate:  1.0,
			numRequests:   1000,
			numConcurrent: 1,
		},
		{
			name:          "80% 缓存命中 - 1000 请求",
			cacheHitRate:  0.8,
			numRequests:   1000,
			numConcurrent: 1,
		},
		{
			name:          "50% 缓存命中 - 1000 请求",
			cacheHitRate:  0.5,
			numRequests:   1000,
			numConcurrent: 1,
		},
		{
			name:          "100% 缓存命中 - 高并发 10000 请求",
			cacheHitRate:  1.0,
			numRequests:   10000,
			numConcurrent: 10,
		},
	}

	// 运行所有测试用例
	for _, tc := range testCases {
		fmt.Printf("--- 测试场景: %s ---\n", tc.name)
		fmt.Printf("缓存命中率: %.0f%%, 请求数: %d, 并发数: %d\n",
			tc.cacheHitRate*100, tc.numRequests, tc.numConcurrent)
		fmt.Println()

		// 测试仅使用 Redis 缓存
		redisResult := runRedisOnlyTest(tc.cacheHitRate, tc.numRequests, tc.numConcurrent)

		// 测试使用本地缓存 + Redis 缓存
		localResult := runLocalCacheTest(tc.cacheHitRate, tc.numRequests, tc.numConcurrent)

		// 输出对比结果
		printPerformanceComparison(redisResult, localResult)
		fmt.Println()
	}
}
