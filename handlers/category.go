package handlers

import (
	"encoding/json"
	"forum/database"
	"log"
	"net/http"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// AddCategoryHandler handles adding a new category
func AddCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var category Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil || category.Name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert the category into the database
	_, err = database.DB.Exec("INSERT INTO categories (name) VALUES (?)", category.Name)
	if err != nil {
		log.Printf("Error adding category: %v", err)
		http.Error(w, "Failed to add category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Category added successfully"))
}

// GetCategoriesHandler fetches all categories
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query all categories
	rows, err := database.DB.Query("SELECT id, name FROM categories")
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare response
	var categories []Category
	for rows.Next() {
		var category Category
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			http.Error(w, "Failed to parse categories", http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// AddPostCategoryHandler handles associating a post with categories
func AddPostCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var association struct {
		PostID     int   `json:"post_id"`
		CategoryIDs []int `json:"category_ids"`
	}

	err := json.NewDecoder(r.Body).Decode(&association)
	if err != nil || association.PostID == 0 || len(association.CategoryIDs) == 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert associations into the database
	for _, categoryID := range association.CategoryIDs {
		_, err := database.DB.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", association.PostID, categoryID)
		if err != nil {
			log.Printf("Error adding post-category association: %v", err)
			http.Error(w, "Failed to associate post with categories", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Post associated with categories successfully"))
}
