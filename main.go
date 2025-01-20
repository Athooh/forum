package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/database"
	"forum/handlers"
	"forum/utils"
)

func main() {
	// initialize database
	database.InitDB()

	// Initialize upload directory
	if err := utils.InitializeUploadDirectory(); err != nil {
		log.Fatalf("Failed to initialize upload directory: %v", err)
	}

	// Initialize templates
	if err := handlers.InitTemplates("."); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	// Serve static files (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Handle uploaded files
	http.HandleFunc("/uploads/", utils.ConfigureUploadHandler())

	// Define explicit routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			handlers.NotFoundHandler(w, r)
			return
		}
		handlers.GuestMiddleware(http.HandlerFunc(handlers.HomeHandler)).ServeHTTP(w, r)
	})

	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/signup", handlers.SignUpHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Wrap the DashboardHandler with AuthMiddleware
	http.Handle("/dashboard", handlers.AuthMiddleware(http.HandlerFunc(handlers.DashboardHandler)))
	http.Handle("/create-post", handlers.AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
	http.Handle("/posts", handlers.AuthMiddleware(http.HandlerFunc(handlers.ListPostsHandler)))
	http.Handle("/view-post", handlers.GuestMiddleware(http.HandlerFunc(handlers.ViewPostHandler)))
	http.Handle("/edit-post", handlers.AuthMiddleware(
		handlers.OwnershipMiddleware(
			http.HandlerFunc(handlers.EditPostHandler))))
	http.Handle("/delete-post", handlers.AuthMiddleware(
		handlers.OwnershipMiddleware(
			http.HandlerFunc(handlers.DeletePostHandler))))
	http.Handle("/like-post", handlers.AuthMiddleware(http.HandlerFunc(handlers.LikePostHandler)))
	http.Handle("/dislike-post", handlers.AuthMiddleware(http.HandlerFunc(handlers.DislikePostHandler)))
	http.Handle("/add-comment", handlers.AuthMiddleware(http.HandlerFunc(handlers.AddCommentHandler)))
	http.Handle("/edit-comment", handlers.AuthMiddleware(http.HandlerFunc(handlers.EditCommentHandler)))
	http.Handle("/delete-comment", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteCommentHandler)))
	http.Handle("/like-comment", handlers.AuthMiddleware(http.HandlerFunc(handlers.LikeCommentHandler)))
	http.Handle("/dislike-comment", handlers.AuthMiddleware(http.HandlerFunc(handlers.DislikeCommentHandler)))

	http.Handle("/posts-by-category", handlers.AuthMiddleware(http.HandlerFunc(handlers.PostsByCategoryHandler)))
	http.Handle("/user-posts", handlers.AuthMiddleware(http.HandlerFunc(handlers.UserPostHandler)))
	http.Handle("/liked-posts", handlers.AuthMiddleware(http.HandlerFunc(handlers.FilterLikesHandler)))

	// Add this line after your other route definitions
	http.HandleFunc("/404", handlers.NotFoundHandler)

	// Add this with your other routes for testing 500 error
	// run this on the browser to see the 500 error page: http://localhost:8080/test500
	http.HandleFunc("/test500", func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderErrorPage(w, http.StatusInternalServerError, fmt.Errorf("test 500 error"))
	})

	// Start the server
	fmt.Println("Server starting at port localhost:8080")
	http.ListenAndServe(":8080", nil)
}
