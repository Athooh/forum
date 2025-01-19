package handlers

import (
	"context"
	"forum/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostsByCategoryHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	tests := []struct {
		name           string
		category       string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Valid Category",
			category:       "Category1",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Category",
			category:       "",
			authenticated:  false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Category",
			category:       "NonExistentCategory",
			authenticated:  false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/posts-by-category?category="+tt.category, nil)

			if tt.authenticated {
				ctx := context.WithValue(req.Context(), userIDKey, 1)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			PostsByCategoryHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("PostsByCategoryHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestGetCategoriesHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Get Categories",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/categories", nil)
			w := httptest.NewRecorder()

			GetCategoriesHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetCategoriesHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
