package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"forum/database"
)

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// AddCommentHandler handles adding a comment to a post
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the user is authenticated
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var comment Comment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate input
	if comment.PostID == 0 || comment.Content == "" {
		http.Error(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	// Find the user ID from the session token (mocked; replace with real session handling)
	var userID int
	err = database.DB.QueryRow("SELECT id FROM users WHERE id = ?", 1).Scan(&userID) // Replace with session logic
	if err == sql.ErrNoRows {
		http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert comment into the database
	_, err = database.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", comment.PostID, userID, comment.Content)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Comment added successfully"))
}

// GetCommentsHandler fetches all comments for a specific post
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCommentsHandler called") // Add this log
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get post_id from query params
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Query all comments for the post
	rows, err := database.DB.Query("SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.created_at, users.username FROM comments INNER JOIN users ON comments.user_id = users.id WHERE comments.post_id = ? ORDER BY comments.created_at DESC", postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare response
	var comments []map[string]interface{}
	for rows.Next() {
		var id, postID, userID int
		var content, username string
		var createdAt time.Time
		err = rows.Scan(&id, &postID, &userID, &content, &createdAt, &username)
		if err != nil {
			http.Error(w, "Failed to parse comments", http.StatusInternalServerError)
			return
		}

		comments = append(comments, map[string]interface{}{
			"id":         id,
			"post_id":    postID,
			"user_id":    userID,
			"username":   username,
			"content":    content,
			"created_at": createdAt,
		})
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
