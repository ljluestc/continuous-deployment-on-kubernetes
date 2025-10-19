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

func TestNewWebCrawlerService(t *testing.T) {
	service := NewWebCrawlerService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestCreateCrawlJob(t *testing.T) {
	service := NewWebCrawlerService()
	job, err := service.CreateCrawlJob("https://example.com", 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if job.URL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", job.URL)
	}
	if job.Status != "pending" {
		t.Errorf("Expected status pending, got %s", job.Status)
	}
}

func TestGetJob(t *testing.T) {
	service := NewWebCrawlerService()
	job, _ := service.CreateCrawlJob("https://example.com", 2)
	
	// Wait for job to start
	time.Sleep(100 * time.Millisecond)
	
	retrieved, err := service.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.ID != job.ID {
		t.Errorf("Expected job ID %s, got %s", job.ID, retrieved.ID)
	}
}

func TestGetPage(t *testing.T) {
	service := NewWebCrawlerService()
	service.CreateCrawlJob("https://example.com", 1)
	
	// Wait for crawling to complete
	time.Sleep(200 * time.Millisecond)
	
	page, err := service.GetPage("https://example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if page == nil {
		t.Fatal("Expected page to be found")
	}
}

func TestListPages(t *testing.T) {
	service := NewWebCrawlerService()
	service.CreateCrawlJob("https://example.com", 1)
	
	// Wait for crawling to complete
	time.Sleep(200 * time.Millisecond)
	
	pages := service.ListPages()
	if len(pages) == 0 {
		t.Error("Expected at least one page")
	}
}

func TestCreateJobHandler(t *testing.T) {
	service = NewWebCrawlerService()
	
	reqBody := map[string]interface{}{
		"url":   "https://example.com",
		"depth": 2,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/crawl", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	createJobHandler(w, req)
	
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

