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

func TestNewMessagingService(t *testing.T) {
	service := NewMessagingService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if service.messages == nil {
		t.Fatal("Expected messages map to be initialized")
	}
	if service.chats == nil {
		t.Fatal("Expected chats map to be initialized")
	}
	if service.userChats == nil {
		t.Fatal("Expected userChats map to be initialized")
	}
}

func TestSendMessage(t *testing.T) {
	service := NewMessagingService()
	msg, err := service.SendMessage("user1", "user2", "Hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if msg.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", msg.Content)
	}
	if msg.FromUserID != "user1" {
		t.Errorf("Expected from_user_id 'user1', got %s", msg.FromUserID)
	}
	if msg.ToUserID != "user2" {
		t.Errorf("Expected to_user_id 'user2', got %s", msg.ToUserID)
	}
	if msg.Read {
		t.Error("Expected message to not be read")
	}
}

func TestSendMessage_ReuseExistingChat(t *testing.T) {
	service := NewMessagingService()
	msg1, _ := service.SendMessage("user1", "user2", "Hello")
	msg2, _ := service.SendMessage("user1", "user2", "World")
	
	if msg1.ChatID != msg2.ChatID {
		t.Errorf("Expected same chat ID, got %s and %s", msg1.ChatID, msg2.ChatID)
	}
}

func TestSendMessage_ReverseDirection(t *testing.T) {
	service := NewMessagingService()
	msg1, _ := service.SendMessage("user1", "user2", "Hello")
	msg2, _ := service.SendMessage("user2", "user1", "Hi back")
	
	if msg1.ChatID != msg2.ChatID {
		t.Errorf("Expected same chat ID for reverse direction, got %s and %s", msg1.ChatID, msg2.ChatID)
	}
}

func TestGetMessages(t *testing.T) {
	service := NewMessagingService()
	msg, _ := service.SendMessage("user1", "user2", "Hello")
	
	messages, err := service.GetMessages(msg.ChatID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", messages[0].Content)
	}
}

func TestGetMessages_NotFound(t *testing.T) {
	service := NewMessagingService()
	
	messages, err := service.GetMessages("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if messages != nil {
		t.Errorf("Expected nil messages, got %v", messages)
	}
}

func TestGetUserChats(t *testing.T) {
	service := NewMessagingService()
	service.SendMessage("user1", "user2", "Hello")
	service.SendMessage("user1", "user3", "Hi")
	
	chats, err := service.GetUserChats("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(chats))
	}
}

func TestGetUserChats_NotFound(t *testing.T) {
	service := NewMessagingService()
	
	chats, err := service.GetUserChats("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chats) != 0 {
		t.Errorf("Expected 0 chats, got %d", len(chats))
	}
}

func TestMarkAsRead(t *testing.T) {
	service := NewMessagingService()
	msg, _ := service.SendMessage("user1", "user2", "Hello")
	
	err := service.MarkAsRead(msg.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !service.messages[msg.ID].Read {
		t.Error("Expected message to be marked as read")
	}
}

func TestMarkAsRead_NotFound(t *testing.T) {
	service := NewMessagingService()
	
	err := service.MarkAsRead("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent message, got %v", err)
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID("msg", 1)
	if id == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	
	if !contains(slice, "a") {
		t.Error("Expected to find 'a'")
	}
	if !contains(slice, "b") {
		t.Error("Expected to find 'b'")
	}
	if contains(slice, "d") {
		t.Error("Expected to not find 'd'")
	}
	if contains([]string{}, "a") {
		t.Error("Expected to not find in empty slice")
	}
}

func TestSendMessageHandler(t *testing.T) {
	service = NewMessagingService()
	
	reqBody := map[string]interface{}{
		"from_user_id": "user1",
		"to_user_id":   "user2",
		"content":      "Hello",
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	sendMessageHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var msg Message
	json.NewDecoder(w.Body).Decode(&msg)
	if msg.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", msg.Content)
	}
}

func TestSendMessageHandler_InvalidMethod(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodGet, "/send", nil)
	w := httptest.NewRecorder()
	
	sendMessageHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSendMessageHandler_InvalidJSON(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	sendMessageHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetMessagesHandler(t *testing.T) {
	service = NewMessagingService()
	msg, _ := service.SendMessage("user1", "user2", "Hello")
	
	req := httptest.NewRequest(http.MethodGet, "/messages?chat_id="+msg.ChatID, nil)
	w := httptest.NewRecorder()
	
	getMessagesHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var messages []*Message
	json.NewDecoder(w.Body).Decode(&messages)
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestGetMessagesHandler_MissingChatID(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()
	
	getMessagesHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetUserChatsHandler(t *testing.T) {
	service = NewMessagingService()
	service.SendMessage("user1", "user2", "Hello")
	
	req := httptest.NewRequest(http.MethodGet, "/chats?user_id=user1", nil)
	w := httptest.NewRecorder()
	
	getUserChatsHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var chats []*Chat
	json.NewDecoder(w.Body).Decode(&chats)
	if len(chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chats))
	}
}

func TestGetUserChatsHandler_MissingUserID(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	w := httptest.NewRecorder()
	
	getUserChatsHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestMarkAsReadHandler(t *testing.T) {
	service = NewMessagingService()
	msg, _ := service.SendMessage("user1", "user2", "Hello")
	
	reqBody := map[string]interface{}{
		"message_id": msg.ID,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/mark-read", bytes.NewReader(body))
	w := httptest.NewRecorder()
	
	markAsReadHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMarkAsReadHandler_InvalidMethod(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodGet, "/mark-read", nil)
	w := httptest.NewRecorder()
	
	markAsReadHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestMarkAsReadHandler_InvalidJSON(t *testing.T) {
	service = NewMessagingService()
	
	req := httptest.NewRequest(http.MethodPost, "/mark-read", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	markAsReadHandler(w, req)
	
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
