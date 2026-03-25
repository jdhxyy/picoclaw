package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/media"
)

func TestFileStorage(t *testing.T) {
	// Create temp workspace
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")

	// Create FileStorage
	fs, err := NewFileStorage(workspaceDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Test Store
	t.Run("Store", func(t *testing.T) {
		// Create a temp file to store
		srcFile := filepath.Join(tempDir, "source.txt")
		content := []byte("Hello, World!")
		if err := os.WriteFile(srcFile, content, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		ref, err := fs.Store(srcFile, media.MediaMeta{
			Filename: "test.txt",
			Source:   "test",
		}, "")
		if err != nil {
			t.Fatalf("Failed to store file: %v", err)
		}

		expectedRef := "media://test.txt"
		if ref != expectedRef {
			t.Errorf("Expected ref %q, got %q", expectedRef, ref)
		}

		// Verify file exists in workspace/files/
		storedPath := filepath.Join(workspaceDir, FileStorageDir, "test.txt")
		if _, err := os.Stat(storedPath); err != nil {
			t.Errorf("Stored file not found at %s: %v", storedPath, err)
		}

		// Verify content
		storedContent, err := os.ReadFile(storedPath)
		if err != nil {
			t.Fatalf("Failed to read stored file: %v", err)
		}
		if string(storedContent) != string(content) {
			t.Errorf("Content mismatch: expected %q, got %q", content, storedContent)
		}
	})

	// Test Resolve
	t.Run("Resolve", func(t *testing.T) {
		path, err := fs.Resolve("media://test.txt")
		if err != nil {
			t.Fatalf("Failed to resolve: %v", err)
		}

		expectedPath := filepath.Join(workspaceDir, FileStorageDir, "test.txt")
		if path != expectedPath {
			t.Errorf("Expected path %q, got %q", expectedPath, path)
		}
	})

	// Test ResolveWithMeta
	t.Run("ResolveWithMeta", func(t *testing.T) {
		path, meta, err := fs.ResolveWithMeta("media://test.txt")
		if err != nil {
			t.Fatalf("Failed to resolve with meta: %v", err)
		}

		expectedPath := filepath.Join(workspaceDir, FileStorageDir, "test.txt")
		if path != expectedPath {
			t.Errorf("Expected path %q, got %q", expectedPath, path)
		}

		if meta.Filename != "test.txt" {
			t.Errorf("Expected filename %q, got %q", "test.txt", meta.Filename)
		}
	})

	// Test Cover Strategy (overwrite existing file)
	t.Run("CoverStrategy", func(t *testing.T) {
		// Create a new source file with different content
		srcFile := filepath.Join(tempDir, "source2.txt")
		newContent := []byte("New content!")
		if err := os.WriteFile(srcFile, newContent, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Store with same filename (should overwrite)
		ref, err := fs.Store(srcFile, media.MediaMeta{
			Filename: "test.txt",
			Source:   "test",
		}, "")
		if err != nil {
			t.Fatalf("Failed to store file: %v", err)
		}

		// Verify content is new
		storedPath := filepath.Join(workspaceDir, FileStorageDir, "test.txt")
		storedContent, err := os.ReadFile(storedPath)
		if err != nil {
			t.Fatalf("Failed to read stored file: %v", err)
		}
		if string(storedContent) != string(newContent) {
			t.Errorf("Content should be overwritten: expected %q, got %q", newContent, storedContent)
		}

		t.Logf("Cover strategy works: ref = %s", ref)
	})

	// Test ReleaseAll (should be no-op)
	t.Run("ReleaseAll", func(t *testing.T) {
		err := fs.ReleaseAll("")
		if err != nil {
			t.Errorf("ReleaseAll should not return error: %v", err)
		}

		// Verify file still exists
		storedPath := filepath.Join(workspaceDir, FileStorageDir, "test.txt")
		if _, err := os.Stat(storedPath); err != nil {
			t.Errorf("File should still exist after ReleaseAll: %v", err)
		}
	})
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.txt", "test.txt"},
		{"file/with/path.txt", "path.txt"},
		{"file:colon.txt", "file_colon.txt"},
		{"file*star?.txt", "file_star_.txt"},
		{"file\"quote.txt", "file_quote.txt"},
		{"file<less>greater.txt", "file_less_greater.txt"},
		{"file|pipe.txt", "file_pipe.txt"},
		{"CON.txt", "_CON.txt"},
		{"", "unnamed_file"},
		{".", "unnamed_file"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetFilePath(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	filename := "test.txt"

	path := GetFilePath(workspaceDir, filename)
	expected := filepath.Join(workspaceDir, FileStorageDir, "test.txt")

	if path != expected {
		t.Errorf("GetFilePath(%q, %q) = %q, expected %q", workspaceDir, filename, path, expected)
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")

	// Create files directory
	os.MkdirAll(filepath.Join(workspaceDir, FileStorageDir), 0755)

	// Test non-existent file
	if FileExists(workspaceDir, "nonexistent.txt") {
		t.Error("FileExists should return false for non-existent file")
	}

	// Create file
	testFile := filepath.Join(workspaceDir, FileStorageDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !FileExists(workspaceDir, "exists.txt") {
		t.Error("FileExists should return true for existing file")
	}
}
