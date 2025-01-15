package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"forum/database"
	"forum/utils"
)

// ToggleReaction handles the logic for toggling a reaction on a post
func ToggleReaction(userID, postID int, reactionType string) error {
	// First, check if user has any existing reaction
	var existingReaction string
	err := database.DB.QueryRow(`
		SELECT reaction_type 
		FROM reactions 
		WHERE user_id = ? AND post_id = ?
	`, userID, postID).Scan(&existingReaction)

	// If there's no existing reaction, add new one
	if err == sql.ErrNoRows {
		_, err = database.DB.Exec(`
			INSERT INTO reactions (user_id, post_id, reaction_type) 
			VALUES (?, ?, ?)
		`, userID, postID, reactionType)
		return err
	}

	// If there is an error other than no rows, return it
	if err != nil {
		return err
	}

	// If clicking the same reaction type, remove it
	if existingReaction == reactionType {
		_, err = database.DB.Exec(`
			DELETE FROM reactions 
			WHERE user_id = ? AND post_id = ?
		`, userID, postID)
		return err
	}

	// If changing reaction type, update it
	_, err = database.DB.Exec(`
		UPDATE reactions 
		SET reaction_type = ? 
		WHERE user_id = ? AND post_id = ?
	`, reactionType, userID, postID)
	return err
}

func FilterLikesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	posts, err := database.GetLikesByUserID(userID)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving posts: %v", err))
		return
	}

	// Get category counts
	categoryCounts, err := database.GetCategoryPostCounts()
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving category counts: %v", err))
		return
	}

	// Check if user is authenticated
	isAuthenticated := userID != 0

	// Get username if authenticated
	var username string
	if isAuthenticated {
		username = getUsername(userID)
	}

	// Add username to posts
	for i := range posts {
		posts[i].Username = getUsername(posts[i].UserID)
		if len(posts[i].Content) > 200 {
			posts[i].Preview = posts[i].Content[:200] + "..."
		} else {
			posts[i].Preview = posts[i].Content
		}
		posts[i].CreatedAtHuman = utils.TimeAgo(posts[i].CreatedAt)
	}

	data := map[string]interface{}{
		"Title":           "Forum - Home",
		"IsLoggedIn":      isAuthenticated,
		"Username":        username,
		"Posts":           posts,
		"Categories":      categoryCounts,
		"IsAuthenticated": isAuthenticated,
	}

	err = templates.ExecuteTemplate(w, "dashboard", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		return
	}
}
