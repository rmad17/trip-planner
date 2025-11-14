package channels

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"triplanner/notifications"
)

func TestEmailProvider_Validate(t *testing.T) {
	config := EmailConfig{
		FromEmail:    "noreply@tripplanner.com",
		FromName:     "Trip Planner",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user",
		SMTPPassword: "pass",
	}

	provider := NewEmailProvider(EmailProviderSMTP, config)

	tests := []struct {
		name         string
		notification *notifications.Notification
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid email notification",
			notification: &notifications.Notification{
				RecipientEmail: "user@example.com",
				Subject:        "Test Subject",
				Content:        "Test Content",
			},
			wantErr: false,
		},
		{
			name: "missing recipient email",
			notification: &notifications.Notification{
				Subject: "Test Subject",
				Content: "Test Content",
			},
			wantErr:     true,
			errContains: "recipient email",
		},
		{
			name: "missing subject",
			notification: &notifications.Notification{
				RecipientEmail: "user@example.com",
				Content:        "Test Content",
			},
			wantErr:     true,
			errContains: "subject",
		},
		{
			name: "missing content",
			notification: &notifications.Notification{
				RecipientEmail: "user@example.com",
				Subject:        "Test Subject",
			},
			wantErr:     true,
			errContains: "content",
		},
		{
			name: "valid with HTML content only",
			notification: &notifications.Notification{
				RecipientEmail: "user@example.com",
				Subject:        "Test Subject",
				ContentHTML:    "<p>Test Content</p>",
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

func TestEmailProvider_GetChannel(t *testing.T) {
	config := EmailConfig{
		FromEmail: "noreply@tripplanner.com",
	}

	provider := NewEmailProvider(EmailProviderSMTP, config)
	assert.Equal(t, notifications.ChannelEmail, provider.GetChannel())
}

func TestEmailProvider_GetProviderName(t *testing.T) {
	tests := []struct {
		name         string
		providerType EmailProviderType
		expected     string
	}{
		{"SMTP provider", EmailProviderSMTP, "smtp"},
		{"SendGrid provider", EmailProviderSendGrid, "sendgrid"},
		{"SES provider", EmailProviderSES, "ses"},
		{"Mailgun provider", EmailProviderMailgun, "mailgun"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EmailConfig{FromEmail: "test@example.com"}
			provider := NewEmailProvider(tt.providerType, config)
			assert.Equal(t, tt.expected, provider.GetProviderName())
		})
	}
}

func TestEmailProvider_GetCapabilities(t *testing.T) {
	config := EmailConfig{FromEmail: "test@example.com"}
	provider := NewEmailProvider(EmailProviderSMTP, config)

	caps := provider.GetCapabilities()

	assert.True(t, caps.SupportsBatch)
	assert.True(t, caps.SupportsTracking)
	assert.True(t, caps.SupportsTemplates)
	assert.True(t, caps.SupportsAttachments)
	assert.True(t, caps.SupportsRichContent)
	assert.Greater(t, caps.MaxBatchSize, 0)
	assert.Greater(t, caps.RateLimitPerMinute, 0)
}

func TestEmailProvider_HealthCheck(t *testing.T) {
	config := EmailConfig{
		FromEmail:    "test@example.com",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user",
		SMTPPassword: "pass",
	}

	provider := NewEmailProvider(EmailProviderSMTP, config)
	ctx := context.Background()

	// Health check will fail because we're not connecting to a real SMTP server
	// but we can test that it doesn't panic
	err := provider.HealthCheck(ctx)
	assert.Error(t, err) // Expected to fail with no real server
}

func TestEmailProvider_SendBatch(t *testing.T) {
	config := EmailConfig{
		FromEmail: "noreply@example.com",
		FromName:  "Test",
	}

	provider := NewEmailProvider(EmailProviderSMTP, config)
	ctx := context.Background()

	notifications := []*notifications.Notification{
		{
			RecipientEmail: "user1@example.com",
			Subject:        "Test 1",
			Content:        "Content 1",
		},
		{
			RecipientEmail: "user2@example.com",
			Subject:        "Test 2",
			Content:        "Content 2",
		},
	}

	results, err := provider.SendBatch(ctx, notifications)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
