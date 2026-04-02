//go:build perf

package test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestConcurrencyPerformance(t *testing.T) {
	baseURL := testBaseURL()

	testCases := []struct {
		name            string
		concurrentUsers int
		requestsPerUser int
		expectedMinQPS  int
	}{
		{name: "low-concurrency", concurrentUsers: 10, requestsPerUser: 100, expectedMinQPS: 500},
		{name: "medium-concurrency", concurrentUsers: 50, requestsPerUser: 100, expectedMinQPS: 1000},
		{name: "high-concurrency", concurrentUsers: 100, requestsPerUser: 100, expectedMinQPS: 1500},
		{name: "extreme-concurrency", concurrentUsers: 200, requestsPerUser: 100, expectedMinQPS: 2000},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := runPerformanceTest(baseURL, tc.concurrentUsers, tc.requestsPerUser)
			if result.successCount == 0 {
				t.Fatalf("all requests failed")
			}
			if result.qps < float64(tc.expectedMinQPS) {
				t.Fatalf("qps below expectation: %.2f < %d", result.qps, tc.expectedMinQPS)
			}

			t.Logf("total=%s success=%d failure=%d qps=%.2f avg=%s min=%s max=%s",
				result.totalTime,
				result.successCount,
				result.failureCount,
				result.qps,
				result.avgResponseTime,
				result.minResponseTime,
				result.maxResponseTime,
			)
		})
	}
}

type PerformanceTestResult struct {
	totalTime       time.Duration
	successCount    int
	failureCount    int
	qps             float64
	avgResponseTime time.Duration
	minResponseTime time.Duration
	maxResponseTime time.Duration
}

func runPerformanceTest(baseURL string, concurrentUsers, requestsPerUser int) *PerformanceTestResult {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	result := &PerformanceTestResult{
		minResponseTime: time.Hour,
	}

	startTime := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			client := &http.Client{Timeout: 30 * time.Second}
			for j := 0; j < requestsPerUser; j++ {
				reqStart := time.Now()
				resp, err := client.Get(baseURL + "/api/articles")
				respTime := time.Since(reqStart)

				mu.Lock()
				if respTime < result.minResponseTime {
					result.minResponseTime = respTime
				}
				if respTime > result.maxResponseTime {
					result.maxResponseTime = respTime
				}

				if err != nil || resp.StatusCode != http.StatusOK {
					result.failureCount++
				} else {
					result.successCount++
					result.avgResponseTime += respTime
				}
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}

				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	result.totalTime = time.Since(startTime)

	if result.successCount > 0 {
		totalRequests := result.successCount + result.failureCount
		result.qps = float64(totalRequests) / result.totalTime.Seconds()
		result.avgResponseTime /= time.Duration(result.successCount)
	}

	fmt.Printf("performance result: %+v\n", result)
	return result
}
