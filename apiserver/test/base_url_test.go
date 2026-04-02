package test

import "os"

func testBaseURL() string {
	if baseURL := os.Getenv("API_BASE_URL"); baseURL != "" {
		return baseURL
	}

	return "http://localhost:18080"
}
