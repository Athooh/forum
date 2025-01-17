package handlers

import (
	"database/sql"
	"forum/database"
	"forum/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Mock templates for testing
var mockLoginTmpl = `
<!DOCTYPE html>
<html>
<body>
 <form method="POST" action="/login">
 <input type="text" name="username">
 <input type="password" name="password">
 <button type="submit">Login</button>
 </form>
 {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
</body>
</html>
`

var mockSignupTmpl = `
<!DOCTYPE html>
<html>
<body>
 <form method="POST" action="/signup">
 <input type="email" name="email">
 <input type="text" name="username">
 <input type="password" name="password">
 <button type="submit">Sign Up</button>
 </form>
 {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
</body>
</html>
`

func setupTestTemplates(t *testing.T) {
    // Create a temporary directory for templates
    tmpDir, err := os.MkdirTemp("", "forum-templates-*")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    t.Cleanup(func() {
        os.RemoveAll(tmpDir)
    })

    // Create directories
    templatesDir := filepath.Join(tmpDir, "templates")
    errorDir := filepath.Join(templatesDir, "error")
    err = os.MkdirAll(errorDir, 0755)
    if err != nil {
        t.Fatalf("Failed to create directories: %v", err)
    }

    // Write template files
    files := map[string]string{
        filepath.Join(templatesDir, "login.html"):    mockLoginTmpl,
        filepath.Join(templatesDir, "signup.html"):   mockSignupTmpl,
        filepath.Join(errorDir, "404.html"):         "<html><body>404 Not Found</body></html>",
    }

    for path, content := range files {
        err := os.WriteFile(path, []byte(content), 0666)
        if err != nil {
            t.Fatalf("Failed to write template %s: %v", path, err)
        }
    }

    // Initialize templates using the temporary directory
    err = InitTemplates(tmpDir)
    if err != nil {
        t.Fatalf("Failed to initialize templates: %v", err)
    }
}

func setupTestDB(t *testing.T) {
	// Initialize test database connection
	var err error
	database.DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create necessary tables
	_, err = database.DB.Exec(`
 CREATE TABLE users (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 email TEXT UNIQUE,
 username TEXT UNIQUE,
 password TEXT
 )
 `)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = database.DB.Exec(`
 CREATE TABLE sessions (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 user_id INTEGER,
 session_token TEXT UNIQUE,
 expires_at DATETIME,
 FOREIGN KEY (user_id) REFERENCES users(id)
 )
 `)
	if err != nil {
		t.Fatalf("Failed to create sessions table: %v", err)
	}
}

func clearTestDB(t *testing.T) {
	_, err := database.DB.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatalf("Failed to clear sessions table: %v", err)
	}
	_, err = database.DB.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clear users table: %v", err)
	}
}

func TestLoginHandler(t *testing.T) {
	setupTestDB(t)
	setupTestTemplates(t)
	defer clearTestDB(t)

	// Create a test user
	hashedPassword, _ := utils.HashPassword("testpassword123")
	_, err := database.DB.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		"test@example.com",
		"testuser",
		hashedPassword,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		username       string
		password       string
		expectedStatus int
		expectedPath   string
	}{
		{
			name:           "Valid Login",
			method:         http.MethodPost,
			username:       "testuser",
			password:       "testpassword123",
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/dashboard",
		},
		{
			name:           "Invalid Password",
			method:         http.MethodPost,
			username:       "testuser",
			password:       "wrongpassword",
			expectedStatus: http.StatusOK, // Returns to login page with error
			expectedPath:   "",
		},
		{
			name:           "Empty Username",
			method:         http.MethodPost,
			username:       "",
			password:       "testpassword123",
			expectedStatus: http.StatusBadRequest,
			expectedPath:   "",
		},
		{
			name:           "GET Request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				form := url.Values{}
				form.Add("username", tt.username)
				form.Add("password", tt.password)
				req = httptest.NewRequest(tt.method, "/login", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.method, "/login", nil)
			}

			rr := httptest.NewRecorder()
			LoginHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedPath != "" {
				location := rr.Header().Get("Location")
				if location != tt.expectedPath {
					t.Errorf("handler returned wrong redirect path: got %v want %v", location, tt.expectedPath)
				}
			}

			if tt.expectedStatus == http.StatusSeeOther {
				cookies := rr.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "session_token" {
						sessionCookie = cookie
						break
					}
				}
				if sessionCookie == nil {
					t.Error("no session cookie set after successful login")
				}
			}
		})
	}
}

func TestSignUpHandler(t *testing.T) {
	setupTestDB(t)
	setupTestTemplates(t)
	defer clearTestDB(t)

	tests := []struct {
		name           string
		method         string
		email          string
		username       string
		password       string
		expectedStatus int
		expectedPath   string
	}{
		{
			name:           "Valid Registration",
			method:         http.MethodPost,
			email:          "newuser@example.com",
			username:       "newuser",
			password:       "StrongPass123!",
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login",
		},
		{
			name:           "Invalid Email",
			method:         http.MethodPost,
			email:          "invalid-email",
			username:       "newuser",
			password:       "StrongPass123!",
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name:           "Weak Password",
			method:         http.MethodPost,
			email:          "newuser@example.com",
			username:       "newuser",
			password:       "weak",
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name:           "GET Request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				form := url.Values{}
				form.Add("email", tt.email)
				form.Add("username", tt.username)
				form.Add("password", tt.password)
				req = httptest.NewRequest(tt.method, "/signup", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.method, "/signup", nil)
			}

			rr := httptest.NewRecorder()
			SignUpHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedPath != "" {
				location := rr.Header().Get("Location")
				if location != tt.expectedPath {
					t.Errorf("handler returned wrong redirect path: got %v want %v", location, tt.expectedPath)
				}
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	setupTestDB(t)
	defer clearTestDB(t)

	// Create a test user and session
	hashedPassword, _ := utils.HashPassword("testpassword123")
	result, err := database.DB.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		"test@example.com",
		"testuser",
		hashedPassword,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, _ := result.LastInsertId()
	sessionToken := "test-session-token"
	_, err = database.DB.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
		userID,
		sessionToken,
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
		expectedPath   string
	}{
		{
			name:           "Valid Logout",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/",
		},
		{
			name:           "No Session",
			sessionToken:   "",
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})
			}

			rr := httptest.NewRecorder()
			LogoutHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if location := rr.Header().Get("Location"); location != tt.expectedPath {
				t.Errorf("handler returned wrong redirect path: got %v want %v", location, tt.expectedPath)
			}

			cookies := rr.Result().Cookies()
			for _, cookie := range cookies {
				if cookie.Name == "session_token" {
					if !cookie.Expires.Before(time.Now()) {
						t.Error("session cookie not properly expired")
					}
					break
				}
			}
		})
	}
}
