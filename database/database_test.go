package database

import (
	"database/sql"
	"testing"
)

func TestInitDB(t *testing.T) {
	InitDB()

	tables := []string{"users", "sessions", "posts", "comments", "categories", "post_reactions"}
	for _, table := range tables {
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		var name string
		err := DB.QueryRow(query, table).Scan(&name)
		if err == sql.ErrNoRows {
			t.Fatalf("expected table %s to be created, but it was not", table)
		} else if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != table {
			t.Fatalf("expected table name %s, got %s", table, name)
		}
	}
}

// TestCategorySeed tests the seeding of predefined categories.
func TestCategorySeed(t *testing.T) {
	InitDB()

	expectedCategories := []string{
		"Technology", "Design", "Marketing", "Development", "Science",
		"Health", "Education", "Business", "Lifestyle", "Entertainment",
	}

	for _, category := range expectedCategories {
		query := "SELECT name FROM categories WHERE name=?"
		var name string
		err := DB.QueryRow(query, category).Scan(&name)
		if err == sql.ErrNoRows {
			t.Fatalf("expected category %s to be seeded, but it was not", category)
		} else if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != category {
			t.Fatalf("expected category name %s, got %s", category, name)
		}
	}
}

// TestAddComment tests the AddComment function.
func TestAddComment(t *testing.T) {
	InitDB()

	err := AddComment(1, 1, "This is a test comment")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	query := "SELECT content FROM comments WHERE post_id=1 AND user_id=1"
	var content string
	err = DB.QueryRow(query).Scan(&content)
	if err != nil {
		t.Fatalf("expected comment to be added, but it was not")
	}
	if content != "This is a test comment" {
		t.Fatalf("expected comment content to be 'This is a test comment', got %s", content)
	}
}

// TestCreatePost tests the CreatePost function.
func TestCreatePost(t *testing.T) {
	InitDB()

	err := CreatePost(1, "Test Title", "Test Content", "Technology")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	query := "SELECT title, content, category FROM posts WHERE user_id=1"
	var title, content, category string
	err = DB.QueryRow(query).Scan(&title, &content, &category)
	if err != nil {
		t.Fatalf("expected post to be added, but it was not")
	}
	if title != "Test Title" || content != "Test Content" || category != "Technology" {
		t.Fatalf("expected post title 'Test Title', content 'Test Content', category 'Technology', got title %s, content %s, category %s", title, content, category)
	}
}
