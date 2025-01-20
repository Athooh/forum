package handlers

import (
	"context"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAddCommentHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test post
	_, err := database.DB.Exec(`
		INSERT INTO posts (id, user_id, title, content, category) 
		VALUES (1, 1, 'Test Post', 'Content', 'Category1')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		postID         string
		comment        string
		userID         int
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Valid Comment",
			method:         http.MethodPost,
			postID:         "1",
			comment:        "Test Comment",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			postID:         "1",
			authenticated:  true,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Empty Comment",
			method:         http.MethodPost,
			postID:         "1",
			comment:        "",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Post ID",
			method:         http.MethodPost,
			postID:         "invalid",
			comment:        "Test Comment",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("comment", tt.comment)

			req := httptest.NewRequest(tt.method, "/add-comment?post_id="+tt.postID, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			AddCommentHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("AddCommentHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestEditCommentHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test comment
	_, err := database.DB.Exec(`
		INSERT INTO comments (id, post_id, user_id, username, content) 
		VALUES (1, 1, 1, 'testuser', 'Original Comment')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		commentID      string
		content        string
		userID         int
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Valid Edit",
			method:         http.MethodPost,
			commentID:      "1",
			content:        "Updated Comment",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			commentID:      "1",
			authenticated:  true,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Empty Content",
			method:         http.MethodPost,
			commentID:      "1",
			content:        "",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized Edit",
			method:         http.MethodPost,
			commentID:      "1",
			content:        "Updated Comment",
			userID:         2,
			authenticated:  true,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("content", tt.content)

			req := httptest.NewRequest(tt.method, "/edit-comment?id="+tt.commentID, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			EditCommentHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("EditCommentHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

