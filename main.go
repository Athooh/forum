package main

import (
	"log"
	"net/http"

	"forum/database"
	"forum/handlers"
)

func main() {
	// Initialize database
	database.InitDatabase()
	defer database.DB.Close()

	// Routes
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	log.Println("Registering routes...")

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreatePostHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetPostsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	log.Println("Route /posts registered")

	http.HandleFunc("/posts/all", handlers.GetPostsHandler)
	log.Println("Route /posts/all registered")

	http.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.AddCommentHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetCommentsHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reactions", handlers.AddReactionHandler)        // POST for adding reactions
	http.HandleFunc("/reactions/count", handlers.GetReactionsHandler) // GET for fetching reaction counts

	http.HandleFunc("/categories", handlers.AddCategoryHandler)       // POST for adding categories
	http.HandleFunc("/categories/all", handlers.GetCategoriesHandler) // GET for fetching categories

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
