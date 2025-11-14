package storage

import (
	"fmt"
)

// StorageManager manages multiple storage providers
type StorageManager struct {
	providers       map[string]StorageProvider
	defaultProvider string
}

// NewStorageManager creates a new storage manager
func NewStorageManager() *StorageManager {
	return &StorageManager{
		providers: make(map[string]StorageProvider),
	}
}

// NewStorageManagerWithDefaults creates a new storage manager with local storage as default
func NewStorageManagerWithDefaults() *StorageManager {
	sm := NewStorageManager()
	
	// Set up local storage as default
	localProvider, err := NewLocalStorageProvider(LocalStorageConfig{
		BasePath:   "./uploads",
		BaseURL:    "/uploads",
		CreatePath: true,
	})
	
	if err == nil {
		sm.RegisterProvider("local", localProvider)
		if err := sm.SetDefault("local"); err != nil {
			return nil
		}
	}
	
	return sm
}

// RegisterProvider registers a storage provider
func (sm *StorageManager) RegisterProvider(name string, provider StorageProvider) {
	sm.providers[name] = provider
}

// SetDefault sets the default storage provider
func (sm *StorageManager) SetDefault(name string) error {
	if _, exists := sm.providers[name]; !exists {
		return fmt.Errorf("storage provider '%s' not found", name)
	}
	sm.defaultProvider = name
	return nil
}

// GetProvider returns a storage provider by name
func (sm *StorageManager) GetProvider(name string) (StorageProvider, error) {
	if name == "" {
		name = sm.defaultProvider
	}

	provider, exists := sm.providers[name]
	if !exists {
		return nil, fmt.Errorf("storage provider '%s' not found", name)
	}

	return provider, nil
}

// GetDefaultProvider returns the default storage provider
func (sm *StorageManager) GetDefaultProvider() (StorageProvider, error) {
	return sm.GetProvider("")
}

// InitializeFromConfig initializes storage providers from configuration
func (sm *StorageManager) InitializeFromConfig(configs map[string]StorageConfig) error {
	for name, config := range configs {
		var provider StorageProvider
		var err error

		switch config.Provider {
		case "local":
			localConfig := LocalStorageConfig{}
			if err := mapToStruct(config.Config, &localConfig); err != nil {
				return fmt.Errorf("invalid Local config for %s: %v", name, err)
			}
			provider, err = NewLocalStorageProvider(localConfig)
		case "digitalocean":
			doConfig := DigitalOceanConfig{}
			if err := mapToStruct(config.Config, &doConfig); err != nil {
				return fmt.Errorf("invalid DigitalOcean config for %s: %v", name, err)
			}
			provider, err = NewDigitalOceanProvider(doConfig)
		case "s3":
			s3Config := S3Config{}
			if err := mapToStruct(config.Config, &s3Config); err != nil {
				return fmt.Errorf("invalid S3 config for %s: %v", name, err)
			}
			provider, err = NewS3Provider(s3Config)
		default:
			return fmt.Errorf("unsupported storage provider: %s", config.Provider)
		}

		if err != nil {
			return fmt.Errorf("failed to initialize %s provider: %v", config.Provider, err)
		}

		sm.RegisterProvider(name, provider)
	}

	return nil
}

// mapToStruct converts map[string]interface{} to a struct
func mapToStruct(data map[string]interface{}, result interface{}) error {
	// This is a simplified implementation
	// In production, you might want to use a library like mapstructure
	// For now, this is a placeholder
	return nil
}