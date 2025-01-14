package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"forum/database"
)

// AddCommentHandler handles adding a comment to a post.
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed, fmt.Errorf("invalid request method"))
		return
	}

	userID := r.Context().Value(userIDKey).(int)
	postIDStr := r.URL.Query().Get("post_id")
	content := r.FormValue("comment")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Post ID: %v", err))
		return
	}

	if content == "" {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("comment content cannot be empty"))
		return
	}

	err = database.AddComment(postID, userID, content)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to add comment: %v", err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/view-post?id=%d", postID), http.StatusSeeOther)
}

// EditCommentHandler handles comment editing
func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	userID := r.Context().Value(userIDKey).(int)
	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid comment ID: %v", err))
		return
	}

	// Verify the comment belongs to the user
	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound, fmt.Errorf("comment not found: %v", err))
		return
	}

	if comment.UserID != userID {
		RenderErrorPage(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	content := r.FormValue("content")
	if content == "" {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("comment content cannot be empty"))
		return
	}

	err = database.UpdateComment(commentID, content)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to update comment: %v", err))
		return
	}

	// Return success response for AJAX requests
	w.WriteHeader(http.StatusOK)
}

// DeleteCommentHandler handles comment deletion
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	userID := r.Context().Value(userIDKey).(int)
	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid comment ID: %v", err))
		return
	}

	// Verify the comment belongs to the user
	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound, fmt.Errorf("comment not found: %v", err))
		return
	}

	if comment.UserID != userID {
		RenderErrorPage(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	err = database.DeleteComment(commentID)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to delete comment: %v", err))
		return
	}

	// Return success response for AJAX requests
	w.WriteHeader(http.StatusOK)
}

// LikeCommentHandler handles liking a comment by an authenticated user
func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	commentIDStr := r.URL.Query().Get("id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Comment ID: %v", err))
		return
	}

	err = database.ToggleCommentReaction(userID, commentID, "like")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to like comment: %v", err))
		return
	}

	likes, dislikes, _ := database.GetCommentReactionCounts(commentID)

	// Set response headers and send the JSON response with updated counts
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"likes": %d, "dislikes": %d}`, likes, dislikes)))
}

// DislikeCommentHandler handles disliking a comment.
func DislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	commentIDStr := r.URL.Query().Get("id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Comment ID: %v", err))
		return
	}

	err = database.ToggleCommentReaction(userID, commentID, "dislike")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to dislike comment: %v", err))
		return
	}

	// Return updated reaction counts
	likes, dislikes, _ := database.GetCommentReactionCounts(commentID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"likes": %d, "dislikes": %d}`, likes, dislikes)))
}
