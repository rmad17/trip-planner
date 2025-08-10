package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorageConfig holds configuration for local storage
type LocalStorageConfig struct {
	BasePath   string `json:"base_path" yaml:"base_path"`       // Base directory for file storage
	BaseURL    string `json:"base_url" yaml:"base_url"`         // Base URL for serving files
	CreatePath bool   `json:"create_path" yaml:"create_path"`   // Whether to create directory if not exists
}

// LocalStorageProvider implements StorageProvider for local filesystem
type LocalStorageProvider struct {
	config LocalStorageConfig
}

// NewLocalStorageProvider creates a new local storage provider
func NewLocalStorageProvider(config LocalStorageConfig) (*LocalStorageProvider, error) {
	// Set default values
	if config.BasePath == "" {
		config.BasePath = "./uploads"
	}
	if config.BaseURL == "" {
		config.BaseURL = "/uploads"
	}
	
	// Create directory if it doesn't exist and createPath is true
	if config.CreatePath {
		if err := os.MkdirAll(config.BasePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create base directory: %v", err)
		}
	}
	
	// Check if directory exists
	if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("base directory does not exist: %s", config.BasePath)
	}
	
	return &LocalStorageProvider{
		config: config,
	}, nil
}

// Upload uploads a file to local storage
func (l *LocalStorageProvider) Upload(key string, file io.Reader, contentType string) (*UploadResult, error) {
	// Create full file path
	fullPath := filepath.Join(l.config.BasePath, key)
	
	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	
	// Create file
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()
	
	// Copy data
	size, err := io.Copy(f, file)
	if err != nil {
		os.Remove(fullPath) // Clean up on error
		return nil, fmt.Errorf("failed to write file: %v", err)
	}
	
	return &UploadResult{
		Key:         key,
		URL:         l.GetPublicURL(key),
		Size:        size,
		ContentType: contentType,
		Provider:    l.GetProvider(),
	}, nil
}

// Download downloads a file from local storage
func (l *LocalStorageProvider) Download(key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.config.BasePath, key)
	
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	return file, nil
}

// Delete removes a file from local storage
func (l *LocalStorageProvider) Delete(key string) error {
	fullPath := filepath.Join(l.config.BasePath, key)
	
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	
	// Try to remove empty directories
	dir := filepath.Dir(fullPath)
	os.Remove(dir) // Ignore error - directory might not be empty
	
	return nil
}

// GetURL generates a presigned URL for accessing the file
// For local storage, this returns the same as GetPublicURL since we don't have presigned URLs
func (l *LocalStorageProvider) GetURL(key string, expiry time.Duration) (string, error) {
	return l.GetPublicURL(key), nil
}

// GetPublicURL returns a public URL for the file
func (l *LocalStorageProvider) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", l.config.BaseURL, key)
}

// GetProvider returns the provider name
func (l *LocalStorageProvider) GetProvider() string {
	return "local"
}