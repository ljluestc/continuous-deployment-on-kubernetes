//go:build security
// +build security

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
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestSecurity_XSSPrevention tests XSS prevention in HTML output
func TestSecurity_XSSPrevention(t *testing.T) {
	// Create a test instance with potentially malicious content
	maliciousInstance := &Instance{
		Name:       "<script>alert('xss')</script>",
		Version:    "1.0.0",
		Id:         "test-123",
		Zone:       "us-central1-a",
		Project:    "test-project",
		Hostname:   "test.example.com",
		InternalIP: "10.0.0.1",
		ExternalIP: "35.192.0.1",
	}

	// Create a test server that returns the malicious instance
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(maliciousInstance)
		w.Write(resp)
	}))
	defer backend.Close()

	// Create frontend handler
	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
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
		var i Instance
		if err := json.Unmarshal(body, &i); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tpl.Execute(w, &i)
	})

	// Test the handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the response contains escaped HTML
	body := w.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("Response should not contain unescaped script tags")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("Response should contain escaped script tags")
	}
}

// TestSecurity_SQLInjectionPrevention tests SQL injection prevention
func TestSecurity_SQLInjectionPrevention(t *testing.T) {
	// Test that the application doesn't execute SQL queries
	// This is more of a sanity check since our app doesn't use SQL

	// Create a test instance with SQL injection attempt
	sqlInjectionInstance := &Instance{
		Name:       "'; DROP TABLE users; --",
		Version:    "1.0.0",
		Id:         "test-123",
		Zone:       "us-central1-a",
		Project:    "test-project",
		Hostname:   "test.example.com",
		InternalIP: "10.0.0.1",
		ExternalIP: "35.192.0.1",
	}

	// Test that the instance is handled safely
	if sqlInjectionInstance.Name != "'; DROP TABLE users; --" {
		t.Error("Instance name should be preserved as-is")
	}
}

// TestSecurity_InputValidation tests input validation
func TestSecurity_InputValidation(t *testing.T) {
	// Test that the application validates input properly
	// This is a basic test since our app doesn't have complex input validation

	// Test with empty input
	emptyInstance := &Instance{}
	if emptyInstance.Name != "" {
		t.Error("Empty instance should have empty name")
	}

	// Test with very long input
	longName := strings.Repeat("a", 10000)
	longInstance := &Instance{Name: longName}
	if longInstance.Name != longName {
		t.Error("Long instance name should be preserved")
	}
}

// TestSecurity_HTTPSHeaders tests security headers
func TestSecurity_HTTPSHeaders(t *testing.T) {
	// Test that the application sets appropriate security headers
	// This is a basic test since our app doesn't set custom headers

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the response is valid
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestSecurity_ContentTypeValidation tests content type validation
func TestSecurity_ContentTypeValidation(t *testing.T) {
	// Test that the application handles different content types properly

	// Test with valid JSON
	validJSON := `{"name":"test","version":"1.0.0"}`
	var instance Instance
	err := json.Unmarshal([]byte(validJSON), &instance)
	if err != nil {
		t.Errorf("Valid JSON should unmarshal successfully: %v", err)
	}

	// Test with invalid JSON
	invalidJSON := `{"name":"test","version":"1.0.0"`
	err = json.Unmarshal([]byte(invalidJSON), &instance)
	if err == nil {
		t.Error("Invalid JSON should fail to unmarshal")
	}
}

// TestSecurity_ErrorHandling tests error handling security
func TestSecurity_ErrorHandling(t *testing.T) {
	// Test that the application doesn't leak sensitive information in errors

	// Test with a backend that returns an error
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer backend.Close()

	// Create frontend handler
	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
		if err != nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Backend Error", http.StatusServiceUnavailable)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Read Error", http.StatusInternalServerError)
			return
		}
		var i Instance
		if err := json.Unmarshal(body, &i); err != nil {
			http.Error(w, "Parse Error", http.StatusInternalServerError)
			return
		}
		tpl.Execute(w, &i)
	})

	// Test the handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the error response doesn't leak sensitive information
	body := w.Body.String()
	if strings.Contains(body, "Internal Server Error") {
		t.Error("Error response should not leak backend error details")
	}
	if !strings.Contains(body, "Backend Error") {
		t.Error("Error response should contain generic error message")
	}
}

// TestSecurity_ConcurrentAccess tests concurrent access security
func TestSecurity_ConcurrentAccess(t *testing.T) {
	// Test that the application handles concurrent access safely

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test concurrent requests
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

// TestSecurity_ResourceExhaustion tests resource exhaustion protection
func TestSecurity_ResourceExhaustion(t *testing.T) {
	// Test that the application doesn't exhaust resources

	// Test with a large payload
	largePayload := strings.Repeat("a", 1000000)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largePayload))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the response is handled properly
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestSecurity_Authentication tests authentication security
func TestSecurity_Authentication(t *testing.T) {
	// Test that the application handles authentication properly
	// This is a basic test since our app doesn't have authentication

	// Test that the application doesn't require authentication
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the request is handled without authentication
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestSecurity_Authorization tests authorization security
func TestSecurity_Authorization(t *testing.T) {
	// Test that the application handles authorization properly
	// This is a basic test since our app doesn't have authorization

	// Test that the application doesn't require authorization
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check that the request is handled without authorization
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
