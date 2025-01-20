package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
	"strings"

	"forum/database"
	"forum/utils"
)

var templates *template.Template
var errorTemplates *template.Template

// At the top of your handlers package, add this type definition
type contextKey string

// Define your key constant
const userIDKey contextKey = "user_id"

// InitTemplates initializes templates from the given base directory
func InitTemplates(baseDir string) error {
    var err error
    templates, err = template.ParseGlob(filepath.Join(baseDir, "templates", "*.html"))
    if err != nil {
        return fmt.Errorf("error parsing templates: %v", err)
    }

    errorTemplates, err = template.ParseGlob(filepath.Join(baseDir, "templates", "error", "*.html"))
    if err != nil {
        return fmt.Errorf("error parsing error templates: %v", err)
    }

    // Add error templates to the main template set
    for _, t := range errorTemplates.Templates() {
        if tmpl := templates.Lookup(t.Name()); tmpl == nil {
            templates, err = templates.AddParseTree(t.Name(), t.Tree)
            if err != nil {
                return fmt.Errorf("error adding template %s: %v", t.Name(), err)
            }
        }
    }

    return nil
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderErrorPage(w, http.StatusNotFound, fmt.Errorf("page not found: %s", r.URL.Path))
		return
	}

	// Get all posts
	posts, err := database.GetAllPosts("")
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
	userID := r.Context().Value(userIDKey)
	isAuthenticated := userID != nil

	// Get username if authenticated
	var username string
	if isAuthenticated {
		username = getUsername(userID.(int))
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

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SignupHandler invoked")

	userID := r.Context().Value(userIDKey).(int) // Retrieve the logged-in user's ID from the context

	posts, err := database.GetAllPosts("")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving posts: %v", err))
		return
	}

	// Retrieve the username for each post's userID
	for i := range posts {
		// Query the users table for the username using the userID of each post
		var username string
		query := `SELECT username FROM users WHERE id = ?`
		err := database.DB.QueryRow(query, posts[i].UserID).Scan(&username)
		if err != nil {
			// If there's an error retrieving the username, skip adding the username
			posts[i].Username = "Unknown"
		} else {
			posts[i].Username = username
		}
		posts[i].CategoryArray = strings.Split(posts[i].Category, ",")
	}

	// Retrieve category post counts
	categoryCounts, err := database.GetCategoryPostCounts()
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving category counts: %v", err))
		return
	}

	// Pass the posts and logged-in user's username to the template
	data := map[string]interface{}{
		"Title":      "Dashboard - Forum",
		"IsLoggedIn": true,
		"Username":   getUsername(userID), // Utility function to fetch username
		"Posts":      posts,
		"Categories": categoryCounts, // Make sure categoryCounts is used here
		"UserID":     userID,
	}

	err = templates.ExecuteTemplate(w, "dashboard", data)
	if err != nil {
		// fmt.Printf("Error executing dashboard template: %v\n", err)
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		return
	}
}

// ListPostsHandler displays all posts
func ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := database.GetAllPosts("")
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error retrieving posts: %v", err))
		return
	}

	data := map[string]interface{}{
		"Title":      "All Posts - Forum",
		"IsLoggedIn": true,
		"Posts":      posts,
	}

	err = templates.ExecuteTemplate(w, "post-list", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("error rendering template: %v", err))
		return
	}
}