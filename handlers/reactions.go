package handlers

import (
	"database/sql"
	"forum/database"
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
