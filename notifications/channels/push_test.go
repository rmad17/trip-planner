package channels

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"triplanner/notifications"
)

func TestPushProvider_Validate(t *testing.T) {
	config := PushConfig{
		Platform: "ios",
	}

	provider := NewPushProvider(config)

	tests := []struct {
		name         string
		notification *notifications.Notification
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid push notification",
			notification: &notifications.Notification{
				RecipientDeviceID: "device-token-123",
				Content:           "Test notification",
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

func TestPushProvider_GetChannel(t *testing.T) {
	config := PushConfig{
		Platform: "ios",
	}

	provider := NewPushProvider(config)
	assert.Equal(t, notifications.ChannelPush, provider.GetChannel())
}

func TestPushProvider_GetProviderName(t *testing.T) {
	tests := []struct {
		platform string
		expected string
	}{
		{"ios", "push-ios"},
		{"android", "push-android"},
		{"web", "push-web"},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			config := PushConfig{Platform: tt.platform}
			provider := NewPushProvider(config)
			assert.Equal(t, tt.expected, provider.GetProviderName())
		})
	}
}

func TestPushProvider_GetCapabilities(t *testing.T) {
	config := PushConfig{Platform: "ios"}
	provider := NewPushProvider(config)

	caps := provider.GetCapabilities()

	assert.True(t, caps.SupportsBatch)
	assert.True(t, caps.SupportsTracking)
	assert.True(t, caps.SupportsTemplates)
	assert.False(t, caps.SupportsAttachments)
	assert.True(t, caps.SupportsRichContent)
	assert.Equal(t, 500, caps.MaxBatchSize)
	assert.Contains(t, caps.SupportedPriorities, "high")
	assert.Contains(t, caps.SupportedPriorities, "normal")
	assert.Contains(t, caps.SupportedPriorities, "low")
}

func TestPushProvider_HealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		config  PushConfig
		wantErr bool
	}{
		{
			name: "iOS without credentials",
			config: PushConfig{
				Platform: "ios",
			},
			wantErr: true,
		},
		{
			name: "iOS with credentials",
			config: PushConfig{
				Platform:   "ios",
				APNsKeyID:  "test-key-id",
				APNsTeamID: "test-team-id",
			},
			wantErr: false,
		},
		{
			name: "web without VAPID keys",
			config: PushConfig{
				Platform: "web",
			},
			wantErr: true,
		},
		{
			name: "web with VAPID keys",
			config: PushConfig{
				Platform:        "web",
				VAPIDPublicKey:  "test-public-key",
				VAPIDPrivateKey: "test-private-key",
			},
			wantErr: false,
		},
		{
			name: "Android platform",
			config: PushConfig{
				Platform: "android",
			},
			wantErr: false, // Android uses Firebase, no specific check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewPushProvider(tt.config)
			err := provider.HealthCheck(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildPushPayload(t *testing.T) {
	entityID := uuid.New()

	tests := []struct {
		name         string
		notification *notifications.Notification
		validate     func(t *testing.T, payload *PushPayload)
	}{
		{
			name: "basic push notification",
			notification: &notifications.Notification{
				Subject: "Test Title",
				Content: "Test Body",
			},
			validate: func(t *testing.T, payload *PushPayload) {
				assert.Equal(t, "Test Title", payload.Title)
				assert.Equal(t, "Test Body", payload.Body)
				assert.Equal(t, "default", payload.Sound)
			},
		},
		{
			name: "with entity metadata",
			notification: &notifications.Notification{
				Subject:    "Test",
				Content:    "Body",
				EntityType: "trip",
				EntityID:   &entityID,
			},
			validate: func(t *testing.T, payload *PushPayload) {
				assert.Equal(t, "trip", payload.Data["entity_type"])
				assert.Equal(t, entityID.String(), payload.Data["entity_id"])
			},
		},
		{
			name: "with channel data",
			notification: &notifications.Notification{
				Subject: "Test",
				Content: "Body",
				ChannelData: map[string]interface{}{
					"icon":  "/icon.png",
					"badge": "/badge.png",
					"image": "/image.jpg",
					"url":   "https://example.com",
				},
			},
			validate: func(t *testing.T, payload *PushPayload) {
				assert.Equal(t, "/icon.png", payload.Icon)
				assert.Equal(t, "/badge.png", payload.Badge)
				assert.Equal(t, "/image.jpg", payload.Image)
				assert.Equal(t, "https://example.com", payload.URL)
			},
		},
		{
			name: "with custom data",
			notification: &notifications.Notification{
				Subject: "Test",
				Content: "Body",
				ChannelData: map[string]interface{}{
					"custom_field": "custom_value",
					"number":       123,
				},
			},
			validate: func(t *testing.T, payload *PushPayload) {
				assert.Equal(t, "custom_value", payload.Data["custom_field"])
				assert.Equal(t, 123, payload.Data["number"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := BuildPushPayload(tt.notification)
			assert.NotNil(t, payload)
			assert.NotNil(t, payload.Data)
			tt.validate(t, payload)
		})
	}
}

func TestPushProvider_SendBatch(t *testing.T) {
	config := PushConfig{
		Platform:   "ios",
		APNsKeyID:  "test-key",
		APNsTeamID: "test-team",
	}

	provider := NewPushProvider(config)
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

func TestPushPayload_Actions(t *testing.T) {
	notification := &notifications.Notification{
		Subject: "Test",
		Content: "Body",
	}

	payload := BuildPushPayload(notification)

	// Initially no actions
	assert.Nil(t, payload.Actions)
	assert.Empty(t, payload.Actions)
}
