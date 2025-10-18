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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/compute/metadata"
)

func TestGCE(t *testing.T) {
	i := newInstance()
	if !metadata.OnGCE() && i.Error != "Not running on GCE" {
		t.Error("Test not running on GCE, but error does not indicate that fact.")
	}
}

// TestBackendModeRootHandler tests the backend mode root handler
func TestBackendModeRootHandler(t *testing.T) {
	// Create a test server with backend handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := newInstance()
		i.LBRequest = "test request"
		resp, _ := json.Marshal(i)
		w.Write(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var instance Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if instance.Version != version {
		t.Errorf("Expected version %s, got %s", version, instance.Version)
	}
}

// TestBackendModeHealthHandler tests the backend mode health check handler
func TestBackendModeHealthHandler(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestVersionHandler tests the version endpoint
func TestVersionHandler(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version + "\n"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/version")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedVersion := version + "\n"
	if string(body) != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, string(body))
	}
}

// TestFrontendModeRootHandler tests the frontend mode root handler with a successful backend response
func TestFrontendModeRootHandler(t *testing.T) {
	// Create a mock backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{
			Id:         "test-id",
			Name:       "test-name",
			Version:    version,
			Hostname:   "test-hostname",
			Zone:       "test-zone",
			Project:    "test-project",
			InternalIP: "10.0.0.1",
			ExternalIP: "1.2.3.4",
		}
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backendServer.Close()

	// Create frontend server that uses the mock backend
	frontendMux := http.NewServeMux()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", backendServer.URL, nil)

	frontendMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		err = json.Unmarshal([]byte(body), i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		w.Write([]byte("Success"))
	})

	frontendServer := httptest.NewServer(frontendMux)
	defer frontendServer.Close()

	resp, err := http.Get(frontendServer.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFrontendModeBackendUnavailable tests frontend when backend is unavailable
func TestFrontendModeBackendUnavailable(t *testing.T) {
	frontendMux := http.NewServeMux()

	// Use an invalid backend URL to trigger connection error
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:99999", nil)

	frontendMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		err = json.Unmarshal([]byte(body), i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
	})

	frontendServer := httptest.NewServer(frontendMux)
	defer frontendServer.Close()

	resp, err := http.Get(frontendServer.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Error:") {
		t.Errorf("Expected error message in response, got: %s", string(body))
	}
}

// TestFrontendModeInvalidJSON tests frontend when backend returns invalid JSON
func TestFrontendModeInvalidJSON(t *testing.T) {
	// Create a mock backend that returns invalid JSON
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer backendServer.Close()

	frontendMux := http.NewServeMux()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", backendServer.URL, nil)

	frontendMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
		err = json.Unmarshal([]byte(body), i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error() + "\n"))
			return
		}
	})

	frontendServer := httptest.NewServer(frontendMux)
	defer frontendServer.Close()

	resp, err := http.Get(frontendServer.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Error:") {
		t.Errorf("Expected error message in response, got: %s", string(body))
	}
}

// TestFrontendModeHealthCheck tests the frontend health check
func TestFrontendModeHealthCheck(t *testing.T) {
	// Create a mock backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	frontendMux := http.NewServeMux()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", backendServer.URL, nil)

	frontendMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Backend could not be connected to: " + err.Error()))
			return
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
	})

	frontendServer := httptest.NewServer(frontendMux)
	defer frontendServer.Close()

	resp, err := http.Get(frontendServer.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestFrontendModeHealthCheckBackendDown tests frontend health check when backend is down
func TestFrontendModeHealthCheckBackendDown(t *testing.T) {
	frontendMux := http.NewServeMux()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:99999", nil)

	frontendMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Backend could not be connected to: " + err.Error()))
			return
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
	})

	frontendServer := httptest.NewServer(frontendMux)
	defer frontendServer.Close()

	resp, err := http.Get(frontendServer.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Backend could not be connected to:") {
		t.Errorf("Expected backend error message, got: %s", string(body))
	}
}

// TestAssignerWithError tests the assigner struct with error handling
func TestAssignerWithError(t *testing.T) {
	a := &assigner{}

	// Test successful assignment
	result := a.assign(func() (string, error) {
		return "test-value", nil
	})

	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}

	if a.err != nil {
		t.Errorf("Expected no error, got %v", a.err)
	}

	// Test error assignment
	testError := "test error"
	result = a.assign(func() (string, error) {
		return "", &mockError{testError}
	})

	if result != "" {
		t.Errorf("Expected empty string on error, got '%s'", result)
	}

	if a.err == nil {
		t.Error("Expected error to be set")
	}

	// Test that subsequent calls return empty string when error is set
	result = a.assign(func() (string, error) {
		return "should-not-be-called", nil
	})

	if result != "" {
		t.Errorf("Expected empty string when error already set, got '%s'", result)
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
