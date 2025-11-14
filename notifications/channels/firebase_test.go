package channels

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"triplanner/notifications"
)

func TestFirebaseProvider_Validate(t *testing.T) {
	config := FirebaseConfig{
		ProjectID:          "test-project",
		ServiceAccountJSON: []byte("{}"),
	}

	provider := NewFirebaseProvider(config)

	tests := []struct {
		name         string
		notification *notifications.Notification
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid Firebase notification",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Content:           "Test notification",
				Subject:           "Test Title",
			},
			wantErr: false,
		},
		{
			name: "missing device token",
			notification: &notifications.Notification{
				Content: "Test notification",
			},
			wantErr:     true,
			errContains: "device token",
		},
		{
			name: "missing content",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
			},
			wantErr:     true,
			errContains: "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Validate(tt.notification)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFirebaseProvider_GetChannel(t *testing.T) {
	config := FirebaseConfig{
		ProjectID: "test-project",
	}

	provider := NewFirebaseProvider(config)
	assert.Equal(t, notifications.ChannelFirebase, provider.GetChannel())
}

func TestFirebaseProvider_GetProviderName(t *testing.T) {
	config := FirebaseConfig{
		ProjectID: "test-project",
	}

	provider := NewFirebaseProvider(config)
	assert.Equal(t, "firebase", provider.GetProviderName())
}

func TestFirebaseProvider_GetCapabilities(t *testing.T) {
	config := FirebaseConfig{}
	provider := NewFirebaseProvider(config)

	caps := provider.GetCapabilities()

	assert.True(t, caps.SupportsBatch)
	assert.True(t, caps.SupportsTracking)
	assert.True(t, caps.SupportsTemplates)
	assert.False(t, caps.SupportsAttachments)
	assert.True(t, caps.SupportsRichContent)
	assert.Equal(t, 500, caps.MaxBatchSize)
	assert.Equal(t, 600000, caps.RateLimitPerMinute)
	assert.Contains(t, caps.SupportedPriorities, "high")
	assert.Contains(t, caps.SupportedPriorities, "normal")
}

func TestFirebaseProvider_HealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		config  FirebaseConfig
		wantErr bool
	}{
		{
			name:    "missing project ID",
			config:  FirebaseConfig{},
			wantErr: true,
		},
		{
			name: "missing credentials",
			config: FirebaseConfig{
				ProjectID: "test-project",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: FirebaseConfig{
				ProjectID:          "test-project",
				ServiceAccountJSON: []byte("{}"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewFirebaseProvider(tt.config)
			err := provider.HealthCheck(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildFCMMessage(t *testing.T) {
	entityID := uuid.New()

	tests := []struct {
		name         string
		notification *notifications.Notification
		validate     func(t *testing.T, msg *FirebaseMessage)
	}{
		{
			name: "basic notification",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Subject:           "Test Title",
				Content:           "Test Body",
			},
			validate: func(t *testing.T, msg *FirebaseMessage) {
				assert.Equal(t, "device-token-123", msg.Token)
				assert.Equal(t, "Test Title", msg.Notification.Title)
				assert.Equal(t, "Test Body", msg.Notification.Body)
			},
		},
		{
			name: "with entity metadata",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Subject:           "Test Title",
				Content:           "Test Body",
				EntityType:        "trip",
				EntityID:          &entityID,
			},
			validate: func(t *testing.T, msg *FirebaseMessage) {
				assert.Equal(t, "trip", msg.Data["entity_type"])
				assert.Equal(t, entityID.String(), msg.Data["entity_id"])
			},
		},
		{
			name: "high priority",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Subject:           "Urgent",
				Content:           "Important message",
				Priority:          notifications.PriorityHigh,
			},
			validate: func(t *testing.T, msg *FirebaseMessage) {
				assert.NotNil(t, msg.Android)
				assert.Equal(t, "high", msg.Android.Priority)
				assert.NotNil(t, msg.APNS)
				assert.Equal(t, "10", msg.APNS.Headers["apns-priority"])
			},
		},
		{
			name: "with channel data",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Subject:           "Test",
				Content:           "Body",
				ChannelData: map[string]interface{}{
					"custom_key": "custom_value",
				},
			},
			validate: func(t *testing.T, msg *FirebaseMessage) {
				assert.Equal(t, "custom_value", msg.Data["custom_key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := BuildFCMMessage(tt.notification)
			assert.NotNil(t, msg)
			tt.validate(t, msg)
		})
	}
}

func TestFirebaseProvider_SendBatch(t *testing.T) {
	config := FirebaseConfig{
		ProjectID:          "test-project",
		ServiceAccountJSON: []byte("{}"),
	}

	provider := NewFirebaseProvider(config)
	ctx := context.Background()

	notifications := []*notifications.Notification{
		{
			RecipientDeviceID: "device-1",
			Subject:           "Test 1",
			Content:           "Content 1",
		},
		{
			RecipientDeviceID: "device-2",
			Subject:           "Test 2",
			Content:           "Content 2",
		},
	}

	results, err := provider.SendBatch(ctx, notifications)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
