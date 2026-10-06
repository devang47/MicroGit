package cmd

import (
	"microgit/utils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected string
	}{
		{
			name:     "empty content",
			content:  []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple content",
			content:  []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.HashContent(tt.content)
			if got != tt.expected {
				t.Errorf("hashContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAddCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "microgit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Init a new repo
	initCmd.RunE(nil, nil)

	// Create a test file
	testContent := []byte("test content")
	if err := os.WriteFile("test.txt", testContent, 0644); err != nil {
		t.Fatalf("error creating test file: %v", err)
	}

	// Run the add command
	addCmd.RunE(nil, []string{"test.txt"})

	// Verify the object was created
	hash := utils.HashContent(testContent)
	objectPath := utils.ObjectPath(hash)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		t.Errorf("object file was not created at %s", objectPath)
	}

	// Verify the index was updated
	indexPath := filepath.Join(utils.DEFAULT_PATH, "index")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("error reading index file: %v", err)
	}

	expected := "test.txt " + hash
	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("index file content mismatch, got %s, want %s", content, expected)
	}
}
