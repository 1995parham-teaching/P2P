package utils

import (
	"path/filepath"
	"strings"
)

// Contains checks if a string slice contains a specific item
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SafePath ensures the file path is within the allowed folder and prevents path traversal
func SafePath(folder, filename string) string {
	// Use filepath.Base to strip any directory components (prevents ../ attacks)
	safeName := filepath.Base(filename)
	return filepath.Join(folder, safeName)
}

// FillString pads a string to the specified length with a padding character
func FillString(s string, length int, padding byte) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(string(padding), length-len(s))
}

// TrimPadding removes padding characters from the end of a string
func TrimPadding(s string, padding string) string {
	return strings.TrimRight(s, padding)
}
