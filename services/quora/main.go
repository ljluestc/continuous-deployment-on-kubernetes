package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Question represents a question on Quora
type Question struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	Views       int64     `json:"views"`
	Upvotes     int64     `json:"upvotes"`
	Downvotes   int64     `json:"downvotes"`
}

// Answer represents an answer to a question
type Answer struct {
	ID         string    `json:"id"`
	QuestionID string    `json:"question_id"`
	UserID     string    `json:"user_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Upvotes    int64     `json:"upvotes"`
	Downvotes  int64     `json:"downvotes"`
}

// QuoraService manages questions and answers
type QuoraService struct {
	mu             sync.RWMutex
	questions      map[string]*Question
	answers        map[string]*Answer
	questionIndex  int64
	answerIndex    int64
	questionsByTag map[string][]string // tag -> []questionID
	answersByQ     map[string][]string // questionID -> []answerID
}

// NewQuoraService creates a new Quora service
func NewQuoraService() *QuoraService {
	return &QuoraService{
		questions:      make(map[string]*Question),
		answers:        make(map[string]*Answer),
		questionsByTag: make(map[string][]string),
		answersByQ:     make(map[string][]string),
	}
}

// CreateQuestion creates a new question
func (s *QuoraService) CreateQuestion(userID, title, description string, tags []string) (*Question, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.questionIndex++
	qID := generateID("q", s.questionIndex)

	question := &Question{
		ID:          qID,
		UserID:      userID,
		Title:       title,
		Description: description,
		Tags:        tags,
		CreatedAt:   time.Now(),
		Views:       0,
		Upvotes:     0,
		Downvotes:   0,
	}

	s.questions[qID] = question
	s.answersByQ[qID] = []string{}

	// Index by tags
	for _, tag := range tags {
		s.questionsByTag[tag] = append(s.questionsByTag[tag], qID)
	}

	return question, nil
}

// GetQuestion retrieves a question
func (s *QuoraService) GetQuestion(questionID string) (*Question, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	question, exists := s.questions[questionID]
	if !exists {
		return nil, nil
	}

	// Increment views
	question.Views++

	return question, nil
}

// CreateAnswer creates a new answer
func (s *QuoraService) CreateAnswer(questionID, userID, content string) (*Answer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.questions[questionID]; !exists {
		return nil, nil
	}

	s.answerIndex++
	aID := generateID("a", s.answerIndex)

	answer := &Answer{
		ID:         aID,
		QuestionID: questionID,
		UserID:     userID,
		Content:    content,
		CreatedAt:  time.Now(),
		Upvotes:    0,
		Downvotes:  0,
	}

	s.answers[aID] = answer
	s.answersByQ[questionID] = append(s.answersByQ[questionID], aID)

	return answer, nil
}

// GetAnswers retrieves all answers for a question
func (s *QuoraService) GetAnswers(questionID string) ([]*Answer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	answerIDs, exists := s.answersByQ[questionID]
	if !exists {
		return []*Answer{}, nil
	}

	answers := make([]*Answer, 0, len(answerIDs))
	for _, aID := range answerIDs {
		if answer, exists := s.answers[aID]; exists {
			answers = append(answers, answer)
		}
	}

	return answers, nil
}

// UpvoteQuestion upvotes a question
func (s *QuoraService) UpvoteQuestion(questionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	question, exists := s.questions[questionID]
	if !exists {
		return nil
	}

	question.Upvotes++
	return nil
}

// UpvoteAnswer upvotes an answer
func (s *QuoraService) UpvoteAnswer(answerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	answer, exists := s.answers[answerID]
	if !exists {
		return nil
	}

	answer.Upvotes++
	return nil
}

// SearchByTag searches questions by tag
func (s *QuoraService) SearchByTag(tag string) ([]*Question, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	questionIDs, exists := s.questionsByTag[tag]
	if !exists {
		return []*Question{}, nil
	}

	questions := make([]*Question, 0, len(questionIDs))
	for _, qID := range questionIDs {
		if question, exists := s.questions[qID]; exists {
			questions = append(questions, question)
		}
	}

	return questions, nil
}

func generateID(prefix string, index int64) string {
	return prefix + "_" + string(rune(index+'0'))
}

var service *QuoraService

func createQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID      string   `json:"user_id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := service.CreateQuestion(req.UserID, req.Title, req.Description, req.Tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func getQuestionHandler(w http.ResponseWriter, r *http.Request) {
	questionID := r.URL.Query().Get("question_id")
	if questionID == "" {
		http.Error(w, "question_id parameter is required", http.StatusBadRequest)
		return
	}

	question, err := service.GetQuestion(questionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if question == nil {
		http.Error(w, "question not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func createAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QuestionID string `json:"question_id"`
		UserID     string `json:"user_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	answer, err := service.CreateAnswer(req.QuestionID, req.UserID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func getAnswersHandler(w http.ResponseWriter, r *http.Request) {
	questionID := r.URL.Query().Get("question_id")
	if questionID == "" {
		http.Error(w, "question_id parameter is required", http.StatusBadRequest)
		return
	}

	answers, err := service.GetAnswers(questionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answers)
}

func upvoteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QuestionID string `json:"question_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.UpvoteQuestion(req.QuestionID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func searchByTagHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "tag parameter is required", http.StatusBadRequest)
		return
	}

	questions, err := service.SearchByTag(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewQuoraService()

	http.HandleFunc("/question/create", createQuestionHandler)
	http.HandleFunc("/question/get", getQuestionHandler)
	http.HandleFunc("/question/upvote", upvoteQuestionHandler)
	http.HandleFunc("/answer/create", createAnswerHandler)
	http.HandleFunc("/answer/list", getAnswersHandler)
	http.HandleFunc("/search", searchByTagHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8088"
	log.Printf("Quora service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

