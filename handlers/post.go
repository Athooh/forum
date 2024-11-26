package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"forum/database"
)

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePostHandler handles post creation
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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
	var post Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate input
	if post.Title == "" || post.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
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

	// Insert post into the database
	_, err = database.DB.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, post.Title, post.Content)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Post created successfully"))
}

// / GetPostsHandler fetches posts, optionally filtered by category
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get optional category_id from query params
	categoryIDStr := r.URL.Query().Get("category_id")
	log.Printf("Received category_id: %s", categoryIDStr)

	// Convert category_id to integer
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID == 0 {
		http.Error(w, "Invalid or missing category_id", http.StatusBadRequest)
		return
	}

	// Prepare query
	var rows *sql.Rows
	var query string
	if categoryID != 0 {
		query = `
			SELECT posts.id, posts.user_id, posts.title, posts.content, posts.created_at, users.username 
			FROM posts
			INNER JOIN post_categories ON posts.id = post_categories.post_id
			INNER JOIN users ON posts.user_id = users.id
			WHERE post_categories.category_id = ?
			ORDER BY posts.created_at DESC
		`
		log.Println("Executing query with category_id...")
		rows, err = database.DB.Query(query, categoryID)
	} else {
		query = `
			SELECT posts.id, posts.user_id, posts.title, posts.content, posts.created_at, users.username 
			FROM posts
			INNER JOIN users ON posts.user_id = users.id
			ORDER BY posts.created_at DESC
		`
		log.Println("Executing query for all posts...")
		rows, err = database.DB.Query(query)
	}

	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare response
	var posts []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		rowCount++
		log.Println("Processing a row from query results...")
		var id, userID int
		var title, content, username string
		var createdAt time.Time
		err = rows.Scan(&id, &userID, &title, &content, &createdAt, &username)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Failed to parse posts", http.StatusInternalServerError)
			return
		}

		// Append the post to the posts slice
		posts = append(posts, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"title":      title,
			"content":    content,
			"username":   username,
			"created_at": createdAt,
		})
	}
	log.Printf("Rows retrieved: %d", rowCount)

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts) // Send the posts slice as the response
}
