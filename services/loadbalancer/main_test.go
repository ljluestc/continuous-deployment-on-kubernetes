//go:build unit
// +build unit

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer()
	if lb == nil {
		t.Fatal("Expected load balancer to be created")
	}
	if lb.serverPool == nil {
		t.Fatal("Expected server pool to be created")
	}
}

func TestAddBackend(t *testing.T) {
	lb := NewLoadBalancer()
	err := lb.AddBackend("http://localhost:8080")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	backends := lb.serverPool.GetBackends()
	if len(backends) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(backends))
	}
}

func TestAddBackend_InvalidURL(t *testing.T) {
	lb := NewLoadBalancer()
	err := lb.AddBackend("://invalid")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestBackend_SetAlive(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	backend := &Backend{URL: u, Alive: true}

	backend.SetAlive(false)
	if backend.IsAlive() {
		t.Error("Expected backend to be not alive")
	}

	backend.SetAlive(true)
	if !backend.IsAlive() {
		t.Error("Expected backend to be alive")
	}
}

func TestServerPool_AddBackend(t *testing.T) {
	pool := &ServerPool{}
	u, _ := url.Parse("http://localhost:8080")
	backend := &Backend{URL: u, Alive: true}

	pool.AddBackend(backend)

	backends := pool.GetBackends()
	if len(backends) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(backends))
	}
}

func TestServerPool_GetNextPeer(t *testing.T) {
	pool := &ServerPool{}
	u1, _ := url.Parse("http://localhost:8080")
	u2, _ := url.Parse("http://localhost:8081")

	backend1 := &Backend{URL: u1, Alive: true}
	backend2 := &Backend{URL: u2, Alive: true}

	pool.AddBackend(backend1)
	pool.AddBackend(backend2)

	peer1 := pool.GetNextPeer()
	if peer1 == nil {
		t.Fatal("Expected to get a peer")
	}

	peer2 := pool.GetNextPeer()
	if peer2 == nil {
		t.Fatal("Expected to get a peer")
	}

	// Should round-robin
	if peer1 == peer2 {
		t.Error("Expected different peers for round-robin")
	}
}

func TestServerPool_GetNextPeer_NoBackends(t *testing.T) {
	pool := &ServerPool{}
	peer := pool.GetNextPeer()
	if peer != nil {
		t.Error("Expected nil peer when no backends")
	}
}

func TestServerPool_GetNextPeer_AllDown(t *testing.T) {
	pool := &ServerPool{}
	u, _ := url.Parse("http://localhost:8080")
	backend := &Backend{URL: u, Alive: false}
	pool.AddBackend(backend)

	peer := pool.GetNextPeer()
	if peer != nil {
		t.Error("Expected nil peer when all backends down")
	}
}

func TestGetStats(t *testing.T) {
	lb := NewLoadBalancer()
	lb.AddBackend("http://localhost:8080")

	stats := lb.GetStats()
	if len(stats) != 1 {
		t.Errorf("Expected 1 stat entry, got %d", len(stats))
	}

	if stats[0]["url"] != "http://localhost:8080" {
		t.Errorf("Expected URL http://localhost:8080, got %v", stats[0]["url"])
	}
}

func TestAddBackendHandler(t *testing.T) {
	lb = NewLoadBalancer()

	reqBody := map[string]interface{}{
		"url": "http://localhost:8080",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/add-backend", bytes.NewReader(body))
	w := httptest.NewRecorder()

	addBackendHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAddBackendHandler_InvalidMethod(t *testing.T) {
	lb = NewLoadBalancer()

	req := httptest.NewRequest(http.MethodGet, "/add-backend", nil)
	w := httptest.NewRecorder()

	addBackendHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	lb = NewLoadBalancer()
	lb.AddBackend("http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	statsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&stats)

	if len(stats) != 1 {
		t.Errorf("Expected 1 stat entry, got %d", len(stats))
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestServeHTTP_NoBackends(t *testing.T) {
	lb = NewLoadBalancer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lb.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestStartHealthCheck(t *testing.T) {
	lb := NewLoadBalancer()
	lb.AddBackend("http://localhost:9999") // Non-existent backend

	// Start health check with short interval
	lb.StartHealthCheck(100 * time.Millisecond)

	// Wait for health check to run
	time.Sleep(200 * time.Millisecond)

	// Backend should be marked as down
	backends := lb.serverPool.GetBackends()
	if len(backends) > 0 && backends[0].IsAlive() {
		t.Error("Expected backend to be marked as down after health check")
	}
}

