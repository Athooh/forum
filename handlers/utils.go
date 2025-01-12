package handlers

import "forum/database"

// getUsername retrieves the username for a given user ID
func getUsername(userID int) string {
	var username string
	query := `SELECT username FROM users WHERE id = ?`
	err := database.DB.QueryRow(query, userID).Scan(&username)
	if err != nil {
		return ""
	}
	return username
}
