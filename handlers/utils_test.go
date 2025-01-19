package handlers

import (
	"forum/database"
	"testing"
)

func TestGetUsername(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Insert test users
	_, err := database.DB.Exec(`
		INSERT INTO users (id, email, username, password) VALUES
		(2, 'test2@example.com', 'testuser2', 'password'),
		(3, 'test3@example.com', 'testuser3', 'password')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	tests := []struct {
		name           string
		userID         int
		expectedResult string
	}{
		{
			name:           "Existing User",
			userID:         1,
			expectedResult: "testuser",
		},
		{
			name:           "Another Existing User",
			userID:         2,
			expectedResult: "testuser2",
		},
		{
			name:           "Non-existent User",
			userID:         999,
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUsername(tt.userID)
			if result != tt.expectedResult {
				t.Errorf("getUsername() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}
