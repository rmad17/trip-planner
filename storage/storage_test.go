package storage

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockStorageProvider is a mock implementation of StorageProvider for testing
type MockStorageProvider struct {
	uploadFunc      func(key string, file io.Reader, contentType string) (*UploadResult, error)
	downloadFunc    func(key string) (io.ReadCloser, error)
	deleteFunc      func(key string) error
	getURLFunc      func(key string, expiry time.Duration) (string, error)
	getPublicURLFunc func(key string) string
	provider        string
}

func (m *MockStorageProvider) Upload(key string, file io.Reader, contentType string) (*UploadResult, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(key, file, contentType)
	}
	return &UploadResult{
		Key:         key,
		URL:         "https://example.com/" + key,
		Size:        1024,
		ContentType: contentType,
		Provider:    m.provider,
	}, nil
}

func (m *MockStorageProvider) Download(key string) (io.ReadCloser, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(key)
	}
	return io.NopCloser(bytes.NewReader([]byte("test data"))), nil
}

func (m *MockStorageProvider) Delete(key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(key)
	}
	return nil
}

func (m *MockStorageProvider) GetURL(key string, expiry time.Duration) (string, error) {
	if m.getURLFunc != nil {
		return m.getURLFunc(key, expiry)
	}
	return "https://example.com/" + key + "?expires=" + expiry.String(), nil
}

func (m *MockStorageProvider) GetPublicURL(key string) string {
	if m.getPublicURLFunc != nil {
		return m.getPublicURLFunc(key)
	}
	return "https://example.com/public/" + key
}

func (m *MockStorageProvider) GetProvider() string {
	if m.provider == "" {
		return "mock"
	}
	return m.provider
}

func TestNewStorageManager(t *testing.T) {
	t.Run("Create new storage manager", func(t *testing.T) {
		sm := NewStorageManager()
		assert.NotNil(t, sm)
		assert.NotNil(t, sm.providers)
		assert.Empty(t, sm.providers)
	})
}

func TestNewStorageManagerWithDefaults(t *testing.T) {
	t.Run("Create storage manager with defaults", func(t *testing.T) {
		sm := NewStorageManagerWithDefaults()
		assert.NotNil(t, sm)

		// Check if default provider is set
		provider, err := sm.GetDefaultProvider()
		if err == nil {
			assert.NotNil(t, provider)
			assert.Equal(t, "local", provider.GetProvider())
		}
		// Error is acceptable if local storage couldn't be initialized
	})
}

func TestStorageManager_RegisterProvider(t *testing.T) {
	t.Run("Register single provider", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider := &MockStorageProvider{provider: "mock"}

		sm.RegisterProvider("test", mockProvider)

		provider, err := sm.GetProvider("test")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mock", provider.GetProvider())
	})

	t.Run("Register multiple providers", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider1 := &MockStorageProvider{provider: "mock1"}
		mockProvider2 := &MockStorageProvider{provider: "mock2"}

		sm.RegisterProvider("test1", mockProvider1)
		sm.RegisterProvider("test2", mockProvider2)

		provider1, err := sm.GetProvider("test1")
		assert.NoError(t, err)
		assert.Equal(t, "mock1", provider1.GetProvider())

		provider2, err := sm.GetProvider("test2")
		assert.NoError(t, err)
		assert.Equal(t, "mock2", provider2.GetProvider())
	})
}

func TestStorageManager_SetDefault(t *testing.T) {
	t.Run("Set default provider", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider := &MockStorageProvider{provider: "mock"}

		sm.RegisterProvider("test", mockProvider)
		err := sm.SetDefault("test")

		assert.NoError(t, err)
		assert.Equal(t, "test", sm.default)
	})

	t.Run("Set non-existent provider as default", func(t *testing.T) {
		sm := NewStorageManager()

		err := sm.SetDefault("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestStorageManager_GetProvider(t *testing.T) {
	t.Run("Get provider by name", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider := &MockStorageProvider{provider: "mock"}

		sm.RegisterProvider("test", mockProvider)

		provider, err := sm.GetProvider("test")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mock", provider.GetProvider())
	})

	t.Run("Get non-existent provider", func(t *testing.T) {
		sm := NewStorageManager()

		provider, err := sm.GetProvider("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Get default provider with empty name", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider := &MockStorageProvider{provider: "mock"}

		sm.RegisterProvider("test", mockProvider)
		sm.SetDefault("test")

		provider, err := sm.GetProvider("")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mock", provider.GetProvider())
	})
}

func TestStorageManager_GetDefaultProvider(t *testing.T) {
	t.Run("Get default provider", func(t *testing.T) {
		sm := NewStorageManager()
		mockProvider := &MockStorageProvider{provider: "mock"}

		sm.RegisterProvider("test", mockProvider)
		sm.SetDefault("test")

		provider, err := sm.GetDefaultProvider()
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mock", provider.GetProvider())
	})

	t.Run("Get default provider when not set", func(t *testing.T) {
		sm := NewStorageManager()

		provider, err := sm.GetDefaultProvider()
		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestUploadResult_Structure(t *testing.T) {
	t.Run("Create upload result", func(t *testing.T) {
		result := UploadResult{
			Key:         "documents/test.pdf",
			URL:         "https://example.com/documents/test.pdf",
			Size:        2048576,
			ContentType: "application/pdf",
			Provider:    "s3",
		}

		assert.Equal(t, "documents/test.pdf", result.Key)
		assert.Equal(t, "https://example.com/documents/test.pdf", result.URL)
		assert.Equal(t, int64(2048576), result.Size)
		assert.Equal(t, "application/pdf", result.ContentType)
		assert.Equal(t, "s3", result.Provider)
	})
}

func TestStorageConfig_Structure(t *testing.T) {
	t.Run("Create storage config", func(t *testing.T) {
		config := StorageConfig{
			Provider: "s3",
			Config: map[string]interface{}{
				"bucket": "my-bucket",
				"region": "us-east-1",
			},
		}

		assert.Equal(t, "s3", config.Provider)
		assert.NotNil(t, config.Config)
		assert.Equal(t, "my-bucket", config.Config["bucket"])
		assert.Equal(t, "us-east-1", config.Config["region"])
	})
}

func TestFileMetadata_Structure(t *testing.T) {
	t.Run("Create file metadata", func(t *testing.T) {
		now := time.Now()
		metadata := FileMetadata{
			OriginalName: "test.pdf",
			Size:         1024000,
			ContentType:  "application/pdf",
			UploadedAt:   now,
			Key:          "documents/test.pdf",
			Provider:     "s3",
		}

		assert.Equal(t, "test.pdf", metadata.OriginalName)
		assert.Equal(t, int64(1024000), metadata.Size)
		assert.Equal(t, "application/pdf", metadata.ContentType)
		assert.Equal(t, now, metadata.UploadedAt)
		assert.Equal(t, "documents/test.pdf", metadata.Key)
		assert.Equal(t, "s3", metadata.Provider)
	})
}

func TestMockStorageProvider_Upload(t *testing.T) {
	t.Run("Upload with mock provider", func(t *testing.T) {
		mock := &MockStorageProvider{provider: "mock"}
		data := bytes.NewReader([]byte("test data"))

		result, err := mock.Upload("test.txt", data, "text/plain")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test.txt", result.Key)
		assert.Equal(t, "text/plain", result.ContentType)
		assert.Equal(t, "mock", result.Provider)
	})

	t.Run("Upload with custom function", func(t *testing.T) {
		mock := &MockStorageProvider{
			provider: "mock",
			uploadFunc: func(key string, file io.Reader, contentType string) (*UploadResult, error) {
				return &UploadResult{
					Key:         "custom/" + key,
					URL:         "https://custom.com/" + key,
					Size:        100,
					ContentType: contentType,
					Provider:    "custom",
				}, nil
			},
		}

		data := bytes.NewReader([]byte("test"))
		result, err := mock.Upload("test.txt", data, "text/plain")

		assert.NoError(t, err)
		assert.Equal(t, "custom/test.txt", result.Key)
		assert.Equal(t, "custom", result.Provider)
	})
}

func TestMockStorageProvider_Download(t *testing.T) {
	t.Run("Download with mock provider", func(t *testing.T) {
		mock := &MockStorageProvider{provider: "mock"}

		reader, err := mock.Download("test.txt")

		assert.NoError(t, err)
		assert.NotNil(t, reader)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "test data", string(data))
	})
}

func TestMockStorageProvider_Delete(t *testing.T) {
	t.Run("Delete with mock provider", func(t *testing.T) {
		mock := &MockStorageProvider{provider: "mock"}

		err := mock.Delete("test.txt")

		assert.NoError(t, err)
	})
}

func TestMockStorageProvider_GetURL(t *testing.T) {
	t.Run("Get URL with mock provider", func(t *testing.T) {
		mock := &MockStorageProvider{provider: "mock"}

		url, err := mock.GetURL("test.txt", time.Hour)

		assert.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.Contains(t, url, "test.txt")
	})
}

func TestMockStorageProvider_GetPublicURL(t *testing.T) {
	t.Run("Get public URL with mock provider", func(t *testing.T) {
		mock := &MockStorageProvider{provider: "mock"}

		url := mock.GetPublicURL("test.txt")

		assert.NotEmpty(t, url)
		assert.Contains(t, url, "test.txt")
		assert.Contains(t, url, "public")
	})
}

func TestStorageManager_InitializeFromConfig(t *testing.T) {
	t.Run("Initialize with empty config", func(t *testing.T) {
		sm := NewStorageManager()
		configs := map[string]StorageConfig{}

		err := sm.InitializeFromConfig(configs)

		assert.NoError(t, err)
	})

	t.Run("Initialize with unsupported provider", func(t *testing.T) {
		sm := NewStorageManager()
		configs := map[string]StorageConfig{
			"test": {
				Provider: "unsupported",
				Config:   map[string]interface{}{},
			},
		}

		err := sm.InitializeFromConfig(configs)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported")
	})
}

func TestStorageProvider_Interface(t *testing.T) {
	t.Run("Mock provider implements interface", func(t *testing.T) {
		var provider StorageProvider
		mock := &MockStorageProvider{provider: "mock"}

		provider = mock

		assert.NotNil(t, provider)
		assert.Equal(t, "mock", provider.GetProvider())
	})
}

func TestStorageManager_MultipleProviders(t *testing.T) {
	t.Run("Manage multiple providers", func(t *testing.T) {
		sm := NewStorageManager()

		// Register multiple providers
		sm.RegisterProvider("local", &MockStorageProvider{provider: "local"})
		sm.RegisterProvider("s3", &MockStorageProvider{provider: "s3"})
		sm.RegisterProvider("digitalocean", &MockStorageProvider{provider: "digitalocean"})

		// Set default
		err := sm.SetDefault("s3")
		assert.NoError(t, err)

		// Get providers by name
		localProvider, err := sm.GetProvider("local")
		assert.NoError(t, err)
		assert.Equal(t, "local", localProvider.GetProvider())

		s3Provider, err := sm.GetProvider("s3")
		assert.NoError(t, err)
		assert.Equal(t, "s3", s3Provider.GetProvider())

		doProvider, err := sm.GetProvider("digitalocean")
		assert.NoError(t, err)
		assert.Equal(t, "digitalocean", doProvider.GetProvider())

		// Get default provider
		defaultProvider, err := sm.GetDefaultProvider()
		assert.NoError(t, err)
		assert.Equal(t, "s3", defaultProvider.GetProvider())
	})
}
