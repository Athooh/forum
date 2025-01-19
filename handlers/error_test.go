package handlers

import (
	"fmt"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderErrorPage(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	tests := []struct {
		name          string
		statusCode    int
		err           error
		expectedCode  int
		expectedTitle string
	}{
		{
			name:          "400 Bad Request",
			statusCode:    http.StatusBadRequest,
			err:           fmt.Errorf("bad request error"),
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "400 Error",
		},
		{
			name:          "401 Unauthorized",
			statusCode:    http.StatusUnauthorized,
			err:           fmt.Errorf("unauthorized error"),
			expectedCode:  http.StatusUnauthorized,
			expectedTitle: "401 Error",
		},
		{
			name:          "403 Forbidden",
			statusCode:    http.StatusForbidden,
			err:           fmt.Errorf("forbidden error"),
			expectedCode:  http.StatusForbidden,
			expectedTitle: "403 Error",
		},
		{
			name:          "404 Not Found",
			statusCode:    http.StatusNotFound,
			err:           fmt.Errorf("not found error"),
			expectedCode:  http.StatusNotFound,
			expectedTitle: "404 Error",
		},
		{
			name:          "405 Method Not Allowed",
			statusCode:    http.StatusMethodNotAllowed,
			err:           fmt.Errorf("method not allowed error"),
			expectedCode:  http.StatusMethodNotAllowed,
			expectedTitle: "405 Error",
		},
		{
			name:          "500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			err:           fmt.Errorf("internal server error"),
			expectedCode:  http.StatusInternalServerError,
			expectedTitle: "500 Error",
		},
		{
			name:          "Unknown Error Code",
			statusCode:    999,
			err:           fmt.Errorf("unknown error"),
			expectedCode:  999,
			expectedTitle: "500 Error", // Falls back to 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RenderErrorPage(w, tt.statusCode, tt.err)

			if w.Code != tt.expectedCode {
				t.Errorf("RenderErrorPage() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			// Check if response contains expected title
			if body := w.Body.String(); !contains(body, tt.expectedTitle) {
				t.Errorf("RenderErrorPage() body = %v, want to contain %v", body, tt.expectedTitle)
			}
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	tests := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{
			name:         "Handle 404",
			path:         "/nonexistent",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			NotFoundHandler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("NotFoundHandler() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			// Check if response contains 404 error message
			if body := w.Body.String(); !contains(body, "404 Error") {
				t.Errorf("NotFoundHandler() body = %v, want to contain '404 Error'", body)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
