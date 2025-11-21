package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	MaxTitleLength   = 200
	MaxContentLength = 10 * 1024 * 1024 // 10MB
	MaxTagLength     = 50
	MaxTagsPerNote   = 20
)

func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	if len(username) > 50 {
		return fmt.Errorf("username too long (max 50 characters)")
	}
	
	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			(r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("username contains invalid characters")
		}
	}
	
	return nil
}

func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	
	if !utf8.ValidString(title) {
		return fmt.Errorf("title contains invalid UTF-8")
	}
	
	if len(title) > MaxTitleLength {
		return fmt.Errorf("title too long (max %d characters)", MaxTitleLength)
	}
	
	// Check for path traversal attempts
	if strings.Contains(title, "..") || strings.Contains(title, "/") || 
		strings.Contains(title, "\\") {
		return fmt.Errorf("title contains invalid characters")
	}
	
	return nil
}

func ValidateContent(content string) error {
	if !utf8.ValidString(content) {
		return fmt.Errorf("content contains invalid UTF-8")
	}
	
	if len(content) > MaxContentLength {
		return fmt.Errorf("content too long (max %d bytes)", MaxContentLength)
	}
	
	return nil
}

func ValidateTags(tags []string) error {
	if len(tags) > MaxTagsPerNote {
		return fmt.Errorf("too many tags (max %d)", MaxTagsPerNote)
	}
	
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return fmt.Errorf("tag cannot be empty")
		}
		
		if len(tag) > MaxTagLength {
			return fmt.Errorf("tag too long (max %d characters): %s", MaxTagLength, tag)
		}
		
		// Check for valid characters
		for _, r := range tag {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
				(r >= '0' && r <= '9') || r == '_' || r == '-' || r == ' ') {
				return fmt.Errorf("tag contains invalid characters: %s", tag)
			}
		}
	}
	
	return nil
}

func SanitizePath(path string) (string, error) {
	// Resolve to absolute path and check for path traversal
	_, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	
	// Check for path traversal
	cleaned := filepath.Clean(path)
	if cleaned != path && cleaned != filepath.Join(".", path) {
		return "", fmt.Errorf("path traversal detected")
	}
	
	return cleaned, nil
}

func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	// Check for invalid characters
	invalidChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}
	
	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}
	
	return nil
}

