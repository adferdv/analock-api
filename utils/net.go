package utils

import (
	"net/http"
	"time"
)

// Singleton instance of custom HTTP client
var defaultHttpClientInstance *http.Client

// Gets the singleton custom HTTP client
func GetDefaultHttpClient() *http.Client {
	if defaultHttpClientInstance == nil {
		defaultHttpClientInstance = buildHttpClient(30 * time.Second)
	}
	return defaultHttpClientInstance
}

// It gets a new custom HTTP client with custom timeout
func GetCustomHttpClient(timeout time.Duration) *http.Client {
	return buildHttpClient(timeout)
}

// Builds a custom HTTP client
//
// It tweaks some settings like response timeout and max connections to get better performance.
func buildHttpClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxConnsPerHost = 100
	transport.MaxIdleConnsPerHost = 100

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
