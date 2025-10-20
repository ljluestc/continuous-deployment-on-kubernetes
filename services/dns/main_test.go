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
	if service.records == nil {
		t.Fatal("Expected records map to be initialized")
	}
	if service.cache == nil {
		t.Fatal("Expected cache map to be initialized")
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
	if record.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", record.IPAddress)
	}
	if record.Type != "A" {
		t.Errorf("Expected type A, got %s", record.Type)
	}
	if record.TTL != 300 {
		t.Errorf("Expected TTL 300, got %d", record.TTL)
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

func TestResolve_FromCache(t *testing.T) {
	service := NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 300)
	
	// First resolve
	record1, _ := service.Resolve("example.com")
	// Second resolve should hit cache
	record2, _ := service.Resolve("example.com")
	
	if record1.IPAddress != record2.IPAddress {
		t.Error("Expected same record from cache")
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

func TestDeleteRecord_NotFound(t *testing.T) {
	service := NewDNSService()
	
	err := service.DeleteRecord("nonexistent.com")
	if err != nil {
		t.Errorf("Expected no error for non-existent domain, got %v", err)
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

func TestListRecords_Empty(t *testing.T) {
	service := NewDNSService()
	
	records := service.ListRecords()
	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
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
	
	var record DNSRecord
	json.NewDecoder(w.Body).Decode(&record)
	if record.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", record.Domain)
	}
}

func TestAddRecordHandler_InvalidMethod(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	w := httptest.NewRecorder()
	
	addRecordHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestAddRecordHandler_InvalidJSON(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	addRecordHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
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
	
	var record DNSRecord
	json.NewDecoder(w.Body).Decode(&record)
	if record.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", record.IPAddress)
	}
}

func TestResolveHandler_MissingDomain(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodGet, "/resolve", nil)
	w := httptest.NewRecorder()
	
	resolveHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestResolveHandler_NotFound(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodGet, "/resolve?domain=nonexistent.com", nil)
	w := httptest.NewRecorder()
	
	resolveHandler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDeleteRecordHandler(t *testing.T) {
	service = NewDNSService()
	service.AddRecord("example.com", "192.168.1.1", "A", 300)
	
	req := httptest.NewRequest(http.MethodDelete, "/delete?domain=example.com", nil)
	w := httptest.NewRecorder()
	
	deleteRecordHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteRecordHandler_InvalidMethod(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodGet, "/delete", nil)
	w := httptest.NewRecorder()
	
	deleteRecordHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestDeleteRecordHandler_MissingDomain(t *testing.T) {
	service = NewDNSService()
	
	req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	w := httptest.NewRecorder()
	
	deleteRecordHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestListRecordsHandler(t *testing.T) {
	service = NewDNSService()
	service.AddRecord("example1.com", "192.168.1.1", "A", 300)
	service.AddRecord("example2.com", "192.168.1.2", "A", 300)
	
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	w := httptest.NewRecorder()
	
	listRecordsHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var records []*DNSRecord
	json.NewDecoder(w.Body).Decode(&records)
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	healthHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", resp["status"])
	}
}
