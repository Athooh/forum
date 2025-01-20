package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		category := r.Form["category[]"]

		if title == "" || content == "" || len(category) == 0 {
			log.Println("Missing required fields")
			RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("title, content, and category are required"))
			return
		}

		// Handle image upload
		var imageURL string
		file, header, err := r.FormFile("image")
		log.Printf("File upload attempt - Header: %+v, Error: %v", header, err)

		if err == nil && header != nil {
			defer file.Close()

			// Add debug logging for content type
			fileType := header.Header.Get("Content-Type")
			log.Printf("File type detected: %s", fileType)

			// Read the file into memory
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				log.Printf("Error reading file: %v", err)
				RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to read file: %v", err))
				return
			}

			// Generate unique filename with original extension
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				// If no extension provided, derive it from content type
				switch fileType {
				case "image/jpeg", "image/jpg":
					ext = ".jpg"
				case "image/png":
					ext = ".png"
				case "image/gif":
					ext = ".gif"
				case "image/webp":
					ext = ".webp"
				case "image/bmp":
					ext = ".bmp"
				default:
					log.Printf("Unsupported file type: %s", fileType)
					RenderErrorPage(w, http.StatusBadRequest, fmt.Errorf("unsupported file type: %s", fileType))
					return
				}
			}

			newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			uploadPath := filepath.Join("uploads", newFilename)

			// Save the original file
			err = os.WriteFile(uploadPath, fileBytes, 0644)
			if err != nil {
				log.Printf("Error saving file: %v", err)
				RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to save file: %v", err))
				return
			}

			// Set the image URL
			imageURL = "/uploads/" + newFilename
			log.Printf("File successfully saved. Image URL: %s", imageURL)
		} else if err != http.ErrMissingFile {
			log.Printf("Error retrieving file: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to save file: %v", err))
			return
		}

		// Create post in database
		err = database.CreatePostWithImage(userID, title, content, category, imageURL)
		if err != nil {
			log.Printf("Error creating post in database: %v", err)
			RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to create post: %v", err))
			return
		}

		log.Printf("Post successfully created with image URL: %s", imageURL)
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
		category := r.Form["category[]"]

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
		comments[i].Username = getUsername(int(comments[i].UserID))
		comments[i].CreatedAtHuman = utils.TimeAgo(comments[i].CreatedAt)
	}

	log.Printf("Post details - ID: %d, ImageURL: %s", id, post.ImageURL)

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

func UserPostHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		log.Println("User ID not found in context")
		RenderErrorPage(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	posts, err := database.GetPostsByUserID(userID)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve posts: %v", err))
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
		posts[i].CategoryArray = strings.Split(posts[i].Category, ",")
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
