package channels

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"triplanner/notifications"
)

func TestWebhookProvider_Validate(t *testing.T) {
	config := WebhookConfig{}
	provider := NewWebhookProvider(config)

	tests := []struct {
		name         string
		notification *notifications.Notification
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid webhook notification",
			notification: &notifications.Notification{
				RecipientWebhook: "https://example.com/webhook",
				Content:          "Test notification",
			},
			wantErr: false,
		},
		{
			name: "missing webhook URL",
			notification: &notifications.Notification{
				Content: "Test notification",
			},
			wantErr:     true,
			errContains: "webhook URL",
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

func TestWebhookProvider_GetChannel(t *testing.T) {
	config := WebhookConfig{}
	provider := NewWebhookProvider(config)
	assert.Equal(t, notifications.ChannelWebhook, provider.GetChannel())
}

func TestWebhookProvider_GetProviderName(t *testing.T) {
	config := WebhookConfig{}
	provider := NewWebhookProvider(config)
	assert.Equal(t, "webhook", provider.GetProviderName())
}

func TestWebhookProvider_GetCapabilities(t *testing.T) {
	config := WebhookConfig{}
	provider := NewWebhookProvider(config)

	caps := provider.GetCapabilities()

	assert.True(t, caps.SupportsBatch)
	assert.False(t, caps.SupportsScheduling)
	assert.False(t, caps.SupportsTracking)
	assert.True(t, caps.SupportsTemplates)
	assert.True(t, caps.SupportsAttachments)
	assert.True(t, caps.SupportsRichContent)
}

func TestWebhookProvider_HealthCheck(t *testing.T) {
	config := WebhookConfig{}
	provider := NewWebhookProvider(config)

	err := provider.HealthCheck(context.Background())
	assert.NoError(t, err) // Webhook provider is always healthy
}

func TestWebhookProvider_BuildPayload(t *testing.T) {
	recipientID := uuid.New()
	senderID := uuid.New()
	entityID := uuid.New()

	id := uuid.New()
	notification := &notifications.Notification{
		Type:           notifications.TypeSystem,
		Channel:        notifications.ChannelWebhook,
		Priority:       notifications.PriorityHigh,
		RecipientID:    &recipientID,
		SenderID:       &senderID,
		Subject:        "Test Subject",
		Content:        "Test Content",
		EntityType:     "trip",
		EntityID:       &entityID,
		Metadata: notifications.JSONB{
			"key": "value",
		},
		ChannelData: notifications.JSONB{
			"custom": "data",
		},
		Tags: []string{"tag1", "tag2"},
	}

	// Simulate ID assignment
	notification.ID = id

	config := WebhookConfig{}
	provider := NewWebhookProvider(config)

	payload := provider.buildPayload(notification)

	assert.Equal(t, notification.ID.String(), payload["id"])
	assert.Equal(t, string(notifications.TypeSystem), payload["type"])
	assert.Equal(t, string(notifications.ChannelWebhook), payload["channel"])
	assert.Equal(t, string(notifications.PriorityHigh), payload["priority"])
	assert.Equal(t, "Test Subject", payload["subject"])
	assert.Equal(t, "Test Content", payload["content"])
	assert.Equal(t, recipientID.String(), payload["recipient_id"])
	assert.Equal(t, senderID.String(), payload["sender_id"])
	assert.Equal(t, "trip", payload["entity_type"])
	assert.Equal(t, entityID.String(), payload["entity_id"])

	metadata, ok := payload["metadata"].(notifications.JSONB)
	assert.True(t, ok, "metadata should be present in payload")
	assert.Equal(t, "value", metadata["key"])

	channelData, ok := payload["channel_data"].(notifications.JSONB)
	assert.True(t, ok, "channel_data should be present in payload")
	assert.Equal(t, "data", channelData["custom"])

	tags, ok := payload["tags"].([]string)
	assert.True(t, ok)
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
}

func TestWebhookProvider_Send(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "TripPlanner-Notification-Service/1.0", r.Header.Get("User-Agent"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	config := WebhookConfig{
		Timeout: 5 * time.Second,
	}

	provider := NewWebhookProvider(config)

	notification := &notifications.Notification{
		// ID will be generated
		RecipientWebhook: server.URL,
		Subject:          "Test",
		Content:          "Test webhook",
	}

	result, err := provider.Send(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, notifications.StatusDelivered, result.Status)
}

func TestWebhookProvider_SendWithAuth(t *testing.T) {
	tests := []struct {
		name       string
		authType   string
		authValue  string
		authHeader string
		validate   func(t *testing.T, r *http.Request)
	}{
		{
			name:      "bearer authentication",
			authType:  "bearer",
			authValue: "test-token",
			validate: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			},
		},
		{
			name:       "API key authentication",
			authType:   "api_key",
			authValue:  "test-api-key",
			authHeader: "X-API-Key",
			validate: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.validate(t, r)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			config := WebhookConfig{
				AuthType:   tt.authType,
				AuthValue:  tt.authValue,
				AuthHeader: tt.authHeader,
			}

			provider := NewWebhookProvider(config)

			notification := &notifications.Notification{
				RecipientWebhook: server.URL,
				Content:          "Test",
			}

			_, err := provider.Send(context.Background(), notification)
			require.NoError(t, err)
		})
	}
}

func TestWebhookProvider_SendWithSignature(t *testing.T) {
	secret := "test-secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("X-Webhook-Signature")
		assert.NotEmpty(t, signature)

		// Signature should be a valid hex string
		_, err := hex.DecodeString(signature)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		SigningSecret: secret,
	}

	provider := NewWebhookProvider(config)

	notification := &notifications.Notification{
		RecipientWebhook: server.URL,
		Content:          "Test",
	}

	result, err := provider.Send(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestWebhookProvider_SendWithRetry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := WebhookConfig{
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	}

	provider := NewWebhookProvider(config)

	notification := &notifications.Notification{
		RecipientWebhook: server.URL,
		Content:          "Test",
	}

	result, err := provider.Send(context.Background(), notification)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, attempts) // Should succeed on second attempt
}

func TestVerifySignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"test": "data"}`)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		expected  bool
	}{
		{
			name:      "valid signature",
			payload:   payload,
			signature: expectedSignature,
			secret:    secret,
			expected:  true,
		},
		{
			name:      "invalid signature",
			payload:   payload,
			signature: "invalid",
			secret:    secret,
			expected:  false,
		},
		{
			name:      "wrong secret",
			payload:   payload,
			signature: expectedSignature,
			secret:    "wrong-secret",
			expected:  false,
		},
		{
			name:      "different payload",
			payload:   []byte(`{"different": "data"}`),
			signature: expectedSignature,
			secret:    secret,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifySignature(tt.payload, tt.signature, tt.secret)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebhookProvider_SendBatch(t *testing.T) {
	receivedCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{}
	provider := NewWebhookProvider(config)
	ctx := context.Background()

	notifs := []*notifications.Notification{
		{
			RecipientWebhook: server.URL,
			Content:          "Message 1",
		},
		{
			RecipientWebhook: server.URL,
			Content:          "Message 2",
		},
	}

	results, err := provider.SendBatch(ctx, notifs)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 2, receivedCount)

	for _, result := range results {
		assert.True(t, result.Success)
		assert.Equal(t, notifications.StatusDelivered, result.Status)
	}
}
