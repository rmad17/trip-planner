package channels

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"triplanner/notifications"
)

// WebhookProvider implements the ChannelProvider interface for webhooks
type WebhookProvider struct {
	config WebhookConfig
	client *http.Client
}

// WebhookConfig holds webhook provider configuration
type WebhookConfig struct {
	// Signing secret for webhook verification
	SigningSecret string

	// HTTP client configuration
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration

	// Headers to include in webhook requests
	CustomHeaders map[string]string

	// Authentication
	AuthType   string // "none", "basic", "bearer", "api_key"
	AuthValue  string // Username:Password for basic, token for bearer, key for api_key
	AuthHeader string // Header name for api_key auth
}

// NewWebhookProvider creates a new webhook provider
func NewWebhookProvider(config WebhookConfig) *WebhookProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &WebhookProvider{
		config: config,
		client: client,
	}
}

// Send sends a webhook notification
func (wp *WebhookProvider) Send(ctx context.Context, notification *notifications.Notification) (*notifications.SendResult, error) {
	result := &notifications.SendResult{
		Status: notifications.StatusSending,
		SentAt: time.Now(),
	}

	// Validate notification
	if err := wp.Validate(notification); err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	// Build webhook payload
	payload := wp.buildPayload(notification)

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = fmt.Sprintf("failed to marshal payload: %v", err)
		result.Status = notifications.StatusFailed
		return result, err
	}

	// Send webhook
	err = wp.sendWebhook(ctx, notification.RecipientWebhook, jsonData)
	if err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	result.Success = true
	result.MessageID = notification.ID.String()
	result.ExternalID = notification.ID.String()
	result.Status = notifications.StatusDelivered
	result.ProviderResponse = payload

	return result, nil
}

// SendBatch sends multiple webhook notifications in a batch
func (wp *WebhookProvider) SendBatch(ctx context.Context, notifs []*notifications.Notification) ([]*notifications.SendResult, error) {
	results := make([]*notifications.SendResult, len(notifs))

	for i, notif := range notifs {
		result, err := wp.Send(ctx, notif)
		if err != nil {
			results[i] = result
			continue
		}
		results[i] = result
	}

	return results, nil
}

// Validate validates the notification data for webhook
func (wp *WebhookProvider) Validate(notification *notifications.Notification) error {
	if notification.RecipientWebhook == "" {
		return errors.New("recipient webhook URL is required")
	}

	return nil
}

// GetChannel returns the channel type
func (wp *WebhookProvider) GetChannel() notifications.NotificationChannel {
	return notifications.ChannelWebhook
}

// GetProviderName returns the provider name
func (wp *WebhookProvider) GetProviderName() string {
	return "webhook"
}

// HealthCheck checks if the provider is healthy
func (wp *WebhookProvider) HealthCheck(ctx context.Context) error {
	// Webhook provider is always healthy if configured
	return nil
}

// GetCapabilities returns the provider capabilities
func (wp *WebhookProvider) GetCapabilities() notifications.ProviderCapabilities {
	return notifications.ProviderCapabilities{
		SupportsBatch:       true,
		SupportsScheduling:  false,
		SupportsTracking:    false,
		SupportsTemplates:   true,
		SupportsAttachments: true,
		SupportsRichContent: true,
		MaxBatchSize:        100,
		RateLimitPerMinute:  60,
	}
}

// Helper methods

func (wp *WebhookProvider) buildPayload(notification *notifications.Notification) map[string]interface{} {
	payload := map[string]interface{}{
		"id":         notification.ID.String(),
		"type":       string(notification.Type),
		"channel":    string(notification.Channel),
		"priority":   string(notification.Priority),
		"subject":    notification.Subject,
		"content":    notification.Content,
		"created_at": notification.CreatedAt,
		"sent_at":    time.Now(),
	}

	// Add recipient information
	if notification.RecipientID != nil {
		payload["recipient_id"] = notification.RecipientID.String()
	}

	// Add sender information
	if notification.SenderID != nil {
		payload["sender_id"] = notification.SenderID.String()
	}

	// Add entity information
	if notification.EntityType != "" {
		payload["entity_type"] = notification.EntityType
	}
	if notification.EntityID != nil {
		payload["entity_id"] = notification.EntityID.String()
	}

	// Add metadata
	if len(notification.Metadata) > 0 {
		payload["metadata"] = notification.Metadata
	}

	// Add channel data
	if len(notification.ChannelData) > 0 {
		payload["channel_data"] = notification.ChannelData
	}

	// Add tags
	if len(notification.Tags) > 0 {
		payload["tags"] = notification.Tags
	}

	return payload
}

func (wp *WebhookProvider) sendWebhook(ctx context.Context, url string, payload []byte) error {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TripPlanner-Notification-Service/1.0")

	// Add custom headers
	for key, value := range wp.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Add authentication
	switch wp.config.AuthType {
	case "basic":
		// AuthValue should be "username:password"
		req.SetBasicAuth(wp.config.AuthValue, "")
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+wp.config.AuthValue)
	case "api_key":
		if wp.config.AuthHeader != "" {
			req.Header.Set(wp.config.AuthHeader, wp.config.AuthValue)
		}
	}

	// Add signature if signing secret is configured
	if wp.config.SigningSecret != "" {
		signature := wp.generateSignature(payload)
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	}

	// Send request with retries
	var lastErr error
	maxRetries := wp.config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := wp.config.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		resp, err := wp.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d): %w", attempt+1, err)
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d (attempt %d): %s", resp.StatusCode, attempt+1, string(body))

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return lastErr
}

func (wp *WebhookProvider) generateSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(wp.config.SigningSecret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies a webhook signature (for incoming webhooks)
func VerifySignature(payload []byte, signature string, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
