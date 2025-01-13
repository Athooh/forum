package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"forum/database"
	"forum/utils"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("LoginHandler invoked")

	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"Title": "Login - Forum",
		}
		templates.ExecuteTemplate(w, "login", data)
		return
	}
	// List all available templates
	// fmt.Println("Available templates:", templates.Templates())
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("missing username or password"))
			return
		}

		var id int
		var hashedPassword string

		query := `SELECT id, password FROM users WHERE username = ?`
		err := database.DB.QueryRow(query, username).Scan(&id, &hashedPassword)
		if err != nil {
			data := map[string]interface{}{
				"Title": "Login - Forum",
				"Error": "Invalid username or password",
			}
			templates.ExecuteTemplate(w, "login", data)
			return
		}

		err = utils.CheckPassword(hashedPassword, password)
		if err != nil {
			data := map[string]interface{}{
				"Title": "Login - Forum",
				"Error": "Invalid username or password",
			}
			templates.ExecuteTemplate(w, "login", data)
			return
		}

		// Check if the user already has an active session
		var sessionToken string
		err = database.DB.QueryRow(`
			SELECT session_token
			FROM sessions
			WHERE user_id = ? AND expires_at > ?`,
			id, time.Now(),
		).Scan(&sessionToken)

		if err == nil {
			// Active session found, delete the previous session
			_, err = database.DB.Exec(`
				DELETE FROM sessions
				WHERE user_id = ?`,
				id,
			)
			if err != nil {
				RenderErrorPage(w, http.StatusInternalServerError, err)
				return
			}
		}

		// Generate a new session token
		sessionToken, err = utils.GenerateSessionToken()
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError, err)
			return
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		query = `INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`
		_, err = database.DB.Exec(query, id, sessionToken, expiresAt)
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"Title": "Signup - Forum",
		}
		templates.ExecuteTemplate(w, "signup", data)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate email format
		if err := utils.ValidateEmail(email); err != nil {
			data := map[string]interface{}{
				"Title": "Signup - Forum",
				"Error": err.Error(),
			}
			templates.ExecuteTemplate(w, "signup", data)
			return
		}

		// Validate password strength
		if err := utils.ValidatePasswordStrength(password); err != nil {
			data := map[string]interface{}{
				"Title": "Signup - Forum",
				"Error": err.Error(),
			}
			templates.ExecuteTemplate(w, "signup", data)
			return
		}

		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			data := map[string]interface{}{
				"Title": "Signup - Forum",
				"Error": "Internal server error",
			}
			templates.ExecuteTemplate(w, "signup", data)
			return
		}

		query := `INSERT INTO users (email, username, password) VALUES (?, ?, ?)`
		_, err = database.DB.Exec(query, email, username, hashedPassword)
		if err != nil {
			var errorMessage string
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				if strings.Contains(err.Error(), "email") {
					errorMessage = "This email is already registered"
				} else {
					errorMessage = "This username is already taken"
				}
			} else {
				errorMessage = "An error occurred during registration"
			}

			data := map[string]interface{}{
				"Title": "Signup - Forum",
				"Error": errorMessage,
			}
			templates.ExecuteTemplate(w, "signup", data)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	query := `DELETE FROM sessions 	WHERE session_token = ?`
	_, err = database.DB.Exec(query, cookie.Value)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
