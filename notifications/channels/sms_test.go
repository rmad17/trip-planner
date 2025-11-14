package channels

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"triplanner/notifications"
)

func TestSMSProvider_Validate(t *testing.T) {
	config := SMSConfig{
		TwilioAccountSID: "test-sid",
		TwilioAuthToken:  "test-token",
		TwilioFromNumber: "+1234567890",
	}

	provider := NewSMSProvider(SMSProviderTwilio, config)

	tests := []struct {
		name         string
		notification *notifications.Notification
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid SMS notification",
			notification: &notifications.Notification{
				RecipientPhone: "+1234567890",
				Content:        "Test message",
			},
			wantErr: false,
		},
		{
			name: "missing recipient phone",
			notification: &notifications.Notification{
				Content: "Test message",
			},
			wantErr:     true,
			errContains: "phone number",
		},
		{
			name: "missing content",
			notification: &notifications.Notification{
				RecipientPhone: "+1234567890",
			},
			wantErr:     true,
			errContains: "content",
		},
		{
			name: "content too long",
			notification: &notifications.Notification{
				RecipientPhone: "+1234567890",
				Content:        string(make([]byte, 1601)), // Over limit
			},
			wantErr:     true,
			errContains: "exceeds maximum length",
		},
		{
			name: "content at max length",
			notification: &notifications.Notification{
				RecipientPhone: "+1234567890",
				Content:        string(make([]byte, 1600)),
			},
			wantErr: false,
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

func TestSMSProvider_GetChannel(t *testing.T) {
	config := SMSConfig{
		TwilioAccountSID: "test-sid",
	}

	provider := NewSMSProvider(SMSProviderTwilio, config)
	assert.Equal(t, notifications.ChannelSMS, provider.GetChannel())
}

func TestSMSProvider_GetProviderName(t *testing.T) {
	tests := []struct {
		name         string
		providerType SMSProviderType
		expected     string
	}{
		{"Twilio provider", SMSProviderTwilio, "twilio"},
		{"AWS SNS provider", SMSProviderAWSSNS, "aws_sns"},
		{"Vonage provider", SMSProviderVonage, "vonage"},
		{"Plivo provider", SMSProviderPlivo, "plivo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SMSConfig{}
			provider := NewSMSProvider(tt.providerType, config)
			assert.Equal(t, tt.expected, provider.GetProviderName())
		})
	}
}

func TestSMSProvider_GetCapabilities(t *testing.T) {
	config := SMSConfig{}
	provider := NewSMSProvider(SMSProviderTwilio, config)

	caps := provider.GetCapabilities()

	assert.True(t, caps.SupportsBatch)
	assert.True(t, caps.SupportsTracking)
	assert.True(t, caps.SupportsTemplates)
	assert.False(t, caps.SupportsAttachments)
	assert.False(t, caps.SupportsRichContent)
	assert.Greater(t, caps.MaxBatchSize, 0)
}

func TestSMSProvider_HealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		providerType SMSProviderType
		config       SMSConfig
		wantErr      bool
	}{
		{
			name:         "Twilio without credentials",
			providerType: SMSProviderTwilio,
			config:       SMSConfig{},
			wantErr:      true,
		},
		{
			name:         "Twilio with credentials",
			providerType: SMSProviderTwilio,
			config: SMSConfig{
				TwilioAccountSID: "test-sid",
				TwilioAuthToken:  "test-token",
			},
			wantErr: false,
		},
		{
			name:         "SNS without credentials",
			providerType: SMSProviderAWSSNS,
			config:       SMSConfig{},
			wantErr:      true,
		},
		{
			name:         "SNS with credentials",
			providerType: SMSProviderAWSSNS,
			config: SMSConfig{
				SNSAccessKeyID:     "test-key",
				SNSSecretAccessKey: "test-secret",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSMSProvider(tt.providerType, tt.config)
			err := provider.HealthCheck(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSProvider_SendBatch(t *testing.T) {
	config := SMSConfig{
		TwilioAccountSID: "test-sid",
		TwilioAuthToken:  "test-token",
	}

	provider := NewSMSProvider(SMSProviderTwilio, config)
	ctx := context.Background()

	notifications := []*notifications.Notification{
		{
			RecipientPhone: "+1234567890",
			Content:        "Message 1",
		},
		{
			RecipientPhone: "+0987654321",
			Content:        "Message 2",
		},
	}

	results, err := provider.SendBatch(ctx, notifications)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
