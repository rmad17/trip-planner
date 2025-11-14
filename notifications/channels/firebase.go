package channels

import (
	"context"
	"errors"
	"time"

	"triplanner/notifications"
)

// FirebaseProvider implements the ChannelProvider interface for Firebase Cloud Messaging
type FirebaseProvider struct {
	config FirebaseConfig
}

// FirebaseConfig holds Firebase FCM configuration
type FirebaseConfig struct {
	ProjectID   string
	Credentials string // JSON credentials file path or content
	// For HTTP v1 API
	ServiceAccountJSON []byte
}

// NewFirebaseProvider creates a new Firebase provider
func NewFirebaseProvider(config FirebaseConfig) *FirebaseProvider {
	return &FirebaseProvider{
		config: config,
	}
}

// Send sends a Firebase push notification
func (fp *FirebaseProvider) Send(ctx context.Context, notification *notifications.Notification) (*notifications.SendResult, error) {
	result := &notifications.SendResult{
		Status: notifications.StatusSending,
		SentAt: time.Now(),
	}

	// Validate notification
	if err := fp.Validate(notification); err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	// Send via FCM
	messageID, err := fp.sendViaFCM(ctx, notification)
	if err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	result.Success = true
	result.MessageID = messageID
	result.ExternalID = messageID
	result.Status = notifications.StatusSent

	return result, nil
}

// SendBatch sends multiple Firebase notifications in a batch
func (fp *FirebaseProvider) SendBatch(ctx context.Context, notifs []*notifications.Notification) ([]*notifications.SendResult, error) {
	// Firebase supports batch sending for up to 500 messages
	results := make([]*notifications.SendResult, len(notifs))

	// In production, use FCM's MulticastMessage for better performance
	for i, notif := range notifs {
		result, err := fp.Send(ctx, notif)
		if err != nil {
			results[i] = result
			continue
		}
		results[i] = result
	}

	return results, nil
}

// Validate validates the notification data for Firebase
func (fp *FirebaseProvider) Validate(notification *notifications.Notification) error {
	if notification.RecipientDeviceID == "" {
		return errors.New("recipient device token is required for Firebase")
	}

	if notification.Content == "" {
		return errors.New("notification content is required")
	}

	return nil
}

// GetChannel returns the channel type
func (fp *FirebaseProvider) GetChannel() notifications.NotificationChannel {
	return notifications.ChannelFirebase
}

// GetProviderName returns the provider name
func (fp *FirebaseProvider) GetProviderName() string {
	return "firebase"
}

// HealthCheck checks if the provider is healthy
func (fp *FirebaseProvider) HealthCheck(ctx context.Context) error {
	// Check if credentials are configured
	if fp.config.ProjectID == "" {
		return errors.New("Firebase project ID not configured")
	}

	if len(fp.config.ServiceAccountJSON) == 0 && fp.config.Credentials == "" {
		return errors.New("Firebase credentials not configured")
	}

	// In production, perform an actual test request to FCM
	return nil
}

// GetCapabilities returns the provider capabilities
func (fp *FirebaseProvider) GetCapabilities() notifications.ProviderCapabilities {
	return notifications.ProviderCapabilities{
		SupportsBatch:       true,
		SupportsScheduling:  false,
		SupportsTracking:    true,
		SupportsTemplates:   true,
		SupportsAttachments: false,
		SupportsRichContent: true,
		MaxBatchSize:        500,
		RateLimitPerMinute:  600000, // FCM has high limits
		SupportedPriorities: []string{"high", "normal"},
	}
}

// Provider-specific implementation

func (fp *FirebaseProvider) sendViaFCM(ctx context.Context, notification *notifications.Notification) (string, error) {
	// FCM implementation
	// In production, use Firebase SDK: firebase.google.com/go/v4

	// Placeholder implementation that outlines the structure
	// You would need to:
	// 1. Initialize Firebase app with credentials
	// 2. Get messaging client
	// 3. Build FCM message
	// 4. Send message
	// 5. Return message ID

	/*
		Example structure:

		app, err := firebase.NewApp(ctx, &firebase.Config{
			ProjectID: fp.config.ProjectID,
		}, option.WithCredentialsJSON(fp.config.ServiceAccountJSON))
		if err != nil {
			return "", fmt.Errorf("failed to initialize Firebase: %w", err)
		}

		client, err := app.Messaging(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get messaging client: %w", err)
		}

		// Build message
		message := &messaging.Message{
			Token: notification.RecipientDeviceID,
			Notification: &messaging.Notification{
				Title: notification.Subject,
				Body:  notification.Content,
			},
			Data: map[string]string{
				"entity_type": notification.EntityType,
				"entity_id":   notification.EntityID.String(),
			},
		}

		// Set priority
		if notification.Priority == notifications.PriorityHigh || notification.Priority == notifications.PriorityCritical {
			message.Android = &messaging.AndroidConfig{
				Priority: "high",
			}
			message.APNS = &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
			}
		}

		// Send message
		response, err := client.Send(ctx, message)
		if err != nil {
			return "", fmt.Errorf("failed to send FCM message: %w", err)
		}

		return response, nil
	*/

	return "", errors.New("Firebase implementation requires Firebase SDK (firebase.google.com/go/v4)")
}

// FirebaseMessage represents a structured FCM message
type FirebaseMessage struct {
	Token        string                 `json:"token"`
	Notification FirebaseNotification   `json:"notification,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	Android      *FirebaseAndroid       `json:"android,omitempty"`
	APNS         *FirebaseAPNS          `json:"apns,omitempty"`
	WebPush      *FirebaseWebPush       `json:"webpush,omitempty"`
}

// FirebaseNotification represents the notification payload
type FirebaseNotification struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	ImageURL string `json:"image,omitempty"`
}

// FirebaseAndroid represents Android-specific options
type FirebaseAndroid struct {
	Priority string                 `json:"priority,omitempty"` // "normal" or "high"
	TTL      string                 `json:"ttl,omitempty"`      // Time to live
	Data     map[string]string      `json:"data,omitempty"`
}

// FirebaseAPNS represents iOS-specific options
type FirebaseAPNS struct {
	Headers map[string]string      `json:"headers,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// FirebaseWebPush represents web push options
type FirebaseWebPush struct {
	Headers      map[string]string      `json:"headers,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	Notification map[string]interface{} `json:"notification,omitempty"`
}

// BuildFCMMessage builds an FCM message from a notification
func BuildFCMMessage(notification *notifications.Notification) *FirebaseMessage {
	msg := &FirebaseMessage{
		Token: notification.RecipientDeviceID,
		Notification: FirebaseNotification{
			Title: notification.Subject,
			Body:  notification.Content,
		},
		Data: make(map[string]string),
	}

	// Add metadata as data payload
	if notification.EntityType != "" {
		msg.Data["entity_type"] = notification.EntityType
	}

	if notification.EntityID != nil {
		msg.Data["entity_id"] = notification.EntityID.String()
	}

	// Add custom data from ChannelData
	if notification.ChannelData != nil {
		for k, v := range notification.ChannelData {
			if str, ok := v.(string); ok {
				msg.Data[k] = str
			}
		}
	}

	// Set platform-specific configs based on priority
	if notification.Priority == notifications.PriorityHigh || notification.Priority == notifications.PriorityCritical {
		msg.Android = &FirebaseAndroid{
			Priority: "high",
		}
		msg.APNS = &FirebaseAPNS{
			Headers: map[string]string{
				"apns-priority": "10",
			},
		}
	}

	return msg
}
