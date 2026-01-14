package utils

import (
	"path/filepath"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "test",
			expected: false,
		},
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item not exists",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "single element match",
			slice:    []string{"test"},
			item:     "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSafePath(t *testing.T) {
	tests := []struct {
		name     string
		folder   string
		filename string
		expected string
	}{
		{
			name:     "simple filename",
			folder:   "/shared",
			filename: "test.pdf",
			expected: filepath.Join("/shared", "test.pdf"),
		},
		{
			name:     "path traversal attempt",
			folder:   "/shared",
			filename: "../../../etc/passwd",
			expected: filepath.Join("/shared", "passwd"),
		},
		{
			name:     "absolute path attempt",
			folder:   "/shared",
			filename: "/etc/passwd",
			expected: filepath.Join("/shared", "passwd"),
		},
		{
			name:     "nested path traversal",
			folder:   "/shared",
			filename: "foo/../../../bar",
			expected: filepath.Join("/shared", "bar"),
		},
		// Note: Windows path separators are only treated as separators on Windows
		// On Unix systems, backslash is a valid filename character
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafePath(tt.folder, tt.filename)
			if result != tt.expected {
				t.Errorf("SafePath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFillString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		length   int
		padding  byte
		expected string
	}{
		{
			name:     "needs padding",
			s:        "test",
			length:   10,
			padding:  ':',
			expected: "test::::::",
		},
		{
			name:     "exact length",
			s:        "test",
			length:   4,
			padding:  ':',
			expected: "test",
		},
		{
			name:     "longer than length",
			s:        "teststring",
			length:   4,
			padding:  ':',
			expected: "test",
		},
		{
			name:     "empty string",
			s:        "",
			length:   5,
			padding:  ':',
			expected: ":::::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FillString(tt.s, tt.length, tt.padding)
			if result != tt.expected {
				t.Errorf("FillString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTrimPadding(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		padding  string
		expected string
	}{
		{
			name:     "with padding",
			s:        "test::::::",
			padding:  ":",
			expected: "test",
		},
		{
			name:     "no padding",
			s:        "test",
			padding:  ":",
			expected: "test",
		},
		{
			name:     "all padding",
			s:        "::::::",
			padding:  ":",
			expected: "",
		},
		{
			name:     "empty string",
			s:        "",
			padding:  ":",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimPadding(tt.s, tt.padding)
			if result != tt.expected {
				t.Errorf("TrimPadding() = %q, want %q", result, tt.expected)
			}
		})
	}
}
