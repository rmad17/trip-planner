package storage

import (
	"fmt"
	"os"
)

var GlobalStorageManager *StorageManager

// InitializeStorage initializes the global storage manager from environment variables
func InitializeStorage() error {
	sm := NewStorageManager()

	provider := os.Getenv("STORAGE_PROVIDER")
	if provider == "" {
		provider = "local"
	}

	switch provider {
	case "local":
		localProvider, err := NewLocalStorageProvider(LocalStorageConfig{
			BasePath:   "./uploads",
			BaseURL:    "/uploads",
			CreatePath: true,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize local storage: %v", err)
		}
		sm.RegisterProvider("local", localProvider)
		if err := sm.SetDefault("local"); err != nil {
			return err
		}

	case "digitalocean":
		doConfig := DigitalOceanConfig{
			AccessKeyID:     os.Getenv("DO_SPACES_ACCESS_KEY"),
			SecretAccessKey: os.Getenv("DO_SPACES_SECRET_KEY"),
			Region:          os.Getenv("DO_SPACES_REGION"),
			Bucket:          os.Getenv("DO_SPACES_BUCKET"),
			Endpoint:        os.Getenv("DO_SPACES_ENDPOINT"),
		}

		// Validate required fields
		if doConfig.AccessKeyID == "" || doConfig.SecretAccessKey == "" ||
			doConfig.Region == "" || doConfig.Bucket == "" {
			return fmt.Errorf("missing required DigitalOcean Spaces configuration")
		}

		doProvider, err := NewDigitalOceanProvider(doConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize DigitalOcean storage: %v", err)
		}
		sm.RegisterProvider("digitalocean", doProvider)
		if err := sm.SetDefault("digitalocean"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported storage provider: %s", provider)
	}

	GlobalStorageManager = sm
	return nil
}

// GetDefaultProvider returns the default storage provider from the global manager
func GetDefaultProvider() (StorageProvider, error) {
	if GlobalStorageManager == nil {
		return nil, fmt.Errorf("storage manager not initialized")
	}
	return GlobalStorageManager.GetDefaultProvider()
}

// GetProvider returns a storage provider by name from the global manager
func GetProvider(name string) (StorageProvider, error) {
	if GlobalStorageManager == nil {
		return nil, fmt.Errorf("storage manager not initialized")
	}
	return GlobalStorageManager.GetProvider(name)
}
