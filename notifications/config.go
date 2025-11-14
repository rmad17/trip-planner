package notifications

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProviderNotFound     = errors.New("provider not found")
	ErrNoActiveProvider     = errors.New("no active provider for channel")
	ErrProviderNotHealthy   = errors.New("provider is not healthy")
	ErrChannelNotSupported  = errors.New("channel not supported")
)

// ProviderConfig represents configuration for a notification provider
type ProviderConfig struct {
	APIKey      string                 `json:"api_key,omitempty"`
	APISecret   string                 `json:"api_secret,omitempty"`
	APIEndpoint string                 `json:"api_endpoint,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	SenderID    string                 `json:"sender_id,omitempty"`
	SenderName  string                 `json:"sender_name,omitempty"`
	SenderEmail string                 `json:"sender_email,omitempty"`
	SenderPhone string                 `json:"sender_phone,omitempty"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// ProviderManager manages notification channel providers
type ProviderManager struct {
	db        *gorm.DB
	providers map[NotificationChannel][]ChannelProvider
	configs   map[string]*NotificationProvider
	mu        sync.RWMutex
}

// NewProviderManager creates a new provider manager
func NewProviderManager(db *gorm.DB) *ProviderManager {
	return &ProviderManager{
		db:        db,
		providers: make(map[NotificationChannel][]ChannelProvider),
		configs:   make(map[string]*NotificationProvider),
	}
}

// Register registers a channel provider
func (pm *ProviderManager) Register(provider ChannelProvider) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	channel := provider.GetChannel()
	pm.providers[channel] = append(pm.providers[channel], provider)

	return nil
}

// GetProvider gets the best available provider for a channel
func (pm *ProviderManager) GetProvider(channel NotificationChannel, providerName string) (ChannelProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers, ok := pm.providers[channel]
	if !ok || len(providers) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoActiveProvider, channel)
	}

	// If specific provider requested, find it
	if providerName != "" {
		for _, p := range providers {
			if p.GetProviderName() == providerName {
				// Check health
				if err := p.HealthCheck(context.Background()); err != nil {
					// Try to find fallback
					continue
				}
				return p, nil
			}
		}
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerName)
	}

	// Return first healthy provider
	for _, p := range providers {
		if err := p.HealthCheck(context.Background()); err == nil {
			return p, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrProviderNotHealthy, channel)
}

// GetProviders gets all providers for a channel
func (pm *ProviderManager) GetProviders(channel NotificationChannel) ([]ChannelProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers, ok := pm.providers[channel]
	if !ok || len(providers) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoActiveProvider, channel)
	}

	return providers, nil
}

// ListChannels lists all supported channels
func (pm *ProviderManager) ListChannels() []NotificationChannel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	channels := make([]NotificationChannel, 0, len(pm.providers))
	for channel := range pm.providers {
		channels = append(channels, channel)
	}

	return channels
}

// LoadConfigs loads provider configurations from database
func (pm *ProviderManager) LoadConfigs(ctx context.Context) error {
	var providers []NotificationProvider
	if err := pm.db.WithContext(ctx).Where("is_active = ?", true).Find(&providers).Error; err != nil {
		return fmt.Errorf("failed to load provider configs: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range providers {
		pm.configs[p.Name] = &p
	}

	return nil
}

// GetConfig gets provider configuration by name
func (pm *ProviderManager) GetConfig(name string) (*NotificationProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	config, ok := pm.configs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	return config, nil
}

// SaveConfig saves or updates provider configuration
func (pm *ProviderManager) SaveConfig(ctx context.Context, config *NotificationProvider) error {
	if config.ID == uuid.Nil {
		// Create new
		if err := pm.db.WithContext(ctx).Create(config).Error; err != nil {
			return fmt.Errorf("failed to create provider config: %w", err)
		}
	} else {
		// Update existing
		if err := pm.db.WithContext(ctx).Save(config).Error; err != nil {
			return fmt.Errorf("failed to update provider config: %w", err)
		}
	}

	pm.mu.Lock()
	pm.configs[config.Name] = config
	pm.mu.Unlock()

	return nil
}

// DeleteConfig deletes a provider configuration
func (pm *ProviderManager) DeleteConfig(ctx context.Context, name string) error {
	var config NotificationProvider
	if err := pm.db.WithContext(ctx).Where("name = ?", name).First(&config).Error; err != nil {
		return fmt.Errorf("failed to find provider config: %w", err)
	}

	if err := pm.db.WithContext(ctx).Delete(&config).Error; err != nil {
		return fmt.Errorf("failed to delete provider config: %w", err)
	}

	pm.mu.Lock()
	delete(pm.configs, name)
	pm.mu.Unlock()

	return nil
}

// UpdateHealth updates provider health status
func (pm *ProviderManager) UpdateHealth(ctx context.Context, name, status, errorMsg string) error {
	pm.mu.Lock()
	config, ok := pm.configs[name]
	pm.mu.Unlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	config.HealthStatus = status
	config.LastError = errorMsg

	if err := pm.db.WithContext(ctx).Model(config).Updates(map[string]interface{}{
		"health_status": status,
		"last_error":    errorMsg,
		"last_health_check": gorm.Expr("NOW()"),
	}).Error; err != nil {
		return fmt.Errorf("failed to update provider health: %w", err)
	}

	return nil
}

// CheckRateLimit checks if provider has exceeded rate limit
func (pm *ProviderManager) CheckRateLimit(ctx context.Context, name string) (bool, error) {
	pm.mu.RLock()
	config, ok := pm.configs[name]
	pm.mu.RUnlock()

	if !ok {
		return false, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	if config.RateLimit == 0 {
		return true, nil // No rate limit
	}

	// Check if usage needs reset
	// In a real implementation, you'd check if usage_reset_at has passed
	// and reset current_usage accordingly

	if config.CurrentUsage >= config.RateLimit {
		return false, nil
	}

	return true, nil
}

// IncrementUsage increments provider usage count
func (pm *ProviderManager) IncrementUsage(ctx context.Context, name string) error {
	pm.mu.Lock()
	config, ok := pm.configs[name]
	pm.mu.Unlock()

	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	if err := pm.db.WithContext(ctx).Model(config).UpdateColumn("current_usage", gorm.Expr("current_usage + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	pm.mu.Lock()
	config.CurrentUsage++
	pm.mu.Unlock()

	return nil
}

// NotificationConfig represents the main notification system configuration
type NotificationConfig struct {
	// Feature flags
	EnableEmail    bool `json:"enable_email"`
	EnableSMS      bool `json:"enable_sms"`
	EnableFirebase bool `json:"enable_firebase"`
	EnablePush     bool `json:"enable_push"`
	EnableWebhook  bool `json:"enable_webhook"`
	EnableInApp    bool `json:"enable_in_app"`

	// Default settings
	DefaultRetries  int    `json:"default_retries"`
	RetryDelay      int    `json:"retry_delay_seconds"`
	MaxRetryDelay   int    `json:"max_retry_delay_seconds"`
	ProcessInterval int    `json:"process_interval_seconds"`
	BatchSize       int    `json:"batch_size"`

	// Queue settings
	EnableQueue     bool   `json:"enable_queue"`
	QueueType       string `json:"queue_type"` // "memory", "redis", "kafka", etc.
	QueueMaxSize    int    `json:"queue_max_size"`

	// Audit settings
	EnableAudit          bool `json:"enable_audit"`
	AuditRetentionDays   int  `json:"audit_retention_days"`
	EnableDetailedAudit  bool `json:"enable_detailed_audit"`

	// Rate limiting
	GlobalRateLimit int `json:"global_rate_limit_per_minute"`

	// Health check
	HealthCheckInterval int `json:"health_check_interval_seconds"`
}

// DefaultConfig returns default notification configuration
func DefaultConfig() *NotificationConfig {
	return &NotificationConfig{
		EnableEmail:          true,
		EnableSMS:            true,
		EnableFirebase:       true,
		EnablePush:           true,
		EnableWebhook:        true,
		EnableInApp:          true,
		DefaultRetries:       3,
		RetryDelay:           60,
		MaxRetryDelay:        3600,
		ProcessInterval:      60,
		BatchSize:            100,
		EnableQueue:          true,
		QueueType:            "memory",
		QueueMaxSize:         10000,
		EnableAudit:          true,
		AuditRetentionDays:   90,
		EnableDetailedAudit:  true,
		GlobalRateLimit:      1000,
		HealthCheckInterval:  300,
	}
}
