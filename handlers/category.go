package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"forum/database"
	"forum/utils"
)

// PostsByCategoryHandler handles category filtering
func PostsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("category parameter is required"))
		return
	}

	userID := r.Context().Value(userIDKey)
	currentUserID := 0
	if userID != nil {
		currentUserID = userID.(int)
	}

	posts, err := database.GetAllPosts(category)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving posts: %v", err))
		return
	}

	for i := range posts {
		// Get username
		var username string
		err := database.DB.QueryRow("SELECT username FROM users WHERE id = ?", posts[i].UserID).Scan(&username)
		if err != nil {
			posts[i].Username = "Unknown"
		} else {
			posts[i].Username = username
		}

		// Get comments count
		commentsCount, err := database.GetCommentsCount(posts[i].ID)
		if err != nil {
			posts[i].CommentsCount = 0
		} else {
			posts[i].CommentsCount = commentsCount
		}

		// Get reaction counts
		likes, dislikes, _ := database.GetReactionCounts(posts[i].ID)
		posts[i].Likes = likes
		posts[i].Dislikes = dislikes

		// Set ownership and format content
		posts[i].IsOwner = (posts[i].UserID == currentUserID)
		posts[i].CreatedAtHuman = utils.TimeAgo(posts[i].CreatedAt)
		if len(posts[i].Content) > 200 {
			posts[i].Preview = posts[i].Content[:200] + "..."
		} else {
			posts[i].Preview = posts[i].Content
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"posts":   posts,
	})
}

// GetCategoriesHandler handles fetching all categories
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Move categories fetching code here
}
