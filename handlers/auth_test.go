package handlers

import (
	"database/sql"
	"forum/database"
	"forum/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginHandler(t *testing.T) {
	setupTestDB(t)
	defer database.DB.Close()

	// Create a properly hashed password for "correctpassword"
	hashedPassword, err := utils.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Insert test user with proper hashed password
	_, err = database.DB.Exec(`
		INSERT INTO users (email, username, password) 
		VALUES ('login@test.com', 'loginuser', ?)
	`, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "GET Login Page",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST Valid Login",
			method:         "POST",
			username:       "loginuser",
			password:       "correctpassword",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "POST Invalid Username",
			method:         "POST",
			username:       "nonexistent",
			password:       "password",
			expectedStatus: http.StatusOK, // Returns to login page with error
		},
		{
			name:           "POST Empty Credentials",
			method:         "POST",
			username:       "",
			password:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == "GET" {
				req = httptest.NewRequest("GET", "/login", nil)
			} else {
				form := url.Values{}
				form.Add("username", tt.username)
				form.Add("password", tt.password)
				req = httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			w := httptest.NewRecorder()
			LoginHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("LoginHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestSignUpHandler(t *testing.T) {
	// Setup test database
	database.DB, _ = sql.Open("sqlite3", ":memory:")
	defer database.DB.Close()

	// Create users table
	database.DB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT UNIQUE,
			username TEXT UNIQUE,
			password TEXT
		)
	`)

	tests := []struct {
		name           string
		method         string
		email          string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "GET Signup Page",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST Valid Signup",
			method:         "POST",
			email:          "test@example.com",
			username:       "newuser",
			password:       "StrongPass123!",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "POST Invalid Email",
			method:         "POST",
			email:          "invalid-email",
			username:       "user",
			password:       "password",
			expectedStatus: http.StatusOK, // Returns to signup page with error
		},
		{
			name:           "POST Weak Password",
			method:         "POST",
			email:          "test@example.com",
			username:       "user",
			password:       "weak",
			expectedStatus: http.StatusOK, // Returns to signup page with error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == "GET" {
				req = httptest.NewRequest("GET", "/signup", nil)
			} else {
				form := url.Values{}
				form.Add("email", tt.email)
				form.Add("username", tt.username)
				form.Add("password", tt.password)
				req = httptest.NewRequest("POST", "/signup", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			w := httptest.NewRecorder()
			SignUpHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("SignUpHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	// Setup test database
	database.DB, _ = sql.Open("sqlite3", ":memory:")
	defer database.DB.Close()

	// Create sessions table
	database.DB.Exec(`
		CREATE TABLE sessions (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			session_token TEXT,
			expires_at DATETIME
		)
	`)

	// Insert test session
	testToken := "test-session-token"
	database.DB.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
		1, testToken, time.Now().Add(24*time.Hour),
	)

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
	}{
		{
			name:           "Valid Logout",
			sessionToken:   testToken,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "No Session Cookie",
			sessionToken:   "",
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/logout", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})
			}

			w := httptest.NewRecorder()
			LogoutHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("LogoutHandler() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check if cookie was cleared
			if tt.sessionToken != "" {
				cookies := w.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == "session_token" {
						if cookie.Value != "" || !cookie.Expires.Before(time.Now()) {
							t.Error("Session cookie was not properly cleared")
						}
					}
				}
			}
		})
	}
}
