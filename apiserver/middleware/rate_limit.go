package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientRateWindow struct {
	minuteStart time.Time
	hourStart   time.Time
	minuteCount int
	hourCount   int
	lastSeen    time.Time
}

type fixedWindowLimiter struct {
	perMinute   int
	perHour     int
	clients     map[string]*clientRateWindow
	lastCleanup time.Time
	mutex       sync.Mutex
}

// RateLimiter enforces API defaults from the contract: 60 requests/minute and 1000 requests/hour.
// Authenticated requests are keyed by token, while anonymous requests fall back to client IP.
func RateLimiter(perMinute, perHour int) gin.HandlerFunc {
	limiter := newFixedWindowLimiter(perMinute, perHour)

	return func(c *gin.Context) {
		if !limiter.Allow(rateLimitKey(c), time.Now()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func newFixedWindowLimiter(perMinute, perHour int) *fixedWindowLimiter {
	if perMinute <= 0 {
		perMinute = 60
	}
	if perHour <= 0 {
		perHour = 1000
	}

	return &fixedWindowLimiter{
		perMinute: perMinute,
		perHour:   perHour,
		clients:   make(map[string]*clientRateWindow),
	}
}

func (l *fixedWindowLimiter) Allow(key string, now time.Time) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.cleanupLocked(now)

	window, exists := l.clients[key]
	if !exists {
		window = &clientRateWindow{
			minuteStart: now.Truncate(time.Minute),
			hourStart:   now.Truncate(time.Hour),
		}
		l.clients[key] = window
	}

	currentMinute := now.Truncate(time.Minute)
	currentHour := now.Truncate(time.Hour)

	if !window.minuteStart.Equal(currentMinute) {
		window.minuteStart = currentMinute
		window.minuteCount = 0
	}

	if !window.hourStart.Equal(currentHour) {
		window.hourStart = currentHour
		window.hourCount = 0
	}

	if window.minuteCount >= l.perMinute || window.hourCount >= l.perHour {
		window.lastSeen = now
		return false
	}

	window.minuteCount++
	window.hourCount++
	window.lastSeen = now
	return true
}

func (l *fixedWindowLimiter) cleanupLocked(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < 5*time.Minute {
		return
	}

	for key, window := range l.clients {
		if now.Sub(window.lastSeen) > 2*time.Hour {
			delete(l.clients, key)
		}
	}

	l.lastCleanup = now
}

func rateLimitKey(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	if authHeader := strings.TrimSpace(c.GetHeader("Authorization")); authHeader != "" {
		return "token:" + authHeader
	}

	return "ip:" + c.ClientIP()
}
