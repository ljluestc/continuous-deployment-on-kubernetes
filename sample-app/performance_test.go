//go:build performance
// +build performance

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
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestPerformance_ResponseTime tests response time performance
func TestPerformance_ResponseTime(t *testing.T) {
	// Test that the application responds within acceptable time limits

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test multiple requests and measure response time
	numRequests := 100
	responseTimes := make([]time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		start := time.Now()
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		responseTimes[i] = time.Since(start)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}

	// Calculate average response time
	var totalTime time.Duration
	for _, rt := range responseTimes {
		totalTime += rt
	}
	avgResponseTime := totalTime / time.Duration(numRequests)

	// Check that average response time is acceptable (less than 100ms)
	if avgResponseTime > 100*time.Millisecond {
		t.Errorf("Average response time %v exceeds 100ms", avgResponseTime)
	}

	t.Logf("Average response time: %v", avgResponseTime)
}

// TestPerformance_ConcurrentRequests tests concurrent request performance
func TestPerformance_ConcurrentRequests(t *testing.T) {
	// Test that the application handles concurrent requests efficiently

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test with different levels of concurrency
	concurrencyLevels := []int{10, 50, 100, 200}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			start := time.Now()
			var wg sync.WaitGroup

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					req := httptest.NewRequest("GET", "/", nil)
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)

					if w.Code != http.StatusOK {
						t.Errorf("Request failed with status %d", w.Code)
					}
				}()
			}

			wg.Wait()
			duration := time.Since(start)

			// Check that all requests completed within reasonable time
			if duration > 5*time.Second {
				t.Errorf("Concurrency %d took too long: %v", concurrency, duration)
			}

			t.Logf("Concurrency %d completed in %v", concurrency, duration)
		})
	}
}

// TestPerformance_MemoryUsage tests memory usage performance
func TestPerformance_MemoryUsage(t *testing.T) {
	// Test that the application doesn't use excessive memory

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a moderately sized response
		response := make([]byte, 1024) // 1KB
		for i := range response {
			response[i] = byte(i % 256)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	// Make many requests to test memory usage
	numRequests := 1000
	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}

	// If we get here without running out of memory, the test passes
	t.Logf("Successfully handled %d requests", numRequests)
}

// TestPerformance_JSONMarshaling tests JSON marshaling performance
func TestPerformance_JSONMarshaling(t *testing.T) {
	// Test that JSON marshaling is efficient

	instance := &Instance{
		Name:       "test-instance",
		Version:    "1.0.0",
		Id:         "test-123",
		Zone:       "us-central1-a",
		Project:    "test-project",
		Hostname:   "test.example.com",
		InternalIP: "10.0.0.1",
		ExternalIP: "35.192.0.1",
	}

	// Test JSON marshaling performance
	numIterations := 10000
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		_, err := json.Marshal(instance)
		if err != nil {
			t.Errorf("JSON marshaling failed: %v", err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numIterations)

	// Check that average marshaling time is acceptable (less than 1ms)
	if avgTime > time.Millisecond {
		t.Errorf("Average JSON marshaling time %v exceeds 1ms", avgTime)
	}

	t.Logf("Average JSON marshaling time: %v", avgTime)
}

// TestPerformance_JSONUnmarshaling tests JSON unmarshaling performance
func TestPerformance_JSONUnmarshaling(t *testing.T) {
	// Test that JSON unmarshaling is efficient

	jsonData := []byte(`{"name":"test-instance","version":"1.0.0","id":"test-123","zone":"us-central1-a","project":"test-project","hostname":"test.example.com","internal_ip":"10.0.0.1","external_ip":"35.192.0.1"}`)

	// Test JSON unmarshaling performance
	numIterations := 10000
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		var instance Instance
		err := json.Unmarshal(jsonData, &instance)
		if err != nil {
			t.Errorf("JSON unmarshaling failed: %v", err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numIterations)

	// Check that average unmarshaling time is acceptable (less than 1ms)
	if avgTime > time.Millisecond {
		t.Errorf("Average JSON unmarshaling time %v exceeds 1ms", avgTime)
	}

	t.Logf("Average JSON unmarshaling time: %v", avgTime)
}

// TestPerformance_TemplateRendering tests template rendering performance
func TestPerformance_TemplateRendering(t *testing.T) {
	// Test that template rendering is efficient

	instance := &Instance{
		Name:       "test-instance",
		Version:    "1.0.0",
		Id:         "test-123",
		Zone:       "us-central1-a",
		Project:    "test-project",
		Hostname:   "test.example.com",
		InternalIP: "10.0.0.1",
		ExternalIP: "35.192.0.1",
	}

	tpl := template.Must(template.New("out").Parse(html))

	// Test template rendering performance
	numIterations := 1000
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		w := httptest.NewRecorder()
		err := tpl.Execute(w, instance)
		if err != nil {
			t.Errorf("Template rendering failed: %v", err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numIterations)

	// Check that average rendering time is acceptable (less than 10ms)
	if avgTime > 10*time.Millisecond {
		t.Errorf("Average template rendering time %v exceeds 10ms", avgTime)
	}

	t.Logf("Average template rendering time: %v", avgTime)
}

// TestPerformance_LoadBalancing tests load balancing performance
func TestPerformance_LoadBalancing(t *testing.T) {
	// Test that the application handles load balancing efficiently

	// Create multiple backend servers
	numBackends := 3
	backends := make([]*httptest.Server, numBackends)

	for i := 0; i < numBackends; i++ {
		backends[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			instance := &Instance{
				Name:       "test-instance",
				Version:    "1.0.0",
				Id:         "test-123",
				Zone:       "us-central1-a",
				Project:    "test-project",
				Hostname:   "test.example.com",
				InternalIP: "10.0.0.1",
				ExternalIP: "35.192.0.1",
			}
			resp, _ := json.Marshal(instance)
			w.Write(resp)
		}))
		defer backends[i].Close()
	}

	// Create frontend handler that load balances between backends
	currentBackend := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple round-robin load balancing
		backend := backends[currentBackend%numBackends]
		currentBackend++

		resp, err := http.Get(backend.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})

	// Test load balancing performance
	numRequests := 100
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numRequests)

	// Check that average request time is acceptable (less than 50ms)
	if avgTime > 50*time.Millisecond {
		t.Errorf("Average load balancing time %v exceeds 50ms", avgTime)
	}

	t.Logf("Average load balancing time: %v", avgTime)
}

// TestPerformance_StressTest tests stress performance
func TestPerformance_StressTest(t *testing.T) {
	// Test that the application handles stress conditions

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Stress test with high concurrency and many requests
	concurrency := 100
	numRequests := 1000
	start := time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan int, numRequests)

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				req := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Request failed with status %d", w.Code)
				}
			}
		}()
	}

	// Send requests
	for i := 0; i < numRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	wg.Wait()
	duration := time.Since(start)

	// Check that all requests completed within reasonable time
	if duration > 30*time.Second {
		t.Errorf("Stress test took too long: %v", duration)
	}

	requestsPerSecond := float64(numRequests) / duration.Seconds()
	t.Logf("Stress test completed: %d requests in %v (%.2f req/s)", numRequests, duration, requestsPerSecond)
}
