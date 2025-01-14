package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"forum/database"
	"forum/utils"
)

// CreatePostHandler handles creating new posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		categories, err := database.GetCategories()
		if err != nil {
			log.Printf("Failed to load categories: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to load categories: %v", err))
			return
		}

		data := map[string]interface{}{
			"Title":      "Create Post - Forum",
			"IsLoggedIn": true,
			"Categories": categories,
		}

		err = templates.ExecuteTemplate(w, "create-post", data)
		if err != nil {
			log.Printf("Failed to render create-post template: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		}
		return
	}

	if r.Method == http.MethodPost {
		// Parse multipart form for image upload
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			log.Printf("Failed to parse multipart form: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to parse form data: %v", err))
			return
		}

		userID, ok := r.Context().Value(userIDKey).(int)
		if !ok {
			log.Println("User ID not found in context")
			RenderErrorPage(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		category := r.FormValue("category")

		if title == "" || content == "" || category == "" {
			log.Println("Missing required fields")
			RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("title, content, and category are required"))
			return
		}

		// Handle image upload
		var imageURL string
		file, header, err := r.FormFile("image")
		if err == nil && header != nil {
			defer file.Close()

			imagePath := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), header.Filename)
			dst, err := os.Create(imagePath)
			if err != nil {
				log.Printf("Failed to create image file: %v", err)
				RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to upload image: %v", err))
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				log.Printf("Failed to save image: %v", err)
				RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to save image: %v", err))
				return
			}

			// Compress and resize the image
			compressedImagePath := fmt.Sprintf("uploads/compressed_%d_%s", time.Now().Unix(), header.Filename)
			err = utils.CompressAndResizeImage(imagePath, compressedImagePath, 900, 700, 80) // Max width 900px, height 700px, quality 85
			if err != nil {
				log.Printf("Failed to compress image: %v", err)
				RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to compress image: %v", err))
				return
			}

			// Remove the original uncompressed image
			err = os.Remove(imagePath)
			if err != nil {
				log.Printf("Failed to delete original image: %v", err)
			}

			imageURL = "/" + compressedImagePath

		} else if err != http.ErrMissingFile {
			log.Printf("Error retrieving file: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to save file: %v", err))
			return
		}

		// Save the post with optional image URL
		err = database.CreatePostWithImage(userID, title, content, category, imageURL)
		if err != nil {
			log.Printf("Failed to create post: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error creating post: %v", err))
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// EditPostHandler handles editing an existing post
func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("id")

	// Convert postID from string to int
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Post ID: %v", err))
		return
	}

	if r.Method == http.MethodGet {
		post, err := database.GetPost(postID)
		if err != nil || post == nil {
			RenderErrorPage(w, http.StatusNotFound, fmt.Errorf("post not found: %v", err))
			return
		}

		categories, err := database.GetCategories()
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to load categories: %v", err))
			return
		}

		// Debugging: Log the retrieved post
		fmt.Printf("Post data: %+v\n", post)

		data := map[string]interface{}{
			"Title":      "Edit Post - Forum",
			"IsLoggedIn": true,
			"Post":       post,
			"Categories": categories,
		}

		// Debugging: Log the template data
		fmt.Printf("Template data: %+v\n", data)

		err = templates.ExecuteTemplate(w, "edit-post", data)
		if err != nil {
			fmt.Printf("Error executing edit-post template: %v\n", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		}
		return
	}

	// Handle post updates (if using POST)
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		category := r.FormValue("category")

		err := database.UpdatePost(postID, title, content, category)
		if err != nil {
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error updating post: %v", err))
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// DeletePostHandler handles deleting a post
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("id")

	// Convert postID from string to int
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Post ID: %v", err))
		return
	}

	err = database.DeletePost(postID)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error deleting post: %v", err))
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ViewPostHandler displays a specific post
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(postID)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid post ID: %s", r.URL.Path))
		return
	}

	// Get post details
	post, err := database.GetPostByID(id)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound, fmt.Errorf("page not found: %s", r.URL.Path))
		return
	}

	// Get latest reaction counts
	likes, dislikes, err := database.GetReactionCounts(id)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error getting reaction counts: %v", err))
		return
	}

	// Update post with latest counts
	post.Likes = likes
	post.Dislikes = dislikes

	// Get comments count
	commentsCount, err := database.GetCommentsCount(id)
	if err != nil {
		commentsCount = 0 // Default to 0 if error
	}
	post.CommentsCount = commentsCount

	comments, err := database.GetCommentsByPostID(id)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving comments: %v", err))
		return
	}

	// Get category counts
	categoryCounts, err := database.GetCategoryPostCounts()
	if err != nil {
		// Log the error but don't fail completely
		fmt.Printf("Error getting category counts: %v\n", err)
		categoryCounts = make(map[string]int) // Initialize empty map as fallback
	}

	// Check if user is authenticated
	userID := r.Context().Value(userIDKey)
	isAuthenticated := userID != nil

	// Get username if authenticated
	var username string
	if isAuthenticated {
		username = getUsername(userID.(int))
	}

	// Add username and time ago for each comment
	for i := range comments {
		comments[i].Username = getUsername(comments[i].UserID)
		comments[i].CreatedAtHuman = utils.TimeAgo(comments[i].CreatedAt)
	}

	data := map[string]interface{}{
		"Title":           post.Title + " - Forum",
		"IsLoggedIn":      isAuthenticated,
		"Username":        username,
		"Post":            post,
		"Comments":        comments,
		"IsAuthenticated": isAuthenticated,
		"UserID":          userID,
		"Categories":      categoryCounts,
	}

	err = templates.ExecuteTemplate(w, "view-post", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		return
	}
}

// LikePostHandler handles liking a post.
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	postIDStr := r.URL.Query().Get("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Post ID: %v", err))
		return
	}

	err = database.ToggleReaction(userID, postID, "like")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to like post: %v", err))
		return
	}

	// Return updated reaction counts
	likes, dislikes, _ := database.GetReactionCounts(postID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"likes": %d, "dislikes": %d}`, likes, dislikes)))
}

// DislikePostHandler handles disliking a post.
func DislikePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	postIDStr := r.URL.Query().Get("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("invalid Post ID: %v", err))
		return
	}

	err = database.ToggleReaction(userID, postID, "dislike")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to dislike post: %v", err))
		return
	}

	// Return updated reaction counts
	likes, dislikes, _ := database.GetReactionCounts(postID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"likes": %d, "dislikes": %d}`, likes, dislikes)))
}
