package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Message represents a message in the system
type Message struct {
	ID          string    `json:"id"`
	FromUserID  string    `json:"from_user_id"`
	ToUserID    string    `json:"to_user_id"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	Read        bool      `json:"read"`
	ChatID      string    `json:"chat_id"`
}

// Chat represents a conversation between users
type Chat struct {
	ID       string   `json:"id"`
	UserIDs  []string `json:"user_ids"`
	Messages []string `json:"messages"` // message IDs
}

// MessagingService manages messages and chats
type MessagingService struct {
	mu           sync.RWMutex
	messages     map[string]*Message
	chats        map[string]*Chat
	userChats    map[string][]string // userID -> []chatID
	messageIndex int64
	chatIndex    int64
}

// NewMessagingService creates a new messaging service
func NewMessagingService() *MessagingService {
	return &MessagingService{
		messages:  make(map[string]*Message),
		chats:     make(map[string]*Chat),
		userChats: make(map[string][]string),
	}
}

// SendMessage sends a message
func (s *MessagingService) SendMessage(fromUserID, toUserID, content string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find or create chat
	chatID := s.findOrCreateChat(fromUserID, toUserID)

	s.messageIndex++
	messageID := generateID("msg", s.messageIndex)

	message := &Message{
		ID:         messageID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		Timestamp:  time.Now(),
		Read:       false,
		ChatID:     chatID,
	}

	s.messages[messageID] = message
	s.chats[chatID].Messages = append(s.chats[chatID].Messages, messageID)

	return message, nil
}

// findOrCreateChat finds or creates a chat between two users
func (s *MessagingService) findOrCreateChat(user1ID, user2ID string) string {
	// Check if chat already exists
	for _, chatID := range s.userChats[user1ID] {
		chat := s.chats[chatID]
		if contains(chat.UserIDs, user2ID) {
			return chatID
		}
	}

	// Create new chat
	s.chatIndex++
	chatID := generateID("chat", s.chatIndex)

	chat := &Chat{
		ID:       chatID,
		UserIDs:  []string{user1ID, user2ID},
		Messages: []string{},
	}

	s.chats[chatID] = chat
	s.userChats[user1ID] = append(s.userChats[user1ID], chatID)
	s.userChats[user2ID] = append(s.userChats[user2ID], chatID)

	return chatID
}

// GetMessages retrieves messages for a chat
func (s *MessagingService) GetMessages(chatID string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, exists := s.chats[chatID]
	if !exists {
		return nil, nil
	}

	messages := make([]*Message, 0, len(chat.Messages))
	for _, msgID := range chat.Messages {
		if msg, exists := s.messages[msgID]; exists {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// GetUserChats retrieves all chats for a user
func (s *MessagingService) GetUserChats(userID string) ([]*Chat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chatIDs, exists := s.userChats[userID]
	if !exists {
		return []*Chat{}, nil
	}

	chats := make([]*Chat, 0, len(chatIDs))
	for _, chatID := range chatIDs {
		if chat, exists := s.chats[chatID]; exists {
			chats = append(chats, chat)
		}
	}

	return chats, nil
}

// MarkAsRead marks a message as read
func (s *MessagingService) MarkAsRead(messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	message, exists := s.messages[messageID]
	if !exists {
		return nil
	}

	message.Read = true
	return nil
}

// Helper functions
func generateID(prefix string, index int64) string {
	return prefix + "_" + string(rune(index+'0'))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var service *MessagingService

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FromUserID string `json:"from_user_id"`
		ToUserID   string `json:"to_user_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := service.SendMessage(req.FromUserID, req.ToUserID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "chat_id parameter is required", http.StatusBadRequest)
		return
	}

	messages, err := service.GetMessages(chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func getUserChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	chats, err := service.GetUserChats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func markAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MessageID string `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.MarkAsRead(req.MessageID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewMessagingService()

	http.HandleFunc("/send", sendMessageHandler)
	http.HandleFunc("/messages", getMessagesHandler)
	http.HandleFunc("/chats", getUserChatsHandler)
	http.HandleFunc("/mark-read", markAsReadHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8084"
	log.Printf("Messaging service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

