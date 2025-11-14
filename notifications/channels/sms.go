package channels

import (
	"context"
	"errors"
	"fmt"
	"time"

	"triplanner/notifications"
)

// SMSProviderType represents different SMS providers
type SMSProviderType string

const (
	SMSProviderTwilio  SMSProviderType = "twilio"
	SMSProviderAWSSNS  SMSProviderType = "aws_sns"
	SMSProviderVonage  SMSProviderType = "vonage"
	SMSProviderPlivo   SMSProviderType = "plivo"
)

// SMSProvider implements the ChannelProvider interface for SMS
type SMSProvider struct {
	providerType SMSProviderType
	config       SMSConfig
}

// SMSConfig holds SMS provider configuration
type SMSConfig struct {
	// Common
	SenderPhone string

	// Twilio specific
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// AWS SNS specific
	SNSRegion          string
	SNSAccessKeyID     string
	SNSSecretAccessKey string

	// Vonage (Nexmo) specific
	VonageAPIKey    string
	VonageAPISecret string
	VonageFromName  string

	// Plivo specific
	PlivoAuthID    string
	PlivoAuthToken string
	PlivoFromNumber string
}

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(providerType SMSProviderType, config SMSConfig) *SMSProvider {
	return &SMSProvider{
		providerType: providerType,
		config:       config,
	}
}

// Send sends an SMS notification
func (sp *SMSProvider) Send(ctx context.Context, notification *notifications.Notification) (*notifications.SendResult, error) {
	result := &notifications.SendResult{
		Status: notifications.StatusSending,
		SentAt: time.Now(),
	}

	// Validate notification
	if err := sp.Validate(notification); err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	// Send based on provider type
	var err error
	var messageID string

	switch sp.providerType {
	case SMSProviderTwilio:
		messageID, err = sp.sendViaTwilio(ctx, notification)
	case SMSProviderAWSSNS:
		messageID, err = sp.sendViaSNS(ctx, notification)
	case SMSProviderVonage:
		messageID, err = sp.sendViaVonage(ctx, notification)
	case SMSProviderPlivo:
		messageID, err = sp.sendViaPlivo(ctx, notification)
	default:
		err = fmt.Errorf("unsupported SMS provider: %s", sp.providerType)
	}

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

// SendBatch sends multiple SMS notifications in a batch
func (sp *SMSProvider) SendBatch(ctx context.Context, notifs []*notifications.Notification) ([]*notifications.SendResult, error) {
	results := make([]*notifications.SendResult, len(notifs))

	for i, notif := range notifs {
		result, err := sp.Send(ctx, notif)
		if err != nil {
			results[i] = result
			continue
		}
		results[i] = result
	}

	return results, nil
}

// Validate validates the notification data for SMS
func (sp *SMSProvider) Validate(notification *notifications.Notification) error {
	if notification.RecipientPhone == "" {
		return errors.New("recipient phone number is required")
	}

	if notification.Content == "" {
		return errors.New("SMS content is required")
	}

	// Check SMS length (standard is 160 characters, extended is 1600)
	if len(notification.Content) > 1600 {
		return errors.New("SMS content exceeds maximum length of 1600 characters")
	}

	return nil
}

// GetChannel returns the channel type
func (sp *SMSProvider) GetChannel() notifications.NotificationChannel {
	return notifications.ChannelSMS
}

// GetProviderName returns the provider name
func (sp *SMSProvider) GetProviderName() string {
	return string(sp.providerType)
}

// HealthCheck checks if the provider is healthy
func (sp *SMSProvider) HealthCheck(ctx context.Context) error {
	// Implement provider-specific health checks
	switch sp.providerType {
	case SMSProviderTwilio:
		return sp.healthCheckTwilio(ctx)
	case SMSProviderAWSSNS:
		return sp.healthCheckSNS(ctx)
	case SMSProviderVonage:
		return sp.healthCheckVonage(ctx)
	case SMSProviderPlivo:
		return sp.healthCheckPlivo(ctx)
	}

	return nil
}

// GetCapabilities returns the provider capabilities
func (sp *SMSProvider) GetCapabilities() notifications.ProviderCapabilities {
	return notifications.ProviderCapabilities{
		SupportsBatch:       true,
		SupportsScheduling:  false,
		SupportsTracking:    true,
		SupportsTemplates:   true,
		SupportsAttachments: false,
		SupportsRichContent: false,
		MaxBatchSize:        1000,
		RateLimitPerMinute:  100,
	}
}

// Provider-specific implementations

func (sp *SMSProvider) sendViaTwilio(ctx context.Context, notification *notifications.Notification) (string, error) {
	// Twilio implementation
	// In production, use Twilio SDK: github.com/twilio/twilio-go

	/*
		Example structure:

		client := twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: sp.config.TwilioAccountSID,
			Password: sp.config.TwilioAuthToken,
		})

		params := &api.CreateMessageParams{}
		params.SetTo(notification.RecipientPhone)
		params.SetFrom(sp.config.TwilioFromNumber)
		params.SetBody(notification.Content)

		resp, err := client.Api.CreateMessage(params)
		if err != nil {
			return "", fmt.Errorf("failed to send SMS via Twilio: %w", err)
		}

		return *resp.Sid, nil
	*/

	return "", errors.New("Twilio implementation requires Twilio SDK (github.com/twilio/twilio-go)")
}

func (sp *SMSProvider) sendViaSNS(ctx context.Context, notification *notifications.Notification) (string, error) {
	// AWS SNS implementation
	// In production, use AWS SDK: github.com/aws/aws-sdk-go-v2/service/sns

	/*
		Example structure:

		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(sp.config.SNSRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				sp.config.SNSAccessKeyID,
				sp.config.SNSSecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return "", fmt.Errorf("failed to load AWS config: %w", err)
		}

		client := sns.NewFromConfig(cfg)

		input := &sns.PublishInput{
			Message:     aws.String(notification.Content),
			PhoneNumber: aws.String(notification.RecipientPhone),
		}

		result, err := client.Publish(ctx, input)
		if err != nil {
			return "", fmt.Errorf("failed to send SMS via SNS: %w", err)
		}

		return *result.MessageId, nil
	*/

	return "", errors.New("SNS implementation requires AWS SDK (github.com/aws/aws-sdk-go-v2/service/sns)")
}

func (sp *SMSProvider) sendViaVonage(ctx context.Context, notification *notifications.Notification) (string, error) {
	// Vonage (Nexmo) implementation
	// In production, use Vonage SDK: github.com/vonage/vonage-go-sdk

	return "", errors.New("Vonage implementation requires Vonage SDK (github.com/vonage/vonage-go-sdk)")
}

func (sp *SMSProvider) sendViaPlivo(ctx context.Context, notification *notifications.Notification) (string, error) {
	// Plivo implementation
	// In production, use Plivo SDK: github.com/plivo/plivo-go

	return "", errors.New("Plivo implementation requires Plivo SDK (github.com/plivo/plivo-go)")
}

// Health check implementations

func (sp *SMSProvider) healthCheckTwilio(ctx context.Context) error {
	// Twilio health check
	if sp.config.TwilioAccountSID == "" || sp.config.TwilioAuthToken == "" {
		return errors.New("Twilio credentials not configured")
	}
	return nil
}

func (sp *SMSProvider) healthCheckSNS(ctx context.Context) error {
	// SNS health check
	if sp.config.SNSAccessKeyID == "" || sp.config.SNSSecretAccessKey == "" {
		return errors.New("AWS SNS credentials not configured")
	}
	return nil
}

func (sp *SMSProvider) healthCheckVonage(ctx context.Context) error {
	// Vonage health check
	if sp.config.VonageAPIKey == "" || sp.config.VonageAPISecret == "" {
		return errors.New("Vonage credentials not configured")
	}
	return nil
}

func (sp *SMSProvider) healthCheckPlivo(ctx context.Context) error {
	// Plivo health check
	if sp.config.PlivoAuthID == "" || sp.config.PlivoAuthToken == "" {
		return errors.New("Plivo credentials not configured")
	}
	return nil
}
