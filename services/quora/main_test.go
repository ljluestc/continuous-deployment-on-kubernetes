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
	if service.questions == nil {
		t.Fatal("Expected questions map to be initialized")
	}
	if service.answers == nil {
		t.Fatal("Expected answers map to be initialized")
	}
	if service.questionsByTag == nil {
		t.Fatal("Expected questionsByTag map to be initialized")
	}
	if service.answersByQ == nil {
		t.Fatal("Expected answersByQ map to be initialized")
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
	if q.UserID != "user1" {
		t.Errorf("Expected user_id 'user1', got %s", q.UserID)
	}
	if q.Views != 0 {
		t.Errorf("Expected 0 views, got %d", q.Views)
	}
	if len(q.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(q.Tags))
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
	
	// Get again to increment views
	retrieved2, _ := service.GetQuestion(q.ID)
	if retrieved2.Views != 2 {
		t.Errorf("Expected 2 views, got %d", retrieved2.Views)
	}
}

func TestGetQuestion_NotFound(t *testing.T) {
	service := NewQuoraService()
	
	q, err := service.GetQuestion("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if q != nil {
		t.Errorf("Expected nil question, got %v", q)
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
	if a.QuestionID != q.ID {
		t.Errorf("Expected question_id %s, got %s", q.ID, a.QuestionID)
	}
	if a.Upvotes != 0 {
		t.Errorf("Expected 0 upvotes, got %d", a.Upvotes)
	}
}

func TestCreateAnswer_QuestionNotFound(t *testing.T) {
	service := NewQuoraService()
	
	a, err := service.CreateAnswer("nonexistent", "user2", "Test Answer")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if a != nil {
		t.Errorf("Expected nil answer, got %v", a)
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

func TestGetAnswers_NotFound(t *testing.T) {
	service := NewQuoraService()
	
	answers, err := service.GetAnswers("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(answers) != 0 {
		t.Errorf("Expected 0 answers, got %d", len(answers))
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

func TestUpvoteQuestion_NotFound(t *testing.T) {
	service := NewQuoraService()
	
	err := service.UpvoteQuestion("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent question, got %v", err)
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

func TestUpvoteAnswer_NotFound(t *testing.T) {
	service := NewQuoraService()
	
	err := service.UpvoteAnswer("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent answer, got %v", err)
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

func TestSearchByTag_NotFound(t *testing.T) {
	service := NewQuoraService()
	
	questions, err := service.SearchByTag("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("Expected 0 questions, got %d", len(questions))
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID("q", 1)
	if id == "" {
		t.Error("Expected non-empty ID")
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
	
	var q Question
	json.NewDecoder(w.Body).Decode(&q)
	if q.Title != "Test Question" {
		t.Errorf("Expected title 'Test Question', got %s", q.Title)
	}
}

func TestCreateQuestionHandler_InvalidMethod(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/question/create", nil)
	w := httptest.NewRecorder()
	
	createQuestionHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCreateQuestionHandler_InvalidJSON(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodPost, "/question/create", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	createQuestionHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetQuestionHandler(t *testing.T) {
	service = NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	req := httptest.NewRequest(http.MethodGet, "/question/get?question_id="+q.ID, nil)
	w := httptest.NewRecorder()
	
	getQuestionHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetQuestionHandler_MissingQuestionID(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/question/get", nil)
	w := httptest.NewRecorder()
	
	getQuestionHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetQuestionHandler_NotFound(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/question/get?question_id=nonexistent", nil)
	w := httptest.NewRecorder()
	
	getQuestionHandler(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateAnswerHandler(t *testing.T) {
	service = NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	reqBody := map[string]interface{}{
		"question_id": q.ID,
		"user_id":     "user2",
		"content":     "Test Answer",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/answer/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	createAnswerHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreateAnswerHandler_InvalidMethod(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/answer/create", nil)
	w := httptest.NewRecorder()
	
	createAnswerHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCreateAnswerHandler_InvalidJSON(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodPost, "/answer/create", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	createAnswerHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAnswersHandler(t *testing.T) {
	service = NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	service.CreateAnswer(q.ID, "user2", "Answer 1")
	
	req := httptest.NewRequest(http.MethodGet, "/answer/list?question_id="+q.ID, nil)
	w := httptest.NewRecorder()
	
	getAnswersHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var answers []*Answer
	json.NewDecoder(w.Body).Decode(&answers)
	if len(answers) != 1 {
		t.Errorf("Expected 1 answer, got %d", len(answers))
	}
}

func TestGetAnswersHandler_MissingQuestionID(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/answer/list", nil)
	w := httptest.NewRecorder()
	
	getAnswersHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpvoteQuestionHandler(t *testing.T) {
	service = NewQuoraService()
	q, _ := service.CreateQuestion("user1", "Test Question", "Description", []string{"go"})
	
	reqBody := map[string]interface{}{
		"question_id": q.ID,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/question/upvote", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	upvoteQuestionHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUpvoteQuestionHandler_InvalidMethod(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/question/upvote", nil)
	w := httptest.NewRecorder()
	
	upvoteQuestionHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestUpvoteQuestionHandler_InvalidJSON(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodPost, "/question/upvote", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	upvoteQuestionHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSearchByTagHandler(t *testing.T) {
	service = NewQuoraService()
	service.CreateQuestion("user1", "Go Question", "Description", []string{"go"})
	
	req := httptest.NewRequest(http.MethodGet, "/search?tag=go", nil)
	w := httptest.NewRecorder()
	
	searchByTagHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var questions []*Question
	json.NewDecoder(w.Body).Decode(&questions)
	if len(questions) != 1 {
		t.Errorf("Expected 1 question, got %d", len(questions))
	}
}

func TestSearchByTagHandler_MissingTag(t *testing.T) {
	service = NewQuoraService()
	
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	w := httptest.NewRecorder()
	
	searchByTagHandler(w, req)
	
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
