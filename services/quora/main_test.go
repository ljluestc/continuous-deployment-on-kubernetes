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

func TestNewQuoraService(t *testing.T) {
	service := NewQuoraService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestCreateQuestion(t *testing.T) {
	service := NewQuoraService()
	q, err := service.CreateQuestion("user1", "Test Question", "Description", []string{"go", "testing"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if q.Title != "Test Question" {
		t.Errorf("Expected title 'Test Question', got %s", q.Title)
	}
}

func TestGetQuestion(t *testing.T) {
	service := NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	retrieved, err := service.GetQuestion(q.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.ID != q.ID {
		t.Errorf("Expected question ID %s, got %s", q.ID, retrieved.ID)
	}
	if retrieved.Views != 1 {
		t.Errorf("Expected 1 view, got %d", retrieved.Views)
	}
}

func TestCreateAnswer(t *testing.T) {
	service := NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	a, err := service.CreateAnswer(q.ID, "user2", "Test Answer")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if a.Content != "Test Answer" {
		t.Errorf("Expected content 'Test Answer', got %s", a.Content)
	}
}

func TestGetAnswers(t *testing.T) {
	service := NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	service.CreateAnswer(q.ID, "user2", "Answer 1")
	service.CreateAnswer(q.ID, "user3", "Answer 2")
	
	answers, err := service.GetAnswers(q.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(answers) != 2 {
		t.Errorf("Expected 2 answers, got %d", len(answers))
	}
}

func TestUpvoteQuestion(t *testing.T) {
	service := NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	err := service.UpvoteQuestion(q.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if service.questions[q.ID].Upvotes != 1 {
		t.Errorf("Expected 1 upvote, got %d", service.questions[q.ID].Upvotes)
	}
}

func TestUpvoteAnswer(t *testing.T) {
	service := NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	a, _ := service.CreateAnswer(q.ID, "user2", "Test Answer")
	
	err := service.UpvoteAnswer(a.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if service.answers[a.ID].Upvotes != 1 {
		t.Errorf("Expected 1 upvote, got %d", service.answers[a.ID].Upvotes)
	}
}

func TestSearchByTag(t *testing.T) {
	service := NewQuoraService()
	service.CreateQuestion("user1", "Go Question", "Description", []string{"go"})
	service.CreateQuestion("user1", "Python Question", "Description", []string{"python"})
	service.CreateQuestion("user1", "Another Go Question", "Description", []string{"go"})
	
	questions, err := service.SearchByTag("go")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(questions))
	}
}

func TestCreateQuestionHandler(t *testing.T) {
	service = NewQuoraService()
	
	reqBody := map[string]interface{}{
		"user_id":     "user1",
		"title":       "Test Question",
		"description": "Description",
		"tags":        []string{"go", "testing"},
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/question/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	createQuestionHandler(w, req)
	
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

