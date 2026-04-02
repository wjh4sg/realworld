package monitoring

import (
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ConcurrencyMonitor 并发监控器
type ConcurrencyMonitor struct {
	mu                sync.Mutex
	goroutineCount    *expvar.Int
	memoryAlloc       *expvar.Int
	memoryTotalAlloc  *expvar.Int
	memorySys         *expvar.Int
	lockContention    *expvar.Int
	lastGC            *expvar.Int
	startTime         time.Time
}

// NewConcurrencyMonitor 创建并发监控器
func NewConcurrencyMonitor() *ConcurrencyMonitor {
	monitor := &ConcurrencyMonitor{
		goroutineCount:   expvar.NewInt("goroutine_count"),
		memoryAlloc:      expvar.NewInt("memory_alloc"),
		memoryTotalAlloc: expvar.NewInt("memory_total_alloc"),
		memorySys:        expvar.NewInt("memory_sys"),
		lockContention:   expvar.NewInt("lock_contention"),
		lastGC:           expvar.NewInt("last_gc"),
		startTime:        time.Now(),
	}

	// 启动监控收集器
	go monitor.startCollector()

	return monitor
}

// startCollector 启动监控收集器
func (m *ConcurrencyMonitor) startCollector() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.collectMetrics()
	}
}

// collectMetrics 收集监控指标
func (m *ConcurrencyMonitor) collectMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 收集goroutine数量
	m.goroutineCount.Set(int64(runtime.NumGoroutine()))

	// 收集内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.memoryAlloc.Set(int64(memStats.Alloc))
	m.memoryTotalAlloc.Set(int64(memStats.TotalAlloc))
	m.memorySys.Set(int64(memStats.Sys))

	// 收集GC信息
	if memStats.LastGC != 0 {
		m.lastGC.Set(int64(memStats.LastGC / uint64(time.Millisecond)))
	}

	// 收集锁竞争信息
	// 注意：在Go 1.18+中，可以使用runtime/metrics包获取更详细的锁竞争信息
}

// RegisterHTTPHandler 注册HTTP处理函数（已废弃，现在在main.go中直接注册）
func (m *ConcurrencyMonitor) RegisterHTTPHandler() {
	http.HandleFunc("/metrics/concurrency", m.HandleMetrics)
	http.HandleFunc("/metrics", m.HandleAllMetrics)
}

// HandleMetrics 处理并发监控指标请求
func (m *ConcurrencyMonitor) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "# Concurrency Metrics\n")
	fmt.Fprintf(w, "# Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(w, "# Uptime: %s\n", time.Since(m.startTime))
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "goroutine_count %d\n", m.goroutineCount.Value())
	fmt.Fprintf(w, "memory_alloc %d\n", m.memoryAlloc.Value())
	fmt.Fprintf(w, "memory_total_alloc %d\n", m.memoryTotalAlloc.Value())
	fmt.Fprintf(w, "memory_sys %d\n", m.memorySys.Value())
	fmt.Fprintf(w, "lock_contention %d\n", m.lockContention.Value())
	fmt.Fprintf(w, "last_gc %d\n", m.lastGC.Value())

	// 添加Go版本信息
	fmt.Fprintf(w, "\n# Go Version: %s\n", runtime.Version())
	fmt.Fprintf(w, "# Num CPU: %d\n", runtime.NumCPU())
}

// HandleAllMetrics 处理所有监控指标请求
func (m *ConcurrencyMonitor) HandleAllMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	expvar.Handler().ServeHTTP(w, r)
}

// GetMetrics 获取当前监控指标
func (m *ConcurrencyMonitor) GetMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"goroutine_count":   m.goroutineCount.Value(),
		"memory_alloc":      m.memoryAlloc.Value(),
		"memory_total_alloc": m.memoryTotalAlloc.Value(),
		"memory_sys":        m.memorySys.Value(),
		"lock_contention":   m.lockContention.Value(),
		"last_gc":           m.lastGC.Value(),
		"uptime":           time.Since(m.startTime).String(),
		"go_version":        runtime.Version(),
		"num_cpu":           runtime.NumCPU(),
	}
}
