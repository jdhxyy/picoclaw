package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sipeed/picoclaw/pkg/media"
)

const FileStorageDir = "files"

// FileStorage implements media.MediaStore interface with persistent file storage.
// Files are stored in workspace/files/ directory with their original filenames.
type FileStorage struct {
	workspaceDir string
}

// NewFileStorage creates a new FileStorage instance.
// workspaceDir is the path to the workspace directory (e.g., ./.picoclaw/workspace).
func NewFileStorage(workspaceDir string) (*FileStorage, error) {
	fileDir := filepath.Join(workspaceDir, FileStorageDir)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create file storage directory: %w", err)
	}
	return &FileStorage{workspaceDir: workspaceDir}, nil
}

// Store saves a local file to the persistent storage.
// Returns a media:// reference to the stored file.
// If a file with the same name exists, it will be overwritten (cover strategy).
func (s *FileStorage) Store(localPath string, meta media.MediaMeta, scope string) (string, error) {
	// Read the source file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Sanitize filename and create destination path
	safeName := SanitizeFilename(meta.Filename)
	destPath := filepath.Join(s.workspaceDir, FileStorageDir, safeName)

	// Write file (will overwrite if exists)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return media:// reference
	return "media://" + safeName, nil
}

// Resolve returns the local file path for a given media:// reference.
func (s *FileStorage) Resolve(ref string) (string, error) {
	filename := strings.TrimPrefix(ref, "media://")
	if filename == ref {
		return "", fmt.Errorf("invalid media ref format: %s", ref)
	}

	path := filepath.Join(s.workspaceDir, FileStorageDir, filename)

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("media not found: %s", ref)
	}
	return path, nil
}

// ResolveWithMeta returns the local file path and metadata for a given media:// reference.
func (s *FileStorage) ResolveWithMeta(ref string) (string, media.MediaMeta, error) {
	path, err := s.Resolve(ref)
	if err != nil {
		return "", media.MediaMeta{}, err
	}

	// Extract filename from path
	filename := filepath.Base(path)
	return path, media.MediaMeta{Filename: filename}, nil
}

// ReleaseAll is a no-op for FileStorage since files are persistent.
// This method exists to satisfy the MediaStore interface.
func (s *FileStorage) ReleaseAll(scope string) error {
	// Files are persistent, no cleanup needed
	return nil
}

// SanitizeFilename sanitizes a filename for safe storage.
// It removes path separators, replaces dangerous characters, and handles reserved names.
func SanitizeFilename(filename string) string {
	// Get base filename (remove any path)
	filename = filepath.Base(filename)

	// Replace dangerous characters
	dangerous := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	for old, new := range dangerous {
		filename = strings.ReplaceAll(filename, old, new)
	}

	// Remove control characters
	filename = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(filename, "")

	// Limit length (preserve extension)
	const maxLen = 200
	if len(filename) > maxLen {
		ext := filepath.Ext(filename)
		name := filename[:len(filename)-len(ext)]
		if len(name) > maxLen-len(ext) {
			name = name[:maxLen-len(ext)]
		}
		filename = name + ext
	}

	// Handle empty filename
	if filename == "" || filename == "." {
		filename = "unnamed_file"
	}

	// Handle Windows reserved names
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4",
		"COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4",
		"LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}

	nameWithoutExt := filename
	ext := filepath.Ext(filename)
	if ext != "" {
		nameWithoutExt = filename[:len(filename)-len(ext)]
	}

	for _, r := range reserved {
		if strings.ToUpper(nameWithoutExt) == r {
			filename = "_" + filename
			break
		}
	}

	return filename
}

// GetFilePath returns the full path for a filename in the file storage.
func GetFilePath(workspaceDir, filename string) string {
	safeName := SanitizeFilename(filename)
	return filepath.Join(workspaceDir, FileStorageDir, safeName)
}

// FileExists checks if a file exists in the file storage.
func FileExists(workspaceDir, filename string) bool {
	path := GetFilePath(workspaceDir, filename)
	_, err := os.Stat(path)
	return err == nil
}
