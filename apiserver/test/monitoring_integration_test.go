//go:build integration

package test

import (
	"net/http"
	"testing"
)

func TestMonitoringEndpoints(t *testing.T) {
	baseURL := testBaseURL()

	endpoints := []string{
		"/healthz",
		"/readyz",
		"/metrics/concurrency",
		"/metrics",
	}

	for _, endpoint := range endpoints {
		endpoint := endpoint
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(baseURL + endpoint)
			if err != nil {
				t.Fatalf("failed to request %s: %v", endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status for %s: %d", endpoint, resp.StatusCode)
			}
		})
	}
}
