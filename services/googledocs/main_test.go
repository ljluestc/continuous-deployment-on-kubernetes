//go:build unit
// +build unit

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGoogleDocsService(t *testing.T) {
	service := NewGoogleDocsService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestCreateDocument(t *testing.T) {
	service := NewGoogleDocsService()
	doc, err := service.CreateDocument("Test Doc", "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if doc.Title != "Test Doc" {
		t.Errorf("Expected title 'Test Doc', got %s", doc.Title)
	}
}

func TestGetDocument(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	retrieved, err := service.GetDocument(doc.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.ID != doc.ID {
		t.Errorf("Expected doc ID %s, got %s", doc.ID, retrieved.ID)
	}
}

func TestEditDocument_Insert(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	edit, err := service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if edit.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", edit.Content)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "Hello" {
		t.Errorf("Expected document content 'Hello', got %s", updated.Content)
	}
}

func TestEditDocument_Delete(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello World", 0)
	
	service.EditDocument(doc.ID, "user1", "delete", "World", 6)
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "Hello " {
		t.Errorf("Expected document content 'Hello ', got %s", updated.Content)
	}
}

func TestShareDocument(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	err := service.ShareDocument(doc.ID, "user2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if len(updated.Editors) != 2 {
		t.Errorf("Expected 2 editors, got %d", len(updated.Editors))
	}
}

func TestGetEditHistory(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	service.EditDocument(doc.ID, "user1", "insert", " World", 5)
	
	edits, err := service.GetEditHistory(doc.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(edits) != 2 {
		t.Errorf("Expected 2 edits, got %d", len(edits))
	}
}

func TestCreateDocumentHandler(t *testing.T) {
	service = NewGoogleDocsService()
	
	reqBody := map[string]interface{}{
		"title":    "Test Doc",
		"owner_id": "user1",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/document/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	createDocumentHandler(w, req)
	
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

