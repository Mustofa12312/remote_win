package system

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DirEntry represents a file or directory
type DirEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Ext   string `json:"ext"`
}

// DirListing is the result of listing a directory
type DirListing struct {
	Path    string     `json:"path"`
	Entries []DirEntry `json:"entries"`
	Error   string     `json:"error,omitempty"`
}

// ListDir lists the contents of a directory
func ListDir(path string) DirListing {
	if path == "" {
		path = "/"
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return DirListing{Path: path, Error: err.Error()}
	}

	var result []DirEntry
	for _, e := range entries {
		info, _ := e.Info()
		size := int64(0)
		if info != nil && !e.IsDir() {
			size = info.Size()
		}
		result = append(result, DirEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Size:  size,
			Ext:   strings.TrimPrefix(filepath.Ext(e.Name()), "."),
		})
	}

	return DirListing{Path: path, Entries: result}
}

// SearchFiles searches recursively for files matching pattern
func SearchFiles(root, pattern string) ([]string, error) {
	if root == "" {
		root = "/"
	}
	pattern = strings.ToLower(pattern)
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if strings.Contains(strings.ToLower(info.Name()), pattern) {
			matches = append(matches, path)
		}
		if len(matches) >= 50 { // limit results
			return fmt.Errorf("limit reached")
		}
		return nil
	})
	if err != nil && err.Error() != "limit reached" && len(matches) == 0 {
		return nil, err
	}
	return matches, nil
}

// DownloadFile reads a file and returns its bytes
func DownloadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	// Limit download to 50MB
	const maxSize = 50 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(f, maxSize))
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return data, nil
}

// UploadFile writes bytes to a path
func UploadFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Helper to read a file (used by screenshot)
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// CommonPaths returns common paths for quick navigation
func CommonPaths() map[string]string {
	home, _ := os.UserHomeDir()
	return map[string]string{
		"Home":      home,
		"Desktop":   filepath.Join(home, "Desktop"),
		"Downloads": filepath.Join(home, "Downloads"),
		"Documents": filepath.Join(home, "Documents"),
		"Root":      "/",
	}
}
