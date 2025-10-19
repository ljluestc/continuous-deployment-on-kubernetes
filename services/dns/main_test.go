//go:build unit
// +build unit

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDNSService(t *testing.T) {
	service := NewDNSService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestAddRecord(t *testing.T) {
	service := NewDNSService()
	record, err := service.AddRecord("example.com", "192.168.1.1", "A", 300)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if record.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", record.Domain)
	}
}

func TestResolve(t *testing.T) {
	service := NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 300)
	
	record, err := service.Resolve("example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if record.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", record.IPAddress)
	}
}

func TestResolve_NotFound(t *testing.T) {
	service := NewDNSService()
	record, _ := service.Resolve("nonexistent.com")
	if record != nil {
		t.Error("Expected nil record for non-existent domain")
	}
}

func TestResolve_CacheExpiry(t *testing.T) {
	service := NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 1) // 1 second TTL
	
	// First resolve should work
	record, _ := service.Resolve("example.com")
	if record == nil {
		t.Fatal("Expected record to be found")
	}
	
	// Wait for cache to expire
	time.Sleep(2 * time.Second)
	
	// Should still resolve from records
	record, _ = service.Resolve("example.com")
	if record == nil {
		t.Error("Expected record to still be found after cache expiry")
	}
}

func TestDeleteRecord(t *testing.T) {
	service := NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 300)
	
	err := service.DeleteRecord("example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	record, _ := service.Resolve("example.com")
	if record != nil {
		t.Error("Expected record to be deleted")
	}
}

func TestListRecords(t *testing.T) {
	service := NewDNSService()
	service.AddRecord("example1.com", "192.168.1.1", "A", 300)
	service.AddRecord("example2.com", "192.168.1.2", "A", 300)
	
	records := service.ListRecords()
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestAddRecordHandler(t *testing.T) {
	service = NewDNSService()
	
	reqBody := map[string]interface{}{
		"domain":     "example.com",
		"ip_address": "192.168.1.1",
		"type":       "A",
		"ttl":        300,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	addRecordHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestResolveHandler(t *testing.T) {
	service = NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 300)
	
	req := httptest.NewRequest(http.MethodGet, "/resolve?domain=example.com", nil)
	w := httptest.NewRecorder()
	
	resolveHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
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

