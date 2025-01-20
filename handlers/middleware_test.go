package handlers

import (
	"context"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Create a mock handler to use with the middleware
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if userID exists in context
		userID := r.Context().Value(userIDKey)
		if userID == nil {
			t.Error("userID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
	}{
		{
			name:           "No Session Cookie",
			sessionToken:   "",
			expectedStatus: http.StatusSeeOther, // Should redirect to login
		},
		{
			name:           "Invalid Session Token",
			sessionToken:   "invalid_token",
			expectedStatus: http.StatusSeeOther, // Should redirect to login
		},
		{
			name:           "Valid Session",
			sessionToken:   "valid_token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add session cookie if test case requires it
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})

				// For valid session token, insert a test session
				if tt.sessionToken == "valid_token" {
					_, err := database.DB.Exec(`
						INSERT INTO sessions (user_id, session_token, expires_at) 
						VALUES (1, 'valid_token', datetime('now', '+1 day'))
					`)
					if err != nil {
						t.Fatalf("Failed to insert test session: %v", err)
					}
				}
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
			handler := AuthMiddleware(mockHandler)
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("AuthMiddleware() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestGuestMiddleware(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Create a mock handler to use with the middleware
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if userID exists in context
		userID := r.Context().Value(userIDKey)
		if userID != nil {
			// For authenticated users, verify the ID
			if id, ok := userID.(int); !ok || id != 1 {
				t.Errorf("Expected userID = 1, got %v", id)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "Guest User",
			sessionToken:   "",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Invalid Session Token",
			sessionToken:   "invalid_token",
			expectedStatus: http.StatusOK,
			expectUserID:   false,
		},
		{
			name:           "Authenticated User",
			sessionToken:   "valid_token",
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add session cookie if test case requires it
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})

				// For valid session token, insert a test session
				if tt.sessionToken == "valid_token" {
					_, err := database.DB.Exec(`
						INSERT INTO sessions (user_id, session_token, expires_at) 
						VALUES (1, 'valid_token', datetime('now', '+1 day'))
					`)
					if err != nil {
						t.Fatalf("Failed to insert test session: %v", err)
					}
				}
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
			handler := GuestMiddleware(mockHandler)
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GuestMiddleware() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestOwnershipMiddleware(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Create a mock handler to use with the middleware
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Insert test data
	_, err := database.DB.Exec(`
		INSERT INTO posts (id, user_id, title, content, category) 
		VALUES (100, 1, 'Test Post', 'Content', 'Category1')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test comment with username
	_, err = database.DB.Exec(`
		INSERT INTO comments (id, post_id, user_id, username, content) 
		VALUES (100, 1, 1, 'testuser', 'Test Comment')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		resourceID     string
		userID         int
		expectedStatus int
	}{
		{
			name:           "Owner Access Post",
			path:           "/edit-post",
			resourceID:     "100",
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-Owner Access Post",
			path:           "/edit-post",
			resourceID:     "100",
			userID:         2,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Owner Access Comment",
			path:           "/edit-comment",
			resourceID:     "100",
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-Owner Access Comment",
			path:           "/edit-comment",
			resourceID:     "100",
			userID:         2,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid Resource ID",
			path:           "/edit-post",
			resourceID:     "invalid",
			userID:         1,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Non-Existent Resource",
			path:           "/edit-post",
			resourceID:     "999",
			userID:         1,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path+"?id="+tt.resourceID, nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
			req = req.WithContext(ctx)

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
			handler := OwnershipMiddleware(mockHandler)
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("OwnershipMiddleware() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
