package channels

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"triplanner/notifications"
)

// EmailProviderType represents different email providers
type EmailProviderType string

const (
	EmailProviderSMTP     EmailProviderType = "smtp"
	EmailProviderSendGrid EmailProviderType = "sendgrid"
	EmailProviderSES      EmailProviderType = "ses"
	EmailProviderMailgun  EmailProviderType = "mailgun"
)

// EmailProvider implements the ChannelProvider interface for email
type EmailProvider struct {
	providerType EmailProviderType
	config       EmailConfig
}

// EmailConfig holds email provider configuration
type EmailConfig struct {
	// Common
	FromEmail string
	FromName  string

	// SMTP specific
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPUseTLS   bool

	// SendGrid specific
	SendGridAPIKey string

	// AWS SES specific
	SESRegion          string
	SESAccessKeyID     string
	SESSecretAccessKey string

	// Mailgun specific
	MailgunDomain string
	MailgunAPIKey string
}

// NewEmailProvider creates a new email provider
func NewEmailProvider(providerType EmailProviderType, config EmailConfig) *EmailProvider {
	return &EmailProvider{
		providerType: providerType,
		config:       config,
	}
}

// Send sends an email notification
func (ep *EmailProvider) Send(ctx context.Context, notification *notifications.Notification) (*notifications.SendResult, error) {
	result := &notifications.SendResult{
		Status: notifications.StatusSending,
		SentAt: time.Now(),
	}

	// Validate notification
	if err := ep.Validate(notification); err != nil {
		result.Success = false
		result.Error = err
		result.ErrorMessage = err.Error()
		result.Status = notifications.StatusFailed
		return result, err
	}

	// Send based on provider type
	var err error
	var messageID string

	switch ep.providerType {
	case EmailProviderSMTP:
		messageID, err = ep.sendViaSMTP(ctx, notification)
	case EmailProviderSendGrid:
		messageID, err = ep.sendViaSendGrid(ctx, notification)
	case EmailProviderSES:
		messageID, err = ep.sendViaSES(ctx, notification)
	case EmailProviderMailgun:
		messageID, err = ep.sendViaMailgun(ctx, notification)
	default:
		err = fmt.Errorf("unsupported email provider: %s", ep.providerType)
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

// SendBatch sends multiple emails in a batch
func (ep *EmailProvider) SendBatch(ctx context.Context, notifs []*notifications.Notification) ([]*notifications.SendResult, error) {
	results := make([]*notifications.SendResult, len(notifs))

	for i, notif := range notifs {
		result, err := ep.Send(ctx, notif)
		if err != nil {
			results[i] = result
			continue
		}
		results[i] = result
	}

	return results, nil
}

// Validate validates the notification data for email
func (ep *EmailProvider) Validate(notification *notifications.Notification) error {
	if notification.RecipientEmail == "" {
		return errors.New("recipient email is required")
	}

	if notification.Subject == "" {
		return errors.New("email subject is required")
	}

	if notification.Content == "" && notification.ContentHTML == "" {
		return errors.New("email content is required")
	}

	return nil
}

// GetChannel returns the channel type
func (ep *EmailProvider) GetChannel() notifications.NotificationChannel {
	return notifications.ChannelEmail
}

// GetProviderName returns the provider name
func (ep *EmailProvider) GetProviderName() string {
	return string(ep.providerType)
}

// HealthCheck checks if the provider is healthy
func (ep *EmailProvider) HealthCheck(ctx context.Context) error {
	// Implement provider-specific health checks
	switch ep.providerType {
	case EmailProviderSMTP:
		return ep.healthCheckSMTP(ctx)
	case EmailProviderSendGrid:
		return ep.healthCheckSendGrid(ctx)
	case EmailProviderSES:
		return ep.healthCheckSES(ctx)
	case EmailProviderMailgun:
		return ep.healthCheckMailgun(ctx)
	}

	return nil
}

// GetCapabilities returns the provider capabilities
func (ep *EmailProvider) GetCapabilities() notifications.ProviderCapabilities {
	return notifications.ProviderCapabilities{
		SupportsBatch:       true,
		SupportsScheduling:  false,
		SupportsTracking:    true,
		SupportsTemplates:   true,
		SupportsAttachments: true,
		SupportsRichContent: true,
		MaxBatchSize:        1000,
		RateLimitPerMinute:  100,
	}
}

// Provider-specific implementations

func (ep *EmailProvider) sendViaSMTP(ctx context.Context, notification *notifications.Notification) (string, error) {
	// Build email message
	from := fmt.Sprintf("%s <%s>", ep.config.FromName, ep.config.FromEmail)
	to := notification.RecipientEmail

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = notification.Subject
	headers["MIME-Version"] = "1.0"

	var message string
	if notification.ContentHTML != "" {
		headers["Content-Type"] = "text/html; charset=UTF-8"
		message = notification.ContentHTML
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
		message = notification.Content
	}

	// Build email body
	var emailBody string
	for k, v := range headers {
		emailBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	emailBody += "\r\n" + message

	// Send email
	addr := fmt.Sprintf("%s:%d", ep.config.SMTPHost, ep.config.SMTPPort)
	auth := smtp.PlainAuth("", ep.config.SMTPUsername, ep.config.SMTPPassword, ep.config.SMTPHost)

	var err error
	if ep.config.SMTPUseTLS {
		// Use TLS
		tlsconfig := &tls.Config{
			ServerName: ep.config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return "", fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, ep.config.SMTPHost)
		if err != nil {
			return "", fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return "", fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err = client.Mail(ep.config.FromEmail); err != nil {
			return "", fmt.Errorf("failed to set sender: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return "", fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return "", fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = w.Write([]byte(emailBody))
		if err != nil {
			return "", fmt.Errorf("failed to write email body: %w", err)
		}

		err = w.Close()
		if err != nil {
			return "", fmt.Errorf("failed to close data writer: %w", err)
		}

		client.Quit()
	} else {
		// Use plain SMTP
		err = smtp.SendMail(addr, auth, ep.config.FromEmail, []string{to}, []byte(emailBody))
		if err != nil {
			return "", fmt.Errorf("failed to send email: %w", err)
		}
	}

	// Generate a message ID (in production, extract from response headers)
	messageID := fmt.Sprintf("smtp-%d", time.Now().UnixNano())
	return messageID, nil
}

func (ep *EmailProvider) sendViaSendGrid(ctx context.Context, notification *notifications.Notification) (string, error) {
	// SendGrid implementation
	// In production, use the SendGrid SDK: github.com/sendgrid/sendgrid-go

	// Placeholder implementation
	return "", errors.New("SendGrid implementation requires SendGrid SDK (github.com/sendgrid/sendgrid-go)")
}

func (ep *EmailProvider) sendViaSES(ctx context.Context, notification *notifications.Notification) (string, error) {
	// AWS SES implementation
	// In production, use AWS SDK: github.com/aws/aws-sdk-go-v2/service/ses

	// Placeholder implementation
	return "", errors.New("SES implementation requires AWS SDK (github.com/aws/aws-sdk-go-v2/service/ses)")
}

func (ep *EmailProvider) sendViaMailgun(ctx context.Context, notification *notifications.Notification) (string, error) {
	// Mailgun implementation
	// In production, use Mailgun SDK: github.com/mailgun/mailgun-go/v4

	// Placeholder implementation
	return "", errors.New("Mailgun implementation requires Mailgun SDK (github.com/mailgun/mailgun-go/v4)")
}

// Health check implementations

func (ep *EmailProvider) healthCheckSMTP(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", ep.config.SMTPHost, ep.config.SMTPPort)

	if ep.config.SMTPUseTLS {
		tlsconfig := &tls.Config{
			ServerName: ep.config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return fmt.Errorf("SMTP health check failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, ep.config.SMTPHost)
		if err != nil {
			return fmt.Errorf("SMTP health check failed: %w", err)
		}
		defer client.Close()

		return nil
	}

	// Plain SMTP
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP health check failed: %w", err)
	}
	defer client.Close()

	return nil
}

func (ep *EmailProvider) healthCheckSendGrid(ctx context.Context) error {
	// SendGrid health check
	return nil
}

func (ep *EmailProvider) healthCheckSES(ctx context.Context) error {
	// SES health check
	return nil
}

func (ep *EmailProvider) healthCheckMailgun(ctx context.Context) error {
	// Mailgun health check
	return nil
}
