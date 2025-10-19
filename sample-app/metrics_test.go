//go:build unit
// +build unit

/**
# Copyright 2015 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestMetrics_NewMetrics tests metrics creation
func TestMetrics_NewMetrics(t *testing.T) {
	metrics := NewMetrics()

	if metrics == nil {
		t.Error("NewMetrics should return a non-nil metrics instance")
	}

	if metrics.RequestCount != 0 {
		t.Error("New metrics should have zero request count")
	}

	if metrics.ErrorCount != 0 {
		t.Error("New metrics should have zero error count")
	}

	if metrics.StatusCodes == nil {
		t.Error("StatusCodes map should be initialized")
	}

	if metrics.Endpoints == nil {
		t.Error("Endpoints map should be initialized")
	}

	if metrics.Errors == nil {
		t.Error("Errors map should be initialized")
	}
}

// TestMetrics_RecordRequest tests request recording
func TestMetrics_RecordRequest(t *testing.T) {
	metrics := NewMetrics()

	// Record a successful request
	metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)

	if metrics.RequestCount != 1 {
		t.Errorf("Expected request count 1, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", metrics.ErrorCount)
	}

	if metrics.StatusCodes[200] != 1 {
		t.Errorf("Expected status code 200 count 1, got %d", metrics.StatusCodes[200])
	}

	if metrics.Endpoints["/test"] != 1 {
		t.Errorf("Expected endpoint /test count 1, got %d", metrics.Endpoints["/test"])
	}

	// Record an error request
	metrics.RecordRequest("/error", 500, 200*time.Millisecond, fmt.Errorf("test error"))

	if metrics.RequestCount != 2 {
		t.Errorf("Expected request count 2, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metrics.ErrorCount)
	}

	if metrics.StatusCodes[500] != 1 {
		t.Errorf("Expected status code 500 count 1, got %d", metrics.StatusCodes[500])
	}

	if metrics.Errors["test error"] != 1 {
		t.Errorf("Expected error 'test error' count 1, got %d", metrics.Errors["test error"])
	}
}

// TestMetrics_GetMetrics tests metrics retrieval
func TestMetrics_GetMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Record some requests
	metrics.RecordRequest("/test1", 200, 100*time.Millisecond, nil)
	metrics.RecordRequest("/test2", 200, 150*time.Millisecond, nil)
	metrics.RecordRequest("/error", 500, 200*time.Millisecond, fmt.Errorf("test error"))

	metricsData := metrics.GetMetrics()

	if metricsData["request_count"] != int64(3) {
		t.Errorf("Expected request count 3, got %v", metricsData["request_count"])
	}

	if metricsData["error_count"] != int64(1) {
		t.Errorf("Expected error count 1, got %v", metricsData["error_count"])
	}

	errorRate := metricsData["error_rate"].(float64)
	if errorRate != 1.0/3.0 {
		t.Errorf("Expected error rate %f, got %f", 1.0/3.0, errorRate)
	}

	avgResponseTime := metricsData["avg_response_time_ms"].(float64)
	expectedAvg := (100 + 150 + 200) / 3.0
	if avgResponseTime != expectedAvg {
		t.Errorf("Expected average response time %f, got %f", expectedAvg, avgResponseTime)
	}
}

// TestMetrics_GetHealthStatus tests health status
func TestMetrics_GetHealthStatus(t *testing.T) {
	metrics := NewMetrics()

	// Test healthy status
	metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)
	health := metrics.GetHealthStatus()

	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	// Test degraded status (high error rate)
	for i := 0; i < 10; i++ {
		metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)
	}
	for i := 0; i < 1; i++ {
		metrics.RecordRequest("/error", 500, 100*time.Millisecond, fmt.Errorf("test error"))
	}

	health = metrics.GetHealthStatus()
	// With 9% error rate, status should be degraded (between 5% and 10% threshold)
	if health["status"] != "degraded" {
		t.Errorf("Expected status 'degraded' for 9%% error rate, got %v", health["status"])
	}

	// Test unhealthy status (very high error rate)
	for i := 0; i < 5; i++ {
		metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)
	}
	for i := 0; i < 5; i++ {
		metrics.RecordRequest("/error", 500, 100*time.Millisecond, fmt.Errorf("test error"))
	}

	health = metrics.GetHealthStatus()
	if health["status"] != "unhealthy" {
		t.Errorf("Expected status 'unhealthy' for 50%% error rate, got %v", health["status"])
	}
}

// TestMetrics_Reset tests metrics reset
func TestMetrics_Reset(t *testing.T) {
	metrics := NewMetrics()

	// Record some data
	metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)
	metrics.RecordRequest("/error", 500, 200*time.Millisecond, fmt.Errorf("test error"))

	// Reset
	metrics.Reset()

	if metrics.RequestCount != 0 {
		t.Error("Request count should be 0 after reset")
	}

	if metrics.ErrorCount != 0 {
		t.Error("Error count should be 0 after reset")
	}

	if len(metrics.StatusCodes) != 0 {
		t.Error("Status codes should be empty after reset")
	}

	if len(metrics.Endpoints) != 0 {
		t.Error("Endpoints should be empty after reset")
	}

	if len(metrics.Errors) != 0 {
		t.Error("Errors should be empty after reset")
	}
}

// TestMetricsHandler tests metrics handler
func TestMetricsHandler(t *testing.T) {
	// Reset global metrics
	ResetMetrics()

	// Record some data
	RecordRequest("/test", 200, 100*time.Millisecond, nil)
	RecordRequest("/error", 500, 200*time.Millisecond, fmt.Errorf("test error"))

	// Test handler
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	MetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got %s", contentType)
	}
}

// TestHealthHandler tests health handler
func TestHealthHandler(t *testing.T) {
	// Reset global metrics
	ResetMetrics()

	// Record some data
	RecordRequest("/test", 200, 100*time.Millisecond, nil)

	// Test handler
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got %s", contentType)
	}
}

// TestPrometheusHandler tests Prometheus handler
func TestPrometheusHandler(t *testing.T) {
	// Reset global metrics
	ResetMetrics()

	// Record some data
	RecordRequest("/test", 200, 100*time.Millisecond, nil)
	RecordRequest("/error", 500, 200*time.Millisecond, fmt.Errorf("test error"))

	// Test handler
	req := httptest.NewRequest("GET", "/prometheus", nil)
	w := httptest.NewRecorder()
	PrometheusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "request_count_total") {
		t.Error("Prometheus output should contain request_count_total")
	}

	if !strings.Contains(body, "error_count_total") {
		t.Error("Prometheus output should contain error_count_total")
	}
}

// TestConcurrentAccess tests concurrent access to metrics
func TestConcurrentAccess(t *testing.T) {
	metrics := NewMetrics()

	// Test concurrent access
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			metrics.RecordRequest("/test", 200, 100*time.Millisecond, nil)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	if metrics.RequestCount != int64(concurrency) {
		t.Errorf("Expected request count %d, got %d", concurrency, metrics.RequestCount)
	}
}
