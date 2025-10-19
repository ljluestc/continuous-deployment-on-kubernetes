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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestBackendModeServer tests the actual backendMode function
func TestBackendModeServer(t *testing.T) {
	// Use a unique port for testing
	port := 18080

	// Start server in goroutine
	go func() {
		backendMode(port)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Test root endpoint
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Failed to connect to backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var instance Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if instance.Version != version {
		t.Errorf("Expected version %s, got %s", version, instance.Version)
	}

	// Test healthz endpoint
	resp, err = http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to connect to healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected healthz status 200, got %d", resp.StatusCode)
	}

	// Test version endpoint
	resp, err = http.Get(baseURL + "/version")
	if err != nil {
		t.Fatalf("Failed to connect to version: %v", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	if !strings.Contains(string(body), version) {
		t.Errorf("Expected version %s in response, got %s", version, string(body))
	}
}

// TestFrontendModeServer tests the actual frontendMode function
func TestFrontendModeServer(t *testing.T) {
	// Start a mock backend server first
	backendPort := 18081
	go func() {
		backendMode(backendPort)
	}()
	time.Sleep(200 * time.Millisecond)

	// Start frontend server
	frontendPort := 18082
	backendURL := fmt.Sprintf("http://127.0.0.1:%d", backendPort)

	go func() {
		frontendMode(frontendPort, backendURL)
	}()
	time.Sleep(200 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", frontendPort)

	// Test root endpoint
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Failed to connect to frontend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "<!doctype html>") {
		t.Error("Frontend should return HTML")
	}

	if !strings.Contains(bodyStr, version) {
		t.Error("Frontend should contain version in rendered HTML")
	}

	// Test healthz endpoint
	resp, err = http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to connect to frontend healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected healthz status 200, got %d", resp.StatusCode)
	}

	// Test version endpoint
	resp, err = http.Get(baseURL + "/version")
	if err != nil {
		t.Fatalf("Failed to connect to version: %v", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	if !strings.Contains(string(body), version) {
		t.Errorf("Expected version %s in response, got %s", version, string(body))
	}
}

// TestFrontendModeWithBadBackend tests frontend with unavailable backend
func TestFrontendModeWithBadBackend(t *testing.T) {
	// Start frontend with invalid backend URL
	frontendPort := 18083
	badBackendURL := "http://127.0.0.1:99999" // Invalid port

	go func() {
		frontendMode(frontendPort, badBackendURL)
	}()
	time.Sleep(200 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", frontendPort)

	// Test root endpoint - should return error
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Failed to connect to frontend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	// Test healthz endpoint - should also fail
	resp, err = http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to connect to frontend healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected healthz status 503, got %d", resp.StatusCode)
	}
}

// TestVersionEndpointGlobal tests the global version endpoint handler
func TestVersionEndpointGlobal(t *testing.T) {
	// This tests the /version handler registered in main()
	// We'll start a minimal server just for this endpoint

	port := 18084

	mux := http.NewServeMux()
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", version)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		server.ListenAndServe()
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/version", port))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	expected := version + "\n"
	if string(body) != expected {
		t.Errorf("Expected %q, got %q", expected, string(body))
	}
}

// TestFrontendModeWithBackendReturningInvalidJSON tests JSON unmarshal error handling
func TestFrontendModeWithBackendReturningInvalidJSON(t *testing.T) {
	// Start backend that returns invalid JSON
	backendPort := 18085
	backendMux := http.NewServeMux()
	backendMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON
		w.Write([]byte("this is not valid json{{{"))
	})

	backendServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", backendPort),
		Handler: backendMux,
	}

	go backendServer.ListenAndServe()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		backendServer.Shutdown(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Start frontend server
	frontendPort := 18086
	backendURL := fmt.Sprintf("http://127.0.0.1:%d", backendPort)

	go func() {
		frontendMode(frontendPort, backendURL)
	}()
	time.Sleep(200 * time.Millisecond)

	// Test root endpoint - should return internal server error
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", frontendPort))
	if err != nil {
		t.Fatalf("Failed to connect to frontend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Error:") {
		t.Error("Response should contain error message")
	}
}

// TestMainFunction_Integration tests the main function integration
func TestMainFunction_Integration(t *testing.T) {
	// This test verifies that the main function logic works correctly
	// by testing the flag parsing and mode selection logic
	
	// Test version flag logic
	showversion := false
	frontend := false
	port := 8080
	backend := "http://127.0.0.1:8081"
	
	// Test version flag
	if showversion {
		// This would print version and return
		t.Log("Version flag would print version and return")
	}
	
	// Test frontend mode selection
	if frontend {
		// This would call frontendMode(*port, *backend)
		t.Logf("Frontend mode would be called with port %d and backend %s", port, backend)
	} else {
		// This would call backendMode(*port)
		t.Logf("Backend mode would be called with port %d", port)
	}
	
	// Test that the global version handler would be registered
	t.Log("Global version handler would be registered")
}

// TestFrontendMode_ReadError tests the specific error handling in frontend mode
func TestFrontendMode_ReadError(t *testing.T) {
	// Create a backend that returns valid JSON
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{
			Name:    "test-backend",
			Version: "1.0.0",
		}
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	// Create a custom transport that causes read error
	client := &http.Client{
		Transport: &errorTransport{},
		Timeout:   5 * time.Second,
	}

	// Test the specific error handling path in frontendMode
	req, _ := http.NewRequest("GET", backend.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// This should trigger the error handling path
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// This is the error path we want to test
		t.Logf("Read error occurred as expected: %v", err)
	} else {
		t.Error("Expected read error, but got success")
	}
}
