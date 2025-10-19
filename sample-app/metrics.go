//go:build !test
// +build !test

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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Metrics represents application metrics
type Metrics struct {
	mu                sync.RWMutex
	RequestCount      int64            `json:"request_count"`
	ErrorCount        int64            `json:"error_count"`
	ResponseTimeSum   time.Duration    `json:"response_time_sum"`
	ResponseTimeCount int64            `json:"response_time_count"`
	StartTime         time.Time        `json:"start_time"`
	LastRequestTime   time.Time        `json:"last_request_time"`
	StatusCodes       map[int]int64    `json:"status_codes"`
	Endpoints         map[string]int64 `json:"endpoints"`
	Errors            map[string]int64 `json:"errors"`
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:   time.Now(),
		StatusCodes: make(map[int]int64),
		Endpoints:   make(map[string]int64),
		Errors:      make(map[string]int64),
	}
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(endpoint string, statusCode int, responseTime time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount++
	m.LastRequestTime = time.Now()
	m.ResponseTimeSum += responseTime
	m.ResponseTimeCount++

	// Record status code
	m.StatusCodes[statusCode]++

	// Record endpoint
	m.Endpoints[endpoint]++

	// Record error if any
	if err != nil {
		m.ErrorCount++
		m.Errors[err.Error()]++
	}
}

// GetMetrics returns current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgResponseTime := time.Duration(0)
	if m.ResponseTimeCount > 0 {
		avgResponseTime = m.ResponseTimeSum / time.Duration(m.ResponseTimeCount)
	}

	uptime := time.Since(m.StartTime)

	return map[string]interface{}{
		"request_count":        m.RequestCount,
		"error_count":          m.ErrorCount,
		"error_rate":           float64(m.ErrorCount) / float64(m.RequestCount),
		"avg_response_time_ms": float64(avgResponseTime.Nanoseconds()) / 1e6,
		"uptime_seconds":       uptime.Seconds(),
		"start_time":           m.StartTime.Format(time.RFC3339),
		"last_request_time":    m.LastRequestTime.Format(time.RFC3339),
		"status_codes":         m.StatusCodes,
		"endpoints":            m.Endpoints,
		"errors":               m.Errors,
	}
}

// GetHealthStatus returns health status based on metrics
func (m *Metrics) GetHealthStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorRate := float64(0)
	if m.RequestCount > 0 {
		errorRate = float64(m.ErrorCount) / float64(m.RequestCount)
	}

	avgResponseTime := time.Duration(0)
	if m.ResponseTimeCount > 0 {
		avgResponseTime = m.ResponseTimeSum / time.Duration(m.ResponseTimeCount)
	}

	// Health status based on error rate and response time
	status := "healthy"
	if errorRate > 0.1 { // 10% error rate
		status = "unhealthy"
	} else if errorRate > 0.05 { // 5% error rate
		status = "degraded"
	}

	if avgResponseTime > 5*time.Second {
		status = "unhealthy"
	} else if avgResponseTime > 2*time.Second {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":            status,
		"error_rate":        errorRate,
		"avg_response_time": avgResponseTime.String(),
		"request_count":     m.RequestCount,
		"error_count":       m.ErrorCount,
		"uptime":            time.Since(m.StartTime).String(),
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount = 0
	m.ErrorCount = 0
	m.ResponseTimeSum = 0
	m.ResponseTimeCount = 0
	m.StartTime = time.Now()
	m.LastRequestTime = time.Time{}
	m.StatusCodes = make(map[int]int64)
	m.Endpoints = make(map[string]int64)
	m.Errors = make(map[string]int64)
}

// Global metrics instance
var globalMetrics = NewMetrics()

// RecordRequest is a convenience function to record a request
func RecordRequest(endpoint string, statusCode int, responseTime time.Duration, err error) {
	globalMetrics.RecordRequest(endpoint, statusCode, responseTime, err)
}

// GetMetrics is a convenience function to get metrics
func GetMetrics() map[string]interface{} {
	return globalMetrics.GetMetrics()
}

// GetHealthStatus is a convenience function to get health status
func GetHealthStatus() map[string]interface{} {
	return globalMetrics.GetHealthStatus()
}

// ResetMetrics is a convenience function to reset metrics
func ResetMetrics() {
	globalMetrics.Reset()
}

// MetricsHandler handles metrics requests
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := GetHealthStatus()

	w.Header().Set("Content-Type", "application/json")

	status := health["status"].(string)
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if status == "degraded" {
		statusCode = http.StatusOK // Still OK but degraded
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Error encoding health status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PrometheusHandler handles Prometheus metrics requests
func PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GetMetrics()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	// Convert metrics to Prometheus format
	fmt.Fprintf(w, "# HELP request_count_total Total number of requests\n")
	fmt.Fprintf(w, "# TYPE request_count_total counter\n")
	fmt.Fprintf(w, "request_count_total %d\n", metrics["request_count"])

	fmt.Fprintf(w, "# HELP error_count_total Total number of errors\n")
	fmt.Fprintf(w, "# TYPE error_count_total counter\n")
	fmt.Fprintf(w, "error_count_total %d\n", metrics["error_count"])

	fmt.Fprintf(w, "# HELP avg_response_time_seconds Average response time in seconds\n")
	fmt.Fprintf(w, "# TYPE avg_response_time_seconds gauge\n")
	fmt.Fprintf(w, "avg_response_time_seconds %f\n", metrics["avg_response_time_ms"].(float64)/1000)

	fmt.Fprintf(w, "# HELP uptime_seconds Uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE uptime_seconds gauge\n")
	fmt.Fprintf(w, "uptime_seconds %f\n", metrics["uptime_seconds"].(float64))

	// Status codes
	if statusCodes, ok := metrics["status_codes"].(map[int]int64); ok {
		fmt.Fprintf(w, "# HELP status_code_total Total number of requests by status code\n")
		fmt.Fprintf(w, "# TYPE status_code_total counter\n")
		for code, count := range statusCodes {
			fmt.Fprintf(w, "status_code_total{code=\"%d\"} %d\n", code, count)
		}
	}

	// Endpoints
	if endpoints, ok := metrics["endpoints"].(map[string]int64); ok {
		fmt.Fprintf(w, "# HELP endpoint_total Total number of requests by endpoint\n")
		fmt.Fprintf(w, "# TYPE endpoint_total counter\n")
		for endpoint, count := range endpoints {
			fmt.Fprintf(w, "endpoint_total{endpoint=\"%s\"} %d\n", endpoint, count)
		}
	}
}
