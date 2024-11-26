package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/database"
	"net/http"
)

type Reaction struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	PostID    *int   `json:"post_id"`
	CommentID *int   `json:"comment_id"`
	Reaction  string `json:"reaction"` // "like" or "dislike"
}

// AddReactionHandler handles adding a reaction
func AddReactionHandler(w http.ResponseWriter, r *http.Request) {
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
	var reaction Reaction
	err = json.NewDecoder(r.Body).Decode(&reaction)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate input
	if reaction.PostID == nil && reaction.CommentID == nil {
		http.Error(w, "Either post_id or comment_id must be provided", http.StatusBadRequest)
		return
	}
	if reaction.Reaction != "like" && reaction.Reaction != "dislike" {
		http.Error(w, "Reaction must be 'like' or 'dislike'", http.StatusBadRequest)
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

	// Insert or update the reaction in the database
	_, err = database.DB.Exec(`
		INSERT INTO likes_dislikes (user_id, post_id, comment_id, reaction)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, post_id, comment_id) DO UPDATE SET reaction = ?`,
		userID, reaction.PostID, reaction.CommentID, reaction.Reaction, reaction.Reaction)
	if err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Reaction added successfully"))
}

// GetReactionsHandler fetches like/dislike counts for a post or comment
func GetReactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Determine whether to fetch reactions for a post or comment
	postID := r.URL.Query().Get("post_id")
	commentID := r.URL.Query().Get("comment_id")

	if postID == "" && commentID == "" {
		http.Error(w, "post_id or comment_id must be provided", http.StatusBadRequest)
		return
	}

	// Query the database for counts
	var likes, dislikes int
	var err error
	if postID != "" {
		err = database.DB.QueryRow("SELECT COUNT(*) FROM likes_dislikes WHERE post_id = ? AND reaction = 'like'", postID).Scan(&likes)
		err = database.DB.QueryRow("SELECT COUNT(*) FROM likes_dislikes WHERE post_id = ? AND reaction = 'dislike'", postID).Scan(&dislikes)
	} else {
		err = database.DB.QueryRow("SELECT COUNT(*) FROM likes_dislikes WHERE comment_id = ? AND reaction = 'like'", commentID).Scan(&likes)
		err = database.DB.QueryRow("SELECT COUNT(*) FROM likes_dislikes WHERE comment_id = ? AND reaction = 'dislike'", commentID).Scan(&dislikes)
	}

	if err != nil {
		http.Error(w, "Failed to fetch reactions", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}
