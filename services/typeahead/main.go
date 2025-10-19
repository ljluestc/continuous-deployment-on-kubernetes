package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// TrieNode represents a node in the trie
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	word     string
	score    int
}

// Trie represents a trie data structure
type Trie struct {
	root *TrieNode
	mu   sync.RWMutex
}

// NewTrie creates a new trie
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert inserts a word into the trie with a score
func (t *Trie) Insert(word string, score int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, ch := range strings.ToLower(word) {
		if node.children[ch] == nil {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.word = word
	node.score = score
}

// Search searches for words with a given prefix
func (t *Trie) Search(prefix string, limit int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	prefix = strings.ToLower(prefix)
	node := t.root

	// Navigate to the prefix
	for _, ch := range prefix {
		if node.children[ch] == nil {
			return []string{}
		}
		node = node.children[ch]
	}

	// Collect all words with this prefix
	results := []struct {
		word  string
		score int
	}{}
	t.collectWords(node, &results)

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Extract words and apply limit
	words := make([]string, 0, len(results))
	for i, r := range results {
		if limit > 0 && i >= limit {
			break
		}
		words = append(words, r.word)
	}

	return words
}

// collectWords collects all words from a node
func (t *Trie) collectWords(node *TrieNode, results *[]struct {
	word  string
	score int
}) {
	if node.isEnd {
		*results = append(*results, struct {
			word  string
			score int
		}{node.word, node.score})
	}

	for _, child := range node.children {
		t.collectWords(child, results)
	}
}

// Delete removes a word from the trie
func (t *Trie) Delete(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.deleteHelper(t.root, strings.ToLower(word), 0)
}

func (t *Trie) deleteHelper(node *TrieNode, word string, index int) bool {
	if index == len(word) {
		if !node.isEnd {
			return false
		}
		node.isEnd = false
		return len(node.children) == 0
	}

	ch := rune(word[index])
	child, exists := node.children[ch]
	if !exists {
		return false
	}

	shouldDeleteChild := t.deleteHelper(child, word, index+1)

	if shouldDeleteChild {
		delete(node.children, ch)
		return len(node.children) == 0 && !node.isEnd
	}

	return false
}

// TypeaheadService manages the typeahead functionality
type TypeaheadService struct {
	trie *Trie
}

// NewTypeaheadService creates a new typeahead service
func NewTypeaheadService() *TypeaheadService {
	return &TypeaheadService{
		trie: NewTrie(),
	}
}

// AddWord adds a word to the typeahead
func (s *TypeaheadService) AddWord(word string, score int) {
	s.trie.Insert(word, score)
}

// GetSuggestions returns suggestions for a prefix
func (s *TypeaheadService) GetSuggestions(prefix string, limit int) []string {
	return s.trie.Search(prefix, limit)
}

// DeleteWord deletes a word from the typeahead
func (s *TypeaheadService) DeleteWord(word string) bool {
	return s.trie.Delete(word)
}

var service *TypeaheadService

func addWordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Word  string `json:"word"`
		Score int    `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	service.AddWord(req.Word, req.Score)
	w.WriteHeader(http.StatusOK)
}

func suggestHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		http.Error(w, "prefix parameter is required", http.StatusBadRequest)
		return
	}

	limit := 10 // default limit
	suggestions := service.GetSuggestions(prefix, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suggestions": suggestions,
	})
}

func deleteWordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "word parameter is required", http.StatusBadRequest)
		return
	}

	if !service.DeleteWord(word) {
		http.Error(w, "word not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewTypeaheadService()

	// Add some sample words
	service.AddWord("apple", 100)
	service.AddWord("application", 90)
	service.AddWord("apply", 80)
	service.AddWord("banana", 70)

	http.HandleFunc("/add", addWordHandler)
	http.HandleFunc("/suggest", suggestHandler)
	http.HandleFunc("/delete", deleteWordHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8083"
	log.Printf("Typeahead service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

