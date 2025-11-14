package storage

import (
	"io"
	"time"
)

// StorageProvider defines the interface for different storage backends
type StorageProvider interface {
	// Upload uploads a file to the storage provider
	Upload(key string, file io.Reader, contentType string) (*UploadResult, error)

	// Download downloads a file from the storage provider
	Download(key string) (io.ReadCloser, error)

	// Delete removes a file from the storage provider
	Delete(key string) error

	// GetURL generates a presigned URL for accessing the file
	GetURL(key string, expiry time.Duration) (string, error)

	// GetPublicURL returns a public URL if the file is publicly accessible
	GetPublicURL(key string) string

	// GetProvider returns the provider name
	GetProvider() string
}

// UploadResult contains information about the uploaded file
type UploadResult struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Provider    string `json:"provider"`
}

// StorageConfig holds configuration for different storage providers
type StorageConfig struct {
	Provider string                 `json:"provider" yaml:"provider"`
	Config   map[string]interface{} `json:"config" yaml:"config"`
}

// FileMetadata contains metadata about uploaded files
type FileMetadata struct {
	OriginalName string    `json:"original_name"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Key          string    `json:"key"`
	Provider     string    `json:"provider"`
}
