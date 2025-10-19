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

func TestNewTrie(t *testing.T) {
	trie := NewTrie()
	if trie == nil {
		t.Fatal("Expected trie to be created")
	}
	if trie.root == nil {
		t.Fatal("Expected root node to be created")
	}
}

func TestTrie_Insert(t *testing.T) {
	trie := NewTrie()
	trie.Insert("apple", 100)

	// Verify the word can be found
	results := trie.Search("app", 10)
	if len(results) != 1 || results[0] != "apple" {
		t.Errorf("Expected to find 'apple', got %v", results)
	}
}

func TestTrie_Search_NoResults(t *testing.T) {
	trie := NewTrie()
	trie.Insert("apple", 100)

	results := trie.Search("ban", 10)
	if len(results) != 0 {
		t.Errorf("Expected no results, got %v", results)
	}
}

func TestTrie_Search_MultipleResults(t *testing.T) {
	trie := NewTrie()
	trie.Insert("apple", 100)
	trie.Insert("application", 90)
	trie.Insert("apply", 80)

	results := trie.Search("app", 10)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Results should be sorted by score
	if results[0] != "apple" {
		t.Errorf("Expected first result to be 'apple', got %s", results[0])
	}
}

func TestTrie_Search_WithLimit(t *testing.T) {
	trie := NewTrie()
	trie.Insert("apple", 100)
	trie.Insert("application", 90)
	trie.Insert("apply", 80)

	results := trie.Search("app", 2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestTrie_Delete(t *testing.T) {
	trie := NewTrie()
	trie.Insert("apple", 100)

	deleted := trie.Delete("apple")
	if !deleted {
		t.Error("Expected word to be deleted")
	}

	results := trie.Search("app", 10)
	if len(results) != 0 {
		t.Errorf("Expected no results after deletion, got %v", results)
	}
}

func TestTrie_Delete_NonExistent(t *testing.T) {
	trie := NewTrie()
	deleted := trie.Delete("nonexistent")
	if deleted {
		t.Error("Expected deletion to fail for non-existent word")
	}
}

func TestTrie_CaseInsensitive(t *testing.T) {
	trie := NewTrie()
	trie.Insert("Apple", 100)

	results := trie.Search("app", 10)
	if len(results) != 1 || results[0] != "Apple" {
		t.Errorf("Expected to find 'Apple' with lowercase search, got %v", results)
	}
}

func TestNewTypeaheadService(t *testing.T) {
	service := NewTypeaheadService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if service.trie == nil {
		t.Fatal("Expected trie to be created")
	}
}

func TestTypeaheadService_AddWord(t *testing.T) {
	service := NewTypeaheadService()
	service.AddWord("test", 100)

	suggestions := service.GetSuggestions("tes", 10)
	if len(suggestions) != 1 || suggestions[0] != "test" {
		t.Errorf("Expected to find 'test', got %v", suggestions)
	}
}

func TestTypeaheadService_GetSuggestions(t *testing.T) {
	service := NewTypeaheadService()
	service.AddWord("apple", 100)
	service.AddWord("application", 90)

	suggestions := service.GetSuggestions("app", 10)
	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestTypeaheadService_DeleteWord(t *testing.T) {
	service := NewTypeaheadService()
	service.AddWord("test", 100)

	deleted := service.DeleteWord("test")
	if !deleted {
		t.Error("Expected word to be deleted")
	}

	suggestions := service.GetSuggestions("tes", 10)
	if len(suggestions) != 0 {
		t.Errorf("Expected no suggestions after deletion, got %v", suggestions)
	}
}

func TestAddWordHandler(t *testing.T) {
	service = NewTypeaheadService()

	reqBody := map[string]interface{}{
		"word":  "test",
		"score": 100,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	w := httptest.NewRecorder()

	addWordHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAddWordHandler_InvalidMethod(t *testing.T) {
	service = NewTypeaheadService()

	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	w := httptest.NewRecorder()

	addWordHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSuggestHandler(t *testing.T) {
	service = NewTypeaheadService()
	service.AddWord("apple", 100)

	req := httptest.NewRequest(http.MethodGet, "/suggest?prefix=app", nil)
	w := httptest.NewRecorder()

	suggestHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	suggestions := response["suggestions"].([]interface{})
	if len(suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(suggestions))
	}
}

func TestSuggestHandler_MissingPrefix(t *testing.T) {
	service = NewTypeaheadService()

	req := httptest.NewRequest(http.MethodGet, "/suggest", nil)
	w := httptest.NewRecorder()

	suggestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDeleteWordHandler(t *testing.T) {
	service = NewTypeaheadService()
	service.AddWord("test", 100)

	req := httptest.NewRequest(http.MethodDelete, "/delete?word=test", nil)
	w := httptest.NewRecorder()

	deleteWordHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteWordHandler_InvalidMethod(t *testing.T) {
	service = NewTypeaheadService()

	req := httptest.NewRequest(http.MethodGet, "/delete", nil)
	w := httptest.NewRecorder()

	deleteWordHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestDeleteWordHandler_NotFound(t *testing.T) {
	service = NewTypeaheadService()

	req := httptest.NewRequest(http.MethodDelete, "/delete?word=nonexistent", nil)
	w := httptest.NewRecorder()

	deleteWordHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
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

