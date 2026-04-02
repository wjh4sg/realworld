package middleware

import (
	"testing"
	"time"
)

func TestFixedWindowLimiterEnforcesMinuteAndHourLimits(t *testing.T) {
	limiter := newFixedWindowLimiter(2, 3)
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	if !limiter.Allow("token:test", now) {
		t.Fatalf("expected first request to pass")
	}
	if !limiter.Allow("token:test", now.Add(10*time.Second)) {
		t.Fatalf("expected second request in same minute to pass")
	}
	if limiter.Allow("token:test", now.Add(20*time.Second)) {
		t.Fatalf("expected third request in same minute to be rate limited")
	}

	if !limiter.Allow("token:test", now.Add(1*time.Minute)) {
		t.Fatalf("expected minute window reset to allow request")
	}
	if limiter.Allow("token:test", now.Add(2*time.Minute)) {
		t.Fatalf("expected hour limit to block fourth request in same hour")
	}
	if !limiter.Allow("token:test", now.Add(1*time.Hour)) {
		t.Fatalf("expected hour window reset to allow request")
	}
}
