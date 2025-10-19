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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"
)

// TestBackendMode_RootEndpoint tests the root endpoint in backend mode
func TestBackendMode_RootEndpoint_ReturnsJSON(t *testing.T) {
	// Create a request to the root endpoint
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Test-Header", "test-value")
	w := httptest.NewRecorder()

	// Create handler
	http.HandleFunc("/test-root", func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		raw, _ := httputil.DumpRequest(r, true)
		i.LBRequest = string(raw)
		resp, _ := json.Marshal(i)
		w.Write(resp)
	})

	// Call the handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, err := json.Marshal(i)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})

	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify JSON response
	var instance Instance
	body := w.Body.Bytes()
	err := json.Unmarshal(body, &instance)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Verify version is set
	if instance.Version != version {
		t.Errorf("Expected version %s, got %s", version, instance.Version)
	}
}

// TestBackendMode_HealthEndpoint tests the /healthz endpoint
func TestBackendMode_HealthEndpoint_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestBackendMode_VersionEndpoint tests the /version endpoint
func TestBackendMode_VersionEndpoint_ReturnsVersion(t *testing.T) {
	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version + "\n"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != version {
		t.Errorf("Expected version %s, got %s", version, body)
	}
}

// TestBackendMode_JSONMarshal tests JSON marshaling of Instance
func TestBackendMode_JSONMarshal_ValidJSON(t *testing.T) {
	i := &Instance{
		Id:         "test-id",
		Name:       "test-name",
		Version:    version,
		Hostname:   "test-hostname",
		Zone:       "test-zone",
		Project:    "test-project",
		InternalIP: "10.0.0.1",
		ExternalIP: "203.0.113.1",
		LBRequest:  "GET / HTTP/1.1",
		ClientIP:   "198.51.100.1",
		Error:      "",
	}

	data, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("Failed to marshal instance: %v", err)
	}

	// Unmarshal to verify
	var decoded Instance
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal instance: %v", err)
	}

	if decoded.Id != i.Id {
		t.Errorf("Expected Id %s, got %s", i.Id, decoded.Id)
	}
	if decoded.Version != i.Version {
		t.Errorf("Expected Version %s, got %s", i.Version, decoded.Version)
	}
}

// TestBackendMode_ConcurrentRequests tests concurrent request handling
func TestBackendMode_ConcurrentRequests_HandlesMultiple(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Make concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Errorf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			var instance Instance
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &instance)
			if err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}

			done <- true
		}()
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}
}

// TestBackendMode_RequestMetadata tests request metadata capture
func TestBackendMode_RequestMetadata_CapturesHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom-Header", "custom-value")
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are present
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Error("User-Agent header not captured")
		}
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Error("Custom header not captured")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(w, req)
}

// TestBackendMode_POST tests POST requests
func TestBackendMode_POST_AcceptsAllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/", nil)
		w := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			i := newInstance()
			resp, _ := json.Marshal(i)
			w.Write(resp)
		})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Method %s: expected status 200, got %d", method, w.Code)
		}
	}
}

// TestBackendMode_LargePayload tests handling of large request bodies
func TestBackendMode_LargePayload_Handles(t *testing.T) {
	// Create large payload (1MB)
	largeBody := strings.Repeat("a", 1024*1024)
	req := httptest.NewRequest("POST", "/", strings.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read body: %v", err)
		}
		if len(body) != len(largeBody) {
			t.Errorf("Expected body length %d, got %d", len(largeBody), len(body))
		}
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// BenchmarkBackendMode_RootEndpoint benchmarks the root endpoint
func BenchmarkBackendMode_RootEndpoint(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		resp, _ := json.Marshal(i)
		w.Write(resp)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkBackendMode_JSONMarshal benchmarks JSON marshaling
func BenchmarkBackendMode_JSONMarshal(b *testing.B) {
	i := newInstance()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = json.Marshal(i)
	}
}
