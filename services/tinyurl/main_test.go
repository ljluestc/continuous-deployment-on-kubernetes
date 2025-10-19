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

func TestNewTinyURLService(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if service.baseURL != "http://test.com" {
		t.Errorf("Expected baseURL to be http://test.com, got %s", service.baseURL)
	}
	if len(service.mappings) != 0 {
		t.Errorf("Expected empty mappings, got %d", len(service.mappings))
	}
}

func TestGenerateShortURL(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	shortURL1 := service.GenerateShortURL("https://example.com")
	shortURL2 := service.GenerateShortURL("https://example.com")

	if len(shortURL1) != 8 {
		t.Errorf("Expected short URL length 8, got %d", len(shortURL1))
	}

	// Different calls should generate different short URLs due to timestamp
	if shortURL1 == shortURL2 {
		t.Error("Expected different short URLs for different calls")
	}
}

func TestCreateShortURL_Basic(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com/very/long/url"

	mapping, err := service.CreateShortURL(longURL, "", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if mapping.LongURL != longURL {
		t.Errorf("Expected long URL %s, got %s", longURL, mapping.LongURL)
	}

	if len(mapping.ShortURL) != 8 {
		t.Errorf("Expected short URL length 8, got %d", len(mapping.ShortURL))
	}

	if mapping.AccessCount != 0 {
		t.Errorf("Expected access count 0, got %d", mapping.AccessCount)
	}
}

func TestCreateShortURL_CustomAlias(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"
	customAlias := "myalias"

	mapping, err := service.CreateShortURL(longURL, customAlias, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if mapping.ShortURL != customAlias {
		t.Errorf("Expected short URL %s, got %s", customAlias, mapping.ShortURL)
	}
}

func TestCreateShortURL_CustomAliasDuplicate(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	customAlias := "myalias"

	_, err := service.CreateShortURL("https://example1.com", customAlias, 0)
	if err != nil {
		t.Fatalf("Expected no error for first creation, got %v", err)
	}

	_, err = service.CreateShortURL("https://example2.com", customAlias, 0)
	if err == nil {
		t.Error("Expected error for duplicate custom alias")
	}
}

func TestCreateShortURL_Deduplication(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"

	mapping1, err := service.CreateShortURL(longURL, "", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mapping2, err := service.CreateShortURL(longURL, "", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if mapping1.ShortURL != mapping2.ShortURL {
		t.Error("Expected same short URL for duplicate long URL")
	}
}

func TestCreateShortURL_WithTTL(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"
	ttl := 1 * time.Second

	mapping, err := service.CreateShortURL(longURL, "", ttl)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if mapping.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}

	expectedExpiry := time.Now().Add(ttl)
	if mapping.ExpiresAt.Before(expectedExpiry.Add(-100*time.Millisecond)) ||
		mapping.ExpiresAt.After(expectedExpiry.Add(100*time.Millisecond)) {
		t.Errorf("Expected ExpiresAt around %v, got %v", expectedExpiry, mapping.ExpiresAt)
	}
}

func TestGetLongURL_Success(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"

	created, _ := service.CreateShortURL(longURL, "", 0)

	retrieved, err := service.GetLongURL(created.ShortURL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.LongURL != longURL {
		t.Errorf("Expected long URL %s, got %s", longURL, retrieved.LongURL)
	}

	if retrieved.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", retrieved.AccessCount)
	}
}

func TestGetLongURL_NotFound(t *testing.T) {
	service := NewTinyURLService("http://test.com")

	_, err := service.GetLongURL("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent short URL")
	}
}

func TestGetLongURL_Expired(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"
	ttl := 1 * time.Millisecond

	created, _ := service.CreateShortURL(longURL, "", ttl)

	time.Sleep(10 * time.Millisecond)

	_, err := service.GetLongURL(created.ShortURL)
	if err == nil {
		t.Error("Expected error for expired short URL")
	}
}

func TestDeleteShortURL_Success(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"

	created, _ := service.CreateShortURL(longURL, "", 0)

	err := service.DeleteShortURL(created.ShortURL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = service.GetLongURL(created.ShortURL)
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestDeleteShortURL_NotFound(t *testing.T) {
	service := NewTinyURLService("http://test.com")

	err := service.DeleteShortURL("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent short URL")
	}
}

func TestGetStats(t *testing.T) {
	service := NewTinyURLService("http://test.com")
	longURL := "https://example.com"

	created, _ := service.CreateShortURL(longURL, "", 0)

	// Access the URL a few times
	service.GetLongURL(created.ShortURL)
	service.GetLongURL(created.ShortURL)
	service.GetLongURL(created.ShortURL)

	stats, err := service.GetStats(created.ShortURL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats.AccessCount != 3 {
		t.Errorf("Expected access count 3, got %d", stats.AccessCount)
	}
}

func TestListAllMappings(t *testing.T) {
	service := NewTinyURLService("http://test.com")

	service.CreateShortURL("https://example1.com", "", 0)
	service.CreateShortURL("https://example2.com", "", 0)
	service.CreateShortURL("https://example3.com", "", 0)

	mappings := service.ListAllMappings()
	if len(mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(mappings))
	}
}

func TestCreateHandler_Success(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	reqBody := map[string]interface{}{
		"long_url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	createHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var mapping URLMapping
	json.NewDecoder(w.Body).Decode(&mapping)

	if mapping.LongURL != "https://example.com" {
		t.Errorf("Expected long URL https://example.com, got %s", mapping.LongURL)
	}
}

func TestCreateHandler_InvalidMethod(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	req := httptest.NewRequest(http.MethodGet, "/create", nil)
	w := httptest.NewRecorder()

	createHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCreateHandler_MissingLongURL(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	createHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateHandler_CustomAlias(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	reqBody := map[string]interface{}{
		"long_url":     "https://example.com",
		"custom_alias": "myalias",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	createHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var mapping URLMapping
	json.NewDecoder(w.Body).Decode(&mapping)

	if mapping.ShortURL != "myalias" {
		t.Errorf("Expected short URL myalias, got %s", mapping.ShortURL)
	}
}

func TestRedirectHandler_Success(t *testing.T) {
	service = NewTinyURLService("http://test.com")
	longURL := "https://example.com"

	created, _ := service.CreateShortURL(longURL, "test123", 0)

	req := httptest.NewRequest(http.MethodGet, "/"+created.ShortURL, nil)
	w := httptest.NewRecorder()

	redirectHandler(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 301, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != longURL {
		t.Errorf("Expected location %s, got %s", longURL, location)
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	redirectHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestStatsHandler_Success(t *testing.T) {
	service = NewTinyURLService("http://test.com")
	created, _ := service.CreateShortURL("https://example.com", "test123", 0)

	req := httptest.NewRequest(http.MethodGet, "/stats?short_url="+created.ShortURL, nil)
	w := httptest.NewRecorder()

	statsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestStatsHandler_MissingParameter(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	statsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDeleteHandler_Success(t *testing.T) {
	service = NewTinyURLService("http://test.com")
	created, _ := service.CreateShortURL("https://example.com", "test123", 0)

	req := httptest.NewRequest(http.MethodDelete, "/delete?short_url="+created.ShortURL, nil)
	w := httptest.NewRecorder()

	deleteHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDeleteHandler_InvalidMethod(t *testing.T) {
	service = NewTinyURLService("http://test.com")

	req := httptest.NewRequest(http.MethodGet, "/delete", nil)
	w := httptest.NewRecorder()

	deleteHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestListHandler(t *testing.T) {
	service = NewTinyURLService("http://test.com")
	service.CreateShortURL("https://example1.com", "", 0)
	service.CreateShortURL("https://example2.com", "", 0)

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	w := httptest.NewRecorder()

	listHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var mappings []*URLMapping
	json.NewDecoder(w.Body).Decode(&mappings)

	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(mappings))
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %s", response["status"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	service = NewTinyURLService("http://test.com")
	created, _ := service.CreateShortURL("https://example.com", "concurrent", 0)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			service.GetLongURL(created.ShortURL)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	stats, _ := service.GetStats(created.ShortURL)
	if stats.AccessCount != 100 {
		t.Errorf("Expected access count 100, got %d", stats.AccessCount)
	}
}

