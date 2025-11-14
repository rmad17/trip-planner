package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&Notification{},
		&NotificationTemplate{},
		&NotificationAudit{},
		&NotificationPreference{},
		&NotificationBatch{},
		&NotificationProvider{},
	)
	require.NoError(t, err)

	return db
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultConfig()

	service := NewService(db, config)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.providerManager)
	assert.NotNil(t, service.templateService)
	assert.NotNil(t, service.auditService)
	assert.NotNil(t, service.prefService)
}

func TestCreateNotificationFromRequest(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultConfig()
	service := NewService(db, config)

	recipientID := uuid.New()
	req := &SendRequest{
		Type:           TypeTransactional,
		Channel:        ChannelEmail,
		Priority:       PriorityNormal,
		RecipientID:    &recipientID,
		RecipientEmail: "test@example.com",
		Subject:        "Test Subject",
		Content:        "Test Content",
		ContentHTML:    "<p>Test Content</p>",
	}

	notification := service.createNotificationFromRequest(req)

	assert.Equal(t, TypeTransactional, notification.Type)
	assert.Equal(t, ChannelEmail, notification.Channel)
	assert.Equal(t, PriorityNormal, notification.Priority)
	assert.Equal(t, StatusPending, notification.Status)
	assert.Equal(t, "test@example.com", notification.RecipientEmail)
	assert.Equal(t, "Test Subject", notification.Subject)
	assert.Equal(t, "Test Content", notification.Content)
	assert.Equal(t, config.DefaultRetries, notification.MaxRetries)
}

func TestValidateSendRequest(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultConfig()
	service := NewService(db, config)

	tests := []struct {
		name    string
		req     *SendRequest
		wantErr bool
	}{
		{
			name: "valid email request",
			req: &SendRequest{
				Channel:        ChannelEmail,
				RecipientEmail: "test@example.com",
				Content:        "Test content",
			},
			wantErr: false,
		},
		{
			name: "missing email for email channel",
			req: &SendRequest{
				Channel: ChannelEmail,
				Content: "Test content",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			req: &SendRequest{
				Channel:        ChannelEmail,
				RecipientEmail: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "valid SMS request",
			req: &SendRequest{
				Channel:        ChannelSMS,
				RecipientPhone: "+1234567890",
				Content:        "Test content",
			},
			wantErr: false,
		},
		{
			name: "missing phone for SMS channel",
			req: &SendRequest{
				Channel: ChannelSMS,
				Content: "Test content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSendRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateService(t *testing.T) {
	db := setupTestDB(t)
	templateService := NewTemplateService(db)
	ctx := context.Background()

	// Create template
	template := &NotificationTemplate{
		Name:        "test_template",
		Type:        TypeTransactional,
		Channel:     ChannelEmail,
		Subject:     "Hello {{.Name}}",
		Content:     "Welcome, {{.Name}}! Your code is {{.Code}}.",
		ContentHTML: "<p>Welcome, {{.Name}}! Your code is {{.Code}}.</p>",
		IsActive:    true,
	}

	err := templateService.Create(ctx, template)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, template.ID)

	// Get template
	retrieved, err := templateService.Get(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.Name, retrieved.Name)

	// Get by name
	byName, err := templateService.GetByName(ctx, "test_template")
	require.NoError(t, err)
	assert.Equal(t, template.ID, byName.ID)

	// Render template
	data := map[string]interface{}{
		"Name": "John Doe",
		"Code": "ABC123",
	}

	content, contentHTML, err := templateService.Render(ctx, template, data)
	require.NoError(t, err)
	assert.Contains(t, content, "John Doe")
	assert.Contains(t, content, "ABC123")
	assert.Contains(t, contentHTML, "John Doe")

	// List templates
	templates, err := templateService.List(ctx, TemplateFilters{})
	require.NoError(t, err)
	assert.Len(t, templates, 1)

	// Update template
	template.Subject = "Updated Subject"
	err = templateService.Update(ctx, template)
	require.NoError(t, err)
	assert.Equal(t, 2, template.Version)

	// Delete template
	err = templateService.Delete(ctx, template.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = templateService.Get(ctx, template.ID)
	assert.Error(t, err)
}

func TestAuditService(t *testing.T) {
	db := setupTestDB(t)
	auditService := NewAuditService(db)
	ctx := context.Background()

	notificationID := uuid.New()

	// Log audit entry
	audit := &NotificationAudit{
		NotificationID: notificationID,
		Status:         StatusPending,
		Event:          "created",
		Message:        "Notification created",
		Timestamp:      time.Now(),
		ActorType:      "system",
	}

	err := auditService.Log(ctx, audit)
	require.NoError(t, err)

	// Get audit trail
	trail, err := auditService.GetTrail(ctx, notificationID)
	require.NoError(t, err)
	assert.Len(t, trail, 1)
	assert.Equal(t, "created", trail[0].Event)

	// Log another entry
	audit2 := &NotificationAudit{
		NotificationID: notificationID,
		Status:         StatusSent,
		Event:          "sent",
		Message:        "Notification sent",
		Timestamp:      time.Now(),
		ActorType:      "system",
	}

	err = auditService.Log(ctx, audit2)
	require.NoError(t, err)

	// Get updated trail
	trail, err = auditService.GetTrail(ctx, notificationID)
	require.NoError(t, err)
	assert.Len(t, trail, 2)
}

func TestPreferenceService(t *testing.T) {
	db := setupTestDB(t)
	prefService := NewPreferenceService(db)
	ctx := context.Background()

	userID := uuid.New()

	// Get default preference (doesn't exist yet)
	pref, err := prefService.Get(ctx, userID, ChannelEmail, TypeMarketing)
	require.NoError(t, err)
	assert.True(t, pref.IsEnabled) // Default is enabled

	// Set preference
	pref.IsEnabled = false
	err = prefService.Set(ctx, pref)
	require.NoError(t, err)

	// Get updated preference
	updated, err := prefService.Get(ctx, userID, ChannelEmail, TypeMarketing)
	require.NoError(t, err)
	assert.False(t, updated.IsEnabled)

	// Check CanSend
	canSend, reason, err := prefService.CanSend(ctx, userID, ChannelEmail, TypeMarketing)
	require.NoError(t, err)
	assert.False(t, canSend)
	assert.Contains(t, reason, "disabled")

	// Enable preference
	pref.IsEnabled = true
	err = prefService.Set(ctx, pref)
	require.NoError(t, err)

	// Check CanSend again
	canSend, _, err = prefService.CanSend(ctx, userID, ChannelEmail, TypeMarketing)
	require.NoError(t, err)
	assert.True(t, canSend)

	// List preferences
	prefs, err := prefService.List(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, prefs, 1)
}

func TestProviderManager(t *testing.T) {
	db := setupTestDB(t)
	pm := NewProviderManager(db)

	// Register a mock provider
	mockProvider := &MockChannelProvider{
		channel:      ChannelEmail,
		providerName: "mock",
	}

	err := pm.Register(mockProvider)
	require.NoError(t, err)

	// Get provider
	provider, err := pm.GetProvider(ChannelEmail, "")
	require.NoError(t, err)
	assert.Equal(t, ChannelEmail, provider.GetChannel())

	// Get specific provider
	provider, err = pm.GetProvider(ChannelEmail, "mock")
	require.NoError(t, err)
	assert.Equal(t, "mock", provider.GetProviderName())

	// List channels
	channels := pm.ListChannels()
	assert.Contains(t, channels, ChannelEmail)
}

// MockChannelProvider is a mock implementation for testing
type MockChannelProvider struct {
	channel      NotificationChannel
	providerName string
}

func (m *MockChannelProvider) Send(ctx context.Context, notification *Notification) (*SendResult, error) {
	return &SendResult{
		Success:    true,
		Status:     StatusSent,
		MessageID:  uuid.New().String(),
		ExternalID: uuid.New().String(),
		SentAt:     time.Now(),
	}, nil
}

func (m *MockChannelProvider) SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error) {
	results := make([]*SendResult, len(notifications))
	for i := range notifications {
		result, _ := m.Send(ctx, notifications[i])
		results[i] = result
	}
	return results, nil
}

func (m *MockChannelProvider) Validate(notification *Notification) error {
	return nil
}

func (m *MockChannelProvider) GetChannel() NotificationChannel {
	return m.channel
}

func (m *MockChannelProvider) GetProviderName() string {
	return m.providerName
}

func (m *MockChannelProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockChannelProvider) GetCapabilities() ProviderCapabilities {
	return ProviderCapabilities{
		SupportsBatch:      true,
		SupportsScheduling: false,
		SupportsTracking:   true,
	}
}
