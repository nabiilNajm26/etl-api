package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword checks if password meets requirements
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// ValidateTableName checks if table name is valid
func ValidateTableName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("table name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("table name must be less than 255 characters")
	}

	return nil
}

// SanitizeTableName creates a safe PostgreSQL table name
func SanitizeTableName(userID, tableName string) string {
	// Remove non-alphanumeric characters and spaces
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(tableName, "_")
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	// Trim underscores from start and end
	sanitized = strings.Trim(sanitized, "_")
	
	// Ensure it doesn't start with a number
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "table_" + sanitized
	}
	
	// Add user prefix to avoid conflicts
	return fmt.Sprintf("user_%s_%s", userID[:8], sanitized)
}