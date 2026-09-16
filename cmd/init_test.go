package cmd

import (
	"microgit/utils"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInitCommand(t *testing.T) {
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

	// Test case 1: Initialize a new repository
	t.Run("Initialize new repository", func(t *testing.T) {
		// Run the init command
		initCmd.RunE(nil, nil)

		// Check if .microgit directory was created
		if _, err := os.Stat(utils.DEFAULT_PATH); os.IsNotExist(err) {
			t.Errorf("Expected .microgit directory to be created")
		}

		// Check if objects directory was created
		objectsDir := filepath.Join(utils.DEFAULT_PATH, "objects")
		if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
			t.Errorf("Expected objects directory to be created")
		}

		// Check if index file was created
		indexFile := filepath.Join(utils.DEFAULT_PATH, "index")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			t.Errorf("Expected index file to be created")
		}

		// Check if HEAD file was created
		headFile := filepath.Join(utils.DEFAULT_PATH, "HEAD")
		if _, err := os.Stat(headFile); os.IsNotExist(err) {
			t.Errorf("Expected HEAD file to be created")
		}

		// Check if LATEST file was created
		latestFile := filepath.Join(utils.DEFAULT_PATH, "LATEST")
		if _, err := os.Stat(latestFile); os.IsNotExist(err) {
			t.Errorf("Expected LATEST file to be created")
		}

		// Check file permissions (Windows doesn't use Unix permission bits).
		if runtime.GOOS != "windows" {
			if info, err := os.Stat(utils.DEFAULT_PATH); err == nil {
				if info.Mode().Perm() != 0755 {
					t.Errorf("Expected .microgit directory to have permissions 0755, got %v", info.Mode().Perm())
				}
			}
		}
	})

	// Test case 2: Try to initialize an already initialized repository
	t.Run("Initialize existing repository", func(t *testing.T) {
		// Run the init command again
		initCmd.RunE(nil, nil)

		// Verify that no duplicate files were created
		entries, err := os.ReadDir(utils.DEFAULT_PATH)
		if err != nil {
			t.Fatalf("Failed to read .microgit directory: %v", err)
		}

		expectedEntries := map[string]bool{
			"objects": true,
			"index":   true,
			"HEAD":    true,
			"LATEST":  true,
		}

		for _, entry := range entries {
			if !expectedEntries[entry.Name()] {
				t.Errorf("Unexpected file/directory found: %s", entry.Name())
			}
		}
	})
}
