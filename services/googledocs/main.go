package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Document represents a collaborative document
type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	Editors   []string  `json:"editors"`
}

// Edit represents an edit operation
type Edit struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	UserID     string    `json:"user_id"`
	Operation  string    `json:"operation"` // insert, delete, replace
	Position   int       `json:"position"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

// GoogleDocsService manages documents and collaborative editing
type GoogleDocsService struct {
	mu        sync.RWMutex
	documents map[string]*Document
	edits     map[string][]*Edit // documentID -> []Edit
	docIndex  int64
	editIndex int64
}

// NewGoogleDocsService creates a new Google Docs service
func NewGoogleDocsService() *GoogleDocsService {
	return &GoogleDocsService{
		documents: make(map[string]*Document),
		edits:     make(map[string][]*Edit),
	}
}

// CreateDocument creates a new document
func (s *GoogleDocsService) CreateDocument(title, ownerID string) (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.docIndex++
	docID := generateID("doc", s.docIndex)

	doc := &Document{
		ID:        docID,
		Title:     title,
		Content:   "",
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
		Editors:   []string{ownerID},
	}

	s.documents[docID] = doc
	s.edits[docID] = []*Edit{}

	return doc, nil
}

// GetDocument retrieves a document
func (s *GoogleDocsService) GetDocument(docID string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.documents[docID]
	if !exists {
		return nil, nil
	}

	return doc, nil
}

// EditDocument edits a document
func (s *GoogleDocsService) EditDocument(docID, userID, operation, content string, position int) (*Edit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, exists := s.documents[docID]
	if !exists {
		return nil, nil
	}

	s.editIndex++
	editID := generateID("edit", s.editIndex)

	edit := &Edit{
		ID:         editID,
		DocumentID: docID,
		UserID:     userID,
		Operation:  operation,
		Position:   position,
		Content:    content,
		Timestamp:  time.Now(),
	}

	// Apply edit to document
	switch operation {
	case "insert":
		if position <= len(doc.Content) {
			doc.Content = doc.Content[:position] + content + doc.Content[position:]
		}
	case "delete":
		if position < len(doc.Content) {
			endPos := position + len(content)
			if endPos > len(doc.Content) {
				endPos = len(doc.Content)
			}
			doc.Content = doc.Content[:position] + doc.Content[endPos:]
		}
	case "replace":
		doc.Content = content
	}

	doc.UpdatedAt = time.Now()
	doc.Version++

	s.edits[docID] = append(s.edits[docID], edit)

	return edit, nil
}

// ShareDocument shares a document with another user
func (s *GoogleDocsService) ShareDocument(docID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, exists := s.documents[docID]
	if !exists {
		return nil
	}

	// Check if user is already an editor
	for _, editor := range doc.Editors {
		if editor == userID {
			return nil
		}
	}

	doc.Editors = append(doc.Editors, userID)
	return nil
}

// GetEditHistory retrieves edit history for a document
func (s *GoogleDocsService) GetEditHistory(docID string) ([]*Edit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	edits, exists := s.edits[docID]
	if !exists {
		return []*Edit{}, nil
	}

	return edits, nil
}

func generateID(prefix string, index int64) string {
	return prefix + "_" + string(rune(index+'0'))
}

var service *GoogleDocsService

func createDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title   string `json:"title"`
		OwnerID string `json:"owner_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doc, err := service.CreateDocument(req.Title, req.OwnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func getDocumentHandler(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("doc_id")
	if docID == "" {
		http.Error(w, "doc_id parameter is required", http.StatusBadRequest)
		return
	}

	doc, err := service.GetDocument(docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if doc == nil {
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func editDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DocumentID string `json:"document_id"`
		UserID     string `json:"user_id"`
		Operation  string `json:"operation"`
		Content    string `json:"content"`
		Position   int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	edit, err := service.EditDocument(req.DocumentID, req.UserID, req.Operation, req.Content, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(edit)
}

func shareDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DocumentID string `json:"document_id"`
		UserID     string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.ShareDocument(req.DocumentID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getEditHistoryHandler(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("doc_id")
	if docID == "" {
		http.Error(w, "doc_id parameter is required", http.StatusBadRequest)
		return
	}

	edits, err := service.GetEditHistory(docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(edits)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewGoogleDocsService()

	http.HandleFunc("/document/create", createDocumentHandler)
	http.HandleFunc("/document/get", getDocumentHandler)
	http.HandleFunc("/document/edit", editDocumentHandler)
	http.HandleFunc("/document/share", shareDocumentHandler)
	http.HandleFunc("/document/history", getEditHistoryHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8087"
	log.Printf("Google Docs service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

