package handler

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

// AsyncHandler 异步处理函数类型
type AsyncHandler func(c *gin.Context) error

// Async 异步处理请求的包装器
func Async(handler AsyncHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 克隆gin上下文，确保goroutine中可以安全使用
		cCopy := c.Copy()
		
		// 使用goroutine异步处理
		go func() {
			if err := handler(cCopy); err != nil {
				log.Printf("Async handler error: %v", err)
			}
		}()
	}
}

// Parallel 并行处理多个任务
func Parallel(tasks ...func() error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	
	for _, task := range tasks {
		wg.Add(1)
		go func(t func() error) {
			defer wg.Done()
			if err := t(); err != nil {
				errChan <- err
			}
		}(task)
	}
	
	wg.Wait()
	close(errChan)
	
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}
