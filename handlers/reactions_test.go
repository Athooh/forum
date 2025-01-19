package handlers

import (
	"context"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterLikesHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test data
	_, err := database.DB.Exec(`
		INSERT INTO posts (id, user_id, title, content, category, image_url, created_at, updated_at) 
		VALUES (1, 1, 'Test Post', 'Content', 'Category1', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test reaction
	_, err = database.DB.Exec(`
		CREATE TABLE IF NOT EXISTS reactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER NOT NULL,
			reaction_type TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (post_id) REFERENCES posts (id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create reactions table: %v", err)
	}

	_, err = database.DB.Exec(`
		INSERT INTO reactions (user_id, post_id, reaction_type) 
		VALUES (1, 1, 'like')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test reaction: %v", err)
	}

	tests := []struct {
		name           string
		userID         int
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Authenticated User with Likes",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Authenticated User without Likes",
			userID:         2,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/filter-likes", nil)

			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			FilterLikesHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("FilterLikesHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestToggleReaction(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test post
	_, err := database.DB.Exec(`
		INSERT INTO posts (id, user_id, title, content, category, image_url, created_at, updated_at) 
		VALUES (1, 1, 'Test Post', 'Content', 'Category1', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name          string
		userID        int
		postID        int
		reactionType  string
		expectedError bool
	}{
		{
			name:          "Add New Like",
			userID:        1,
			postID:        1,
			reactionType:  "like",
			expectedError: false,
		},
		{
			name:          "Toggle Existing Like",
			userID:        1,
			postID:        1,
			reactionType:  "like",
			expectedError: false,
		},
		{
			name:          "Change to Dislike",
			userID:        1,
			postID:        1,
			reactionType:  "dislike",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ToggleReaction(tt.userID, tt.postID, tt.reactionType)
			if (err != nil) != tt.expectedError {
				t.Errorf("ToggleReaction() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
