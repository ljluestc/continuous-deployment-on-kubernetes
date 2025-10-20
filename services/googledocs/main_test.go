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
	if service.documents == nil {
		t.Fatal("Expected documents map to be initialized")
	}
	if service.edits == nil {
		t.Fatal("Expected edits map to be initialized")
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
	if doc.OwnerID != "user1" {
		t.Errorf("Expected owner 'user1', got %s", doc.OwnerID)
	}
	if doc.Content != "" {
		t.Errorf("Expected empty content, got %s", doc.Content)
	}
	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}
	if len(doc.Editors) != 1 {
		t.Errorf("Expected 1 editor, got %d", len(doc.Editors))
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

func TestGetDocument_NotFound(t *testing.T) {
	service := NewGoogleDocsService()
	
	doc, err := service.GetDocument("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if doc != nil {
		t.Errorf("Expected nil document, got %v", doc)
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
	if updated.Version != 2 {
		t.Errorf("Expected version 2, got %d", updated.Version)
	}
}

func TestEditDocument_InsertMiddle(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "HelloWorld", 0)
	
	edit, err := service.EditDocument(doc.ID, "user1", "insert", " ", 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "Hello World" {
		t.Errorf("Expected document content 'Hello World', got %s", updated.Content)
	}
	if edit.Operation != "insert" {
		t.Errorf("Expected operation 'insert', got %s", edit.Operation)
	}
}

func TestEditDocument_InsertOutOfBounds(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	
	// Insert at position beyond content length
	service.EditDocument(doc.ID, "user1", "insert", " World", 100)
	
	updated, _ := service.GetDocument(doc.ID)
	// Should not insert if position is out of bounds
	if updated.Content != "Hello" {
		t.Errorf("Expected document content 'Hello', got %s", updated.Content)
	}
}

func TestEditDocument_Delete(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello World", 0)
	
	edit, err := service.EditDocument(doc.ID, "user1", "delete", "World", 6)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "Hello " {
		t.Errorf("Expected document content 'Hello ', got %s", updated.Content)
	}
	if edit.Operation != "delete" {
		t.Errorf("Expected operation 'delete', got %s", edit.Operation)
	}
}

func TestEditDocument_DeleteOutOfBounds(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	
	// Delete at position beyond content length
	service.EditDocument(doc.ID, "user1", "delete", "test", 100)
	
	updated, _ := service.GetDocument(doc.ID)
	// Should not delete if position is out of bounds
	if updated.Content != "Hello" {
		t.Errorf("Expected document content 'Hello', got %s", updated.Content)
	}
}

func TestEditDocument_DeletePartial(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	
	// Try to delete more than available
	service.EditDocument(doc.ID, "user1", "delete", "looooooo", 3)
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "Hel" {
		t.Errorf("Expected document content 'Hel', got %s", updated.Content)
	}
}

func TestEditDocument_Replace(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	
	edit, err := service.EditDocument(doc.ID, "user1", "replace", "New Content", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if updated.Content != "New Content" {
		t.Errorf("Expected document content 'New Content', got %s", updated.Content)
	}
	if edit.Operation != "replace" {
		t.Errorf("Expected operation 'replace', got %s", edit.Operation)
	}
}

func TestEditDocument_NotFound(t *testing.T) {
	service := NewGoogleDocsService()
	
	edit, err := service.EditDocument("nonexistent", "user1", "insert", "Hello", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if edit != nil {
		t.Errorf("Expected nil edit, got %v", edit)
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
	if updated.Editors[1] != "user2" {
		t.Errorf("Expected second editor to be 'user2', got %s", updated.Editors[1])
	}
}

func TestShareDocument_AlreadyEditor(t *testing.T) {
	service := NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	// Share with the same user twice
	service.ShareDocument(doc.ID, "user2")
	err := service.ShareDocument(doc.ID, "user2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	updated, _ := service.GetDocument(doc.ID)
	if len(updated.Editors) != 2 {
		t.Errorf("Expected 2 editors (no duplicate), got %d", len(updated.Editors))
	}
}

func TestShareDocument_NotFound(t *testing.T) {
	service := NewGoogleDocsService()
	
	err := service.ShareDocument("nonexistent", "user2")
	if err != nil {
		t.Errorf("Expected no error for non-existent document, got %v", err)
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

func TestGetEditHistory_NotFound(t *testing.T) {
	service := NewGoogleDocsService()
	
	edits, err := service.GetEditHistory("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(edits) != 0 {
		t.Errorf("Expected 0 edits, got %d", len(edits))
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID("doc", 1)
	if id == "" {
		t.Error("Expected non-empty ID")
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
	
	var doc Document
	json.NewDecoder(w.Body).Decode(&doc)
	if doc.Title != "Test Doc" {
		t.Errorf("Expected title 'Test Doc', got %s", doc.Title)
	}
}

func TestCreateDocumentHandler_InvalidMethod(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/create", nil)
	w := httptest.NewRecorder()
	
	createDocumentHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCreateDocumentHandler_InvalidJSON(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodPost, "/document/create", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	createDocumentHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetDocumentHandler(t *testing.T) {
	service = NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	req := httptest.NewRequest(http.MethodGet, "/document/get?doc_id="+doc.ID, nil)
	w := httptest.NewRecorder()
	
	getDocumentHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var retrieved Document
	json.NewDecoder(w.Body).Decode(&retrieved)
	if retrieved.ID != doc.ID {
		t.Errorf("Expected doc ID %s, got %s", doc.ID, retrieved.ID)
	}
}

func TestGetDocumentHandler_MissingDocID(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/get", nil)
	w := httptest.NewRecorder()
	
	getDocumentHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetDocumentHandler_NotFound(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/get?doc_id=nonexistent", nil)
	w := httptest.NewRecorder()
	
	getDocumentHandler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestEditDocumentHandler(t *testing.T) {
	service = NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	reqBody := map[string]interface{}{
		"document_id": doc.ID,
		"user_id":     "user1",
		"operation":   "insert",
		"content":     "Hello",
		"position":    0,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/document/edit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	editDocumentHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestEditDocumentHandler_InvalidMethod(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/edit", nil)
	w := httptest.NewRecorder()
	
	editDocumentHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestEditDocumentHandler_InvalidJSON(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodPost, "/document/edit", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	editDocumentHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestShareDocumentHandler(t *testing.T) {
	service = NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	
	reqBody := map[string]interface{}{
		"document_id": doc.ID,
		"user_id":     "user2",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/document/share", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	shareDocumentHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestShareDocumentHandler_InvalidMethod(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/share", nil)
	w := httptest.NewRecorder()
	
	shareDocumentHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestShareDocumentHandler_InvalidJSON(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodPost, "/document/share", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	shareDocumentHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetEditHistoryHandler(t *testing.T) {
	service = NewGoogleDocsService()
	doc, _ := service.CreateDocument("Test Doc", "user1")
	service.EditDocument(doc.ID, "user1", "insert", "Hello", 0)
	
	req := httptest.NewRequest(http.MethodGet, "/document/history?doc_id="+doc.ID, nil)
	w := httptest.NewRecorder()
	
	getEditHistoryHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var edits []*Edit
	json.NewDecoder(w.Body).Decode(&edits)
	if len(edits) != 1 {
		t.Errorf("Expected 1 edit, got %d", len(edits))
	}
}

func TestGetEditHistoryHandler_MissingDocID(t *testing.T) {
	service = NewGoogleDocsService()
	
	req := httptest.NewRequest(http.MethodGet, "/document/history", nil)
	w := httptest.NewRecorder()
	
	getEditHistoryHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
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
