package handlers

import (
	"bytes"
	"context"
	"forum/database"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestCreatePostHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	tests := []struct {
		name           string
		method         string
		userID         int
		title          string
		content        string
		category       string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Valid GET Request",
			method:         http.MethodGet,
			authenticated:  true,
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid POST Request",
			method:         http.MethodPost,
			userID:         1,
			title:          "Test Post",
			content:        "Test Content",
			category:       "Category1",
			authenticated:  true,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Unauthenticated Request",
			method:         http.MethodPost,
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty Fields",
			method:         http.MethodPost,
			userID:         1,
			title:          "",
			content:        "",
			category:       "",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.method == http.MethodPost {
				writer.WriteField("title", tt.title)
				writer.WriteField("content", tt.content)
				writer.WriteField("category", tt.category)
				writer.Close()
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/create-post", body)

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", writer.FormDataContentType())
			}

			// Add authentication if needed
			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			CreatePostHandler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("CreatePostHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestEditPostHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test categories
	_, err := database.DB.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		);
		INSERT INTO categories (name) VALUES ('Category1');
	`)
	if err != nil {
		t.Fatalf("Failed to create categories: %v", err)
	}

	// Insert a test post with a unique ID
	result, err := database.DB.Exec(`
		INSERT INTO posts (user_id, title, content, category, image_url, created_at, updated_at) 
		VALUES (1, 'Original Title', 'Original Content', 'Category1', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get post ID: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		postID         string
		userID         int
		title          string
		content        string
		category       string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Valid GET Request",
			method:         http.MethodGet,
			postID:         strconv.FormatInt(postID, 10),
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid POST Request",
			method:         http.MethodPost,
			postID:         strconv.FormatInt(postID, 10),
			userID:         1,
			title:          "Updated Title",
			content:        "Updated Content",
			category:       "Category1",
			authenticated:  true,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Invalid Post ID",
			method:         http.MethodGet,
			postID:         "invalid",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Post",
			method:         http.MethodGet,
			postID:         "999",
			userID:         1,
			authenticated:  true,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				form := bytes.NewBufferString("")
				writer := multipart.NewWriter(form)
				writer.WriteField("title", tt.title)
				writer.WriteField("content", tt.content)
				writer.WriteField("category", tt.category)
				writer.Close()

				req = httptest.NewRequest(tt.method, "/edit-post?id="+tt.postID, form)
				req.Header.Set("Content-Type", writer.FormDataContentType())
			} else {
				req = httptest.NewRequest(tt.method, "/edit-post?id="+tt.postID, nil)
			}

			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			EditPostHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("EditPostHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
