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
}

func TestGetUserChats(t *testing.T) {
	service := NewMessagingService()
	service.SendMessage("user1", "user2", "Hello")
	
	chats, err := service.GetUserChats("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chats))
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
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	healthHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

