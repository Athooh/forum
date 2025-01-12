package utils

import (
	"testing"
	"time"
)

// TestHashPassword tests the HashPassword function.
func TestHashPassword(t *testing.T) {
	password := "StrongPass123!"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hashedPassword == "" {
		t.Fatalf("expected hashed password, got empty string")
	}
}

// TestCheckPassword tests the CheckPassword function.
func TestCheckPassword(t *testing.T) {
	password := "StrongPass123!"
	hashedPassword, _ := HashPassword(password)

	err := CheckPassword(hashedPassword, password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = CheckPassword(hashedPassword, "WrongPass123!")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestGenerateSessionToken tests the GenerateSessionToken function.
func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected token, got empty string")
	}	
}

// TestValidatePasswordStrength tests the ValidatePasswordStrength function.
func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"Short1!", true},
		{"NoNumber!", true},
		{"nouppercase1!", true},
		{"NOLOWERCASE1!", true},
		{"NoSpecial1", true},
		{"Valid1Password!", false},
	}

	for _, tt := range tests {
		err := ValidatePasswordStrength(tt.password)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePasswordStrength(%v) error = %v, wantErr %v", tt.password, err, tt.wantErr)
		}
	}
}

// TestValidateEmail tests the ValidateEmail function.
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"plainaddress", true},
		{"@missingusername.com", true},
		{"username@.com", true},
		{"username@domain", true},
		{"username@domain.com", false},
		{"user.name+tag+sorting@example.com", false},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEmail(%v) error = %v, wantErr %v", tt.email, err, tt.wantErr)
		}
	}
}

// TestTimeAgo tests the TimeAgo function.
func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		timestamp time.Time
		expected  string
	}{
		{now.Add(-10 * time.Second), "10 secs ago"},
		{now.Add(-1 * time.Minute), "1 mins ago"},
		{now.Add(-1 * time.Hour), "1 hrs ago"},
		{now.Add(-24 * time.Hour), "1 days ago"},
		{now.Add(-7 * 24 * time.Hour), "1 weeks ago"},
		{now.Add(-30 * 24 * time.Hour), "1 months ago"},
		{now.Add(-365 * 24 * time.Hour), "1 years ago"},
	}

	for _, tt := range tests {
		result := TimeAgo(tt.timestamp)
		if result != tt.expected {
			t.Errorf("TimeAgo(%v) = %v, expected %v", tt.timestamp, result, tt.expected)
		}
	}
}

// TestTruncateContent tests the TruncateContent function.
func TestTruncateContent(t *testing.T) {
	tests := []struct {
		content   string
		wordLimit int
		expected  string
	}{
		{"This is a test content", 3, "This is a..."},
		{"Short content", 5, "Short content"},
	}

	for _, tt := range tests {
		result := TruncateContent(tt.content, tt.wordLimit)
		if result != tt.expected {
			t.Errorf("TruncateContent(%v, %v) = %v, expected %v", tt.content, tt.wordLimit, result, tt.expected)
		}
	}
}
