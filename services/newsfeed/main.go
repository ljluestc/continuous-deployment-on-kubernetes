package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Post represents a social media post
type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Likes     int64     `json:"likes"`
	Comments  int64     `json:"comments"`
	Shares    int64     `json:"shares"`
}

// User represents a user in the system
type User struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Following []string `json:"following"`
	Followers []string `json:"followers"`
}

// NewsfeedService manages posts and user relationships
type NewsfeedService struct {
	mu        sync.RWMutex
	posts     map[string]*Post
	users     map[string]*User
	userPosts map[string][]string // userID -> []postID
	postIndex int64
}

// NewNewsfeedService creates a new newsfeed service
func NewNewsfeedService() *NewsfeedService {
	return &NewsfeedService{
		posts:     make(map[string]*Post),
		users:     make(map[string]*User),
		userPosts: make(map[string][]string),
		postIndex: 0,
	}
}

// CreateUser creates a new user
func (s *NewsfeedService) CreateUser(userID, username string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[userID]; exists {
		return nil, fmt.Errorf("user already exists")
	}

	user := &User{
		ID:        userID,
		Username:  username,
		Following: []string{},
		Followers: []string{},
	}

	s.users[userID] = user
	s.userPosts[userID] = []string{}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *NewsfeedService) GetUser(userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Follow makes one user follow another
func (s *NewsfeedService) Follow(followerID, followeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	follower, exists := s.users[followerID]
	if !exists {
		return fmt.Errorf("follower not found")
	}

	followee, exists := s.users[followeeID]
	if !exists {
		return fmt.Errorf("followee not found")
	}

	// Check if already following
	for _, id := range follower.Following {
		if id == followeeID {
			return fmt.Errorf("already following")
		}
	}

	follower.Following = append(follower.Following, followeeID)
	followee.Followers = append(followee.Followers, followerID)

	return nil
}

// Unfollow makes one user unfollow another
func (s *NewsfeedService) Unfollow(followerID, followeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	follower, exists := s.users[followerID]
	if !exists {
		return fmt.Errorf("follower not found")
	}

	followee, exists := s.users[followeeID]
	if !exists {
		return fmt.Errorf("followee not found")
	}

	// Remove from following list
	newFollowing := []string{}
	found := false
	for _, id := range follower.Following {
		if id != followeeID {
			newFollowing = append(newFollowing, id)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("not following")
	}

	follower.Following = newFollowing

	// Remove from followers list
	newFollowers := []string{}
	for _, id := range followee.Followers {
		if id != followerID {
			newFollowers = append(newFollowers, id)
		}
	}
	followee.Followers = newFollowers

	return nil
}

// CreatePost creates a new post
func (s *NewsfeedService) CreatePost(userID, content string) (*Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[userID]; !exists {
		return nil, fmt.Errorf("user not found")
	}

	s.postIndex++
	postID := fmt.Sprintf("post_%d", s.postIndex)

	post := &Post{
		ID:        postID,
		UserID:    userID,
		Content:   content,
		Timestamp: time.Now(),
		Likes:     0,
		Comments:  0,
		Shares:    0,
	}

	s.posts[postID] = post
	s.userPosts[userID] = append(s.userPosts[userID], postID)

	return post, nil
}

// GetPost retrieves a post by ID
func (s *NewsfeedService) GetPost(postID string) (*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	return post, nil
}

// LikePost increments the like count for a post
func (s *NewsfeedService) LikePost(postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.Likes++
	return nil
}

// CommentPost increments the comment count for a post
func (s *NewsfeedService) CommentPost(postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.Comments++
	return nil
}

// SharePost increments the share count for a post
func (s *NewsfeedService) SharePost(postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.Shares++
	return nil
}

// GetUserPosts retrieves all posts by a user
func (s *NewsfeedService) GetUserPosts(userID string) ([]*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	postIDs, exists := s.userPosts[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	posts := make([]*Post, 0, len(postIDs))
	for _, postID := range postIDs {
		if post, exists := s.posts[postID]; exists {
			posts = append(posts, post)
		}
	}

	// Sort by timestamp descending
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	return posts, nil
}

// GetNewsfeed retrieves the newsfeed for a user (posts from followed users)
func (s *NewsfeedService) GetNewsfeed(userID string, limit int) ([]*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Collect posts from followed users
	posts := []*Post{}
	for _, followedID := range user.Following {
		if postIDs, exists := s.userPosts[followedID]; exists {
			for _, postID := range postIDs {
				if post, exists := s.posts[postID]; exists {
					posts = append(posts, post)
				}
			}
		}
	}

	// Sort by timestamp descending
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	return posts, nil
}

// DeletePost deletes a post
func (s *NewsfeedService) DeletePost(postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	// Remove from posts map
	delete(s.posts, postID)

	// Remove from user posts
	userID := post.UserID
	if postIDs, exists := s.userPosts[userID]; exists {
		newPostIDs := []string{}
		for _, id := range postIDs {
			if id != postID {
				newPostIDs = append(newPostIDs, id)
			}
		}
		s.userPosts[userID] = newPostIDs
	}

	return nil
}

// HTTP Handlers

var service *NewsfeedService

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := service.CreateUser(req.UserID, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	user, err := service.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func followHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FollowerID string `json:"follower_id"`
		FolloweeID string `json:"followee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.Follow(req.FollowerID, req.FolloweeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unfollowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FollowerID string `json:"follower_id"`
		FolloweeID string `json:"followee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.Unfollow(req.FollowerID, req.FolloweeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID  string `json:"user_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := service.CreatePost(req.UserID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func likePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PostID string `json:"post_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.LikePost(req.PostID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getNewsfeedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	limit := 50 // default limit
	posts, err := service.GetNewsfeed(userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func getUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	posts, err := service.GetUserPosts(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewNewsfeedService()

	http.HandleFunc("/user/create", createUserHandler)
	http.HandleFunc("/user/get", getUserHandler)
	http.HandleFunc("/user/follow", followHandler)
	http.HandleFunc("/user/unfollow", unfollowHandler)
	http.HandleFunc("/post/create", createPostHandler)
	http.HandleFunc("/post/like", likePostHandler)
	http.HandleFunc("/newsfeed", getNewsfeedHandler)
	http.HandleFunc("/posts", getUserPostsHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8081"
	log.Printf("Newsfeed service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

