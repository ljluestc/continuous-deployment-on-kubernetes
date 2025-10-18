// +build integration

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
	"strings"
	"sync"
	"testing"
	"time"
)

// TestIntegration_FrontendBackend_EndToEnd tests full frontend->backend flow
func TestIntegration_FrontendBackend_EndToEnd(t *testing.T) {
	// Start backend server
	backendHandler := http.NewServeMux()
	backendHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, err := json.Marshal(i)
		if err != nil {
			t.Errorf("Failed to marshal JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})
	backendHandler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	backendHandler.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", version)
	})

	backend := httptest.NewServer(backendHandler)
	defer backend.Close()

	// Start frontend server
	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	frontendHandler := http.NewServeMux()
	frontendHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL + "/")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Error: %s\n", err.Error())
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s\n", err.Error())
			return
		}

		var i Instance
		err = json.Unmarshal(body, &i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s\n", err.Error())
			return
		}

		tpl.Execute(w, &i)
	})

	frontendHandler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL + "/healthz")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Backend could not be connected to: %s", err.Error())
			return
		}
		defer resp.Body.Close()
		ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
	})

	frontendHandler.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", version)
	})

	frontend := httptest.NewServer(frontendHandler)
	defer frontend.Close()

	// Test frontend root endpoint
	resp, err := http.Get(frontend.URL + "/")
	if err != nil {
		t.Fatalf("Failed to query frontend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "<!doctype html>") {
		t.Error("Frontend should return HTML")
	}

	// Test frontend health endpoint
	resp, err = http.Get(frontend.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to query frontend health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Frontend health check should return 200, got %d", resp.StatusCode)
	}

	// Test frontend version endpoint
	resp, err = http.Get(frontend.URL + "/version")
	if err != nil {
		t.Fatalf("Failed to query frontend version: %v", err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(body), version) {
		t.Errorf("Version endpoint should return '%s', got '%s'", version, string(body))
	}

	// Test backend directly
	resp, err = http.Get(backend.URL + "/")
	if err != nil {
		t.Fatalf("Failed to query backend: %v", err)
	}
	defer resp.Body.Close()

	var instance Instance
	body, _ = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &instance)
	if err != nil {
		t.Fatalf("Failed to unmarshal backend response: %v", err)
	}

	if instance.Version != version {
		t.Errorf("Backend should return version '%s', got '%s'", version, instance.Version)
	}
}

// TestIntegration_MultipleBackends tests load balancing scenario
func TestIntegration_MultipleBackends_LoadBalance(t *testing.T) {
	// Start multiple backend servers
	backends := make([]*httptest.Server, 3)
	for i := 0; i < 3; i++ {
		id := i
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			instance := &Instance{
				Id:      fmt.Sprintf("backend-%d", id),
				Version: version,
			}
			resp, _ := json.Marshal(instance)
			w.Write(resp)
		})
		backends[i] = httptest.NewServer(handler)
		defer backends[i].Close()
	}

	// Test each backend is accessible
	for i, backend := range backends {
		resp, err := http.Get(backend.URL)
		if err != nil {
			t.Fatalf("Backend %d failed: %v", i, err)
		}
		defer resp.Body.Close()

		var instance Instance
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &instance)

		expectedID := fmt.Sprintf("backend-%d", i)
		if instance.Id != expectedID {
			t.Errorf("Expected backend ID '%s', got '%s'", expectedID, instance.Id)
		}
	}
}

// TestIntegration_ConcurrentRequests tests concurrent load
func TestIntegration_ConcurrentRequests_HandleLoad(t *testing.T) {
	// Start backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	// Send concurrent requests
	concurrency := 50
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := http.Get(backend.URL)
			if err != nil {
				errors <- fmt.Errorf("Request %d failed: %v", id, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("Request %d got status %d", id, resp.StatusCode)
				return
			}

			var instance Instance
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &instance)
			if err != nil {
				errors <- fmt.Errorf("Request %d failed to unmarshal: %v", id, err)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestIntegration_BackendFailover tests failover handling
func TestIntegration_BackendFailover_HandleFailure(t *testing.T) {
	// Start backend that will be closed
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))

	// Verify backend works
	resp, err := http.Get(backend.URL)
	if err != nil {
		t.Fatalf("Initial request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Initial request should succeed")
	}

	// Close backend to simulate failure
	backend.Close()

	// Frontend should handle failure gracefully
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err = client.Get(backend.URL)

	// Expect connection error
	if err == nil {
		t.Error("Should get error when backend is down")
	}
}

// TestIntegration_HealthChecks tests health check propagation
func TestIntegration_HealthChecks_PropagateStatus(t *testing.T) {
	// Start healthy backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Create frontend health check
	client := &http.Client{Timeout: 5 * time.Second}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(http.StatusOK)
	})

	frontend := httptest.NewServer(handler)
	defer frontend.Close()

	// Test health check passes
	resp, err := http.Get(frontend.URL)
	if err != nil {
		t.Fatalf("Health check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check should pass, got status %d", resp.StatusCode)
	}
}

// TestIntegration_VersionConsistency tests version across endpoints
func TestIntegration_VersionConsistency_SameVersion(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", version)
	})
	handler.HandleFunc("/instance", func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Get version from /version endpoint
	resp, _ := http.Get(server.URL + "/version")
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	versionEndpoint := strings.TrimSpace(string(body))

	// Get version from instance endpoint
	resp, _ = http.Get(server.URL + "/instance")
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var instance Instance
	json.Unmarshal(body, &instance)

	// Versions should match
	if versionEndpoint != instance.Version {
		t.Errorf("Version mismatch: endpoint=%s, instance=%s", versionEndpoint, instance.Version)
	}

	if instance.Version != version {
		t.Errorf("Instance version should be '%s', got '%s'", version, instance.Version)
	}
}

// TestIntegration_StressTest tests system under stress
func TestIntegration_StressTest_SustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	// Run sustained load for 5 seconds
	duration := 5 * time.Second
	concurrency := 20
	deadline := time.Now().Add(duration)

	var wg sync.WaitGroup
	var successCount, errorCount int
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for time.Now().Before(deadline) {
				resp, err := http.Get(backend.URL)
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}
				resp.Body.Close()

				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	t.Logf("Stress test completed: %d successful, %d errors", successCount, errorCount)

	if successCount == 0 {
		t.Error("No successful requests during stress test")
	}

	errorRate := float64(errorCount) / float64(successCount+errorCount)
	if errorRate > 0.01 { // Allow up to 1% error rate
		t.Errorf("Error rate too high: %.2f%%", errorRate*100)
	}
}

// BenchmarkIntegration_FrontendBackend benchmarks end-to-end performance
func BenchmarkIntegration_FrontendBackend(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, _ := client.Get(backend.URL)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var i Instance
		json.Unmarshal(body, &i)
		tpl.Execute(w, &i)
	}))
	defer frontend.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Get(frontend.URL)
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
}
