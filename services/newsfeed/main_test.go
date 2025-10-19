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

func TestNewNewsfeedService(t *testing.T) {
	service := NewNewsfeedService()
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if len(service.users) != 0 {
		t.Errorf("Expected empty users, got %d", len(service.users))
	}
}

func TestCreateUser(t *testing.T) {
	service := NewNewsfeedService()
	user, err := service.CreateUser("user1", "testuser")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != "user1" {
		t.Errorf("Expected user ID user1, got %s", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", user.Username)
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	_, err := service.CreateUser("user1", "testuser2")
	if err == nil {
		t.Error("Expected error for duplicate user")
	}
}

func TestGetUser(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	user, err := service.GetUser("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != "user1" {
		t.Errorf("Expected user ID user1, got %s", user.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	service := NewNewsfeedService()
	_, err := service.GetUser("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestFollow(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")

	err := service.Follow("user1", "user2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	user1, _ := service.GetUser("user1")
	if len(user1.Following) != 1 || user1.Following[0] != "user2" {
		t.Error("Expected user1 to follow user2")
	}

	user2, _ := service.GetUser("user2")
	if len(user2.Followers) != 1 || user2.Followers[0] != "user1" {
		t.Error("Expected user2 to have user1 as follower")
	}
}

func TestFollow_AlreadyFollowing(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")
	service.Follow("user1", "user2")

	err := service.Follow("user1", "user2")
	if err == nil {
		t.Error("Expected error for already following")
	}
}

func TestUnfollow(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")
	service.Follow("user1", "user2")

	err := service.Unfollow("user1", "user2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	user1, _ := service.GetUser("user1")
	if len(user1.Following) != 0 {
		t.Error("Expected user1 to not follow anyone")
	}
}

func TestUnfollow_NotFollowing(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")

	err := service.Unfollow("user1", "user2")
	if err == nil {
		t.Error("Expected error for not following")
	}
}

func TestCreatePost(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")

	post, err := service.CreatePost("user1", "Hello World")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if post.UserID != "user1" {
		t.Errorf("Expected user ID user1, got %s", post.UserID)
	}
	if post.Content != "Hello World" {
		t.Errorf("Expected content 'Hello World', got %s", post.Content)
	}
}

func TestCreatePost_UserNotFound(t *testing.T) {
	service := NewNewsfeedService()
	_, err := service.CreatePost("nonexistent", "Hello")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestLikePost(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	post, _ := service.CreatePost("user1", "Hello")

	err := service.LikePost(post.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := service.GetPost(post.ID)
	if updated.Likes != 1 {
		t.Errorf("Expected 1 like, got %d", updated.Likes)
	}
}

func TestCommentPost(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	post, _ := service.CreatePost("user1", "Hello")

	err := service.CommentPost(post.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := service.GetPost(post.ID)
	if updated.Comments != 1 {
		t.Errorf("Expected 1 comment, got %d", updated.Comments)
	}
}

func TestSharePost(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	post, _ := service.CreatePost("user1", "Hello")

	err := service.SharePost(post.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := service.GetPost(post.ID)
	if updated.Shares != 1 {
		t.Errorf("Expected 1 share, got %d", updated.Shares)
	}
}

func TestGetUserPosts(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	service.CreatePost("user1", "Post 1")
	service.CreatePost("user1", "Post 2")

	posts, err := service.GetUserPosts("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}
}

func TestGetNewsfeed(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")
	service.Follow("user1", "user2")
	service.CreatePost("user2", "Post from user2")

	feed, err := service.GetNewsfeed("user1", 50)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(feed) != 1 {
		t.Errorf("Expected 1 post in feed, got %d", len(feed))
	}
}

func TestDeletePost(t *testing.T) {
	service := NewNewsfeedService()
	service.CreateUser("user1", "testuser")
	post, _ := service.CreatePost("user1", "Hello")

	err := service.DeletePost(post.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = service.GetPost(post.ID)
	if err == nil {
		t.Error("Expected error for deleted post")
	}
}

func TestCreateUserHandler(t *testing.T) {
	service = NewNewsfeedService()

	reqBody := map[string]interface{}{
		"user_id":  "user1",
		"username": "testuser",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	createUserHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetUserHandler(t *testing.T) {
	service = NewNewsfeedService()
	service.CreateUser("user1", "testuser")

	req := httptest.NewRequest(http.MethodGet, "/user/get?user_id=user1", nil)
	w := httptest.NewRecorder()

	getUserHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestFollowHandler(t *testing.T) {
	service = NewNewsfeedService()
	service.CreateUser("user1", "testuser1")
	service.CreateUser("user2", "testuser2")

	reqBody := map[string]interface{}{
		"follower_id": "user1",
		"followee_id": "user2",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/user/follow", bytes.NewReader(body))
	w := httptest.NewRecorder()

	followHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreatePostHandler(t *testing.T) {
	service = NewNewsfeedService()
	service.CreateUser("user1", "testuser")

	reqBody := map[string]interface{}{
		"user_id": "user1",
		"content": "Hello World",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/post/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	createPostHandler(w, req)

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

