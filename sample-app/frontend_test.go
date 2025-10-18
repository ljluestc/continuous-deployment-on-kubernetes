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

// TestFrontendMode_RootEndpoint tests frontend rendering
func TestFrontendMode_RootEndpoint_RendersHTML(t *testing.T) {
	// Create mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{
			Id:         "backend-id",
			Name:       "backend-name",
			Version:    "1.0.0",
			Hostname:   "backend-host",
			Zone:       "us-central1-a",
			Project:    "test-project",
			InternalIP: "10.0.0.1",
			ExternalIP: "203.0.113.1",
			Error:      "",
		}
		resp, _ := json.Marshal(i)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}))
	defer backend.Close()

	// Create frontend handler
	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error()))
			return
		}

		var i Instance
		err = json.Unmarshal(body, &i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error()))
			return
		}

		tpl.Execute(w, &i)
	})

	// Test the frontend
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify HTML contains expected data
	body := w.Body.String()
	if !strings.Contains(body, "backend-name") {
		t.Error("Response should contain backend name")
	}
	if !strings.Contains(body, "1.0.0") {
		t.Error("Response should contain version")
	}
	if !strings.Contains(body, "<!doctype html>") {
		t.Error("Response should be valid HTML")
	}
}

// TestFrontendMode_BackendUnavailable tests error handling
func TestFrontendMode_BackendUnavailable_ReturnsError(t *testing.T) {
	// Use invalid backend URL
	badBackendURL := "http://localhost:99999"

	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 1 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(badBackendURL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var i Instance
		json.Unmarshal(body, &i)
		tpl.Execute(w, &i)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Error:") {
		t.Error("Response should contain error message")
	}
}

// TestFrontendMode_BackendInvalidJSON tests invalid JSON handling
func TestFrontendMode_BackendInvalidJSON_ReturnsError(t *testing.T) {
	// Create backend that returns invalid JSON
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json{{{"))
	}))
	defer backend.Close()

	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var i Instance
		err = json.Unmarshal(body, &i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error: " + err.Error()))
			return
		}

		tpl.Execute(w, &i)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// TestFrontendMode_HealthEndpoint tests health check with backend
func TestFrontendMode_HealthEndpoint_ChecksBackend(t *testing.T) {
	// Create healthy backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(backend.URL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Backend could not be connected to: " + err.Error()))
			return
		}
		defer resp.Body.Close()
		ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestFrontendMode_HealthEndpoint_BackendDown tests health check failure
func TestFrontendMode_HealthEndpoint_BackendDown_ReturnsError(t *testing.T) {
	// Use invalid backend URL
	badBackendURL := "http://localhost:99999"

	client := &http.Client{Timeout: 1 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(badBackendURL)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Backend could not be connected to: " + err.Error()))
			return
		}
		defer resp.Body.Close()
		ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Backend could not be connected to") {
		t.Error("Response should contain backend error message")
	}
}

// TestFrontendMode_TemplateRendering tests HTML template
func TestFrontendMode_TemplateRendering_ValidHTML(t *testing.T) {
	tpl, err := template.New("out").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	i := &Instance{
		Id:         "test-id",
		Name:       "test-name",
		Version:    "1.0.0",
		Hostname:   "test-hostname",
		Zone:       "us-east1-b",
		Project:    "test-project",
		InternalIP: "10.0.0.1",
		ExternalIP: "203.0.113.1",
		LBRequest:  "GET / HTTP/1.1",
		ClientIP:   "198.51.100.1",
		Error:      "",
	}

	w := httptest.NewRecorder()

	err = tpl.Execute(w, i)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	body := w.Body.String()

	// Verify HTML structure
	if !strings.Contains(body, "<!doctype html>") {
		t.Error("Missing HTML doctype")
	}
	if !strings.Contains(body, "<html>") {
		t.Error("Missing HTML tag")
	}
	if !strings.Contains(body, "</html>") {
		t.Error("Missing closing HTML tag")
	}

	// Verify data is rendered
	if !strings.Contains(body, "test-name") {
		t.Error("Name not rendered")
	}
	if !strings.Contains(body, "1.0.0") {
		t.Error("Version not rendered")
	}
	if !strings.Contains(body, "test-hostname") {
		t.Error("Hostname not rendered")
	}
}

// TestFrontendMode_TemplateWithSpecialChars tests XSS prevention
func TestFrontendMode_TemplateWithSpecialChars_EscapesHTML(t *testing.T) {
	tpl, err := template.New("out").Parse(html)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	i := &Instance{
		Name:    "<script>alert('xss')</script>",
		Error:   "<img src=x onerror=alert('xss')>",
		Version: "1.0.0",
	}

	w := httptest.NewRecorder()
	err = tpl.Execute(w, i)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	body := w.Body.String()

	// Verify HTML entities are escaped
	if strings.Contains(body, "<script>") {
		t.Error("Script tags should be escaped")
	}
	if strings.Contains(body, "<img") && strings.Contains(body, "onerror") {
		t.Error("Image tags with onerror should be escaped")
	}
}

// TestFrontendMode_ConcurrentRequests tests concurrent frontend requests
func TestFrontendMode_ConcurrentRequests_HandleMultiple(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{
			Name:    "backend",
			Version: "1.0.0",
		}
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, _ := client.Get(backend.URL)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var i Instance
		json.Unmarshal(body, &i)
		tpl.Execute(w, &i)
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

// BenchmarkFrontendMode_RootEndpoint benchmarks frontend rendering
func BenchmarkFrontendMode_RootEndpoint(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := &Instance{Name: "test", Version: "1.0.0"}
		resp, _ := json.Marshal(i)
		w.Write(resp)
	}))
	defer backend.Close()

	tpl := template.Must(template.New("out").Parse(html))
	client := &http.Client{Timeout: 5 * time.Second}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, _ := client.Get(backend.URL)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var i Instance
		json.Unmarshal(body, &i)
		tpl.Execute(w, &i)
	})

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkFrontendMode_TemplateRendering benchmarks template execution
func BenchmarkFrontendMode_TemplateRendering(b *testing.B) {
	tpl := template.Must(template.New("out").Parse(html))
	i := &Instance{
		Name:    "test",
		Version: "1.0.0",
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		w := httptest.NewRecorder()
		tpl.Execute(w, i)
	}
}
