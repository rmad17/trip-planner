package notifications

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChannelProvider defines the interface that all notification channel providers must implement
type ChannelProvider interface {
	// Send sends a notification through the channel
	Send(ctx context.Context, notification *Notification) (*SendResult, error)

	// SendBatch sends multiple notifications in a batch (optional optimization)
	SendBatch(ctx context.Context, notifications []*Notification) ([]*SendResult, error)

	// Validate validates the notification data for this channel
	Validate(notification *Notification) error

	// GetChannel returns the channel type this provider handles
	GetChannel() NotificationChannel

	// GetProviderName returns the name of the provider (e.g., "sendgrid", "twilio")
	GetProviderName() string

	// HealthCheck checks if the provider is healthy and ready to send
	HealthCheck(ctx context.Context) error

	// GetCapabilities returns the capabilities supported by this provider
	GetCapabilities() ProviderCapabilities
}

// SendResult represents the result of sending a notification
type SendResult struct {
	Success      bool               `json:"success"`
	MessageID    string             `json:"message_id,omitempty"`     // Provider's message ID
	ExternalID   string             `json:"external_id,omitempty"`    // External tracking ID
	Status       NotificationStatus `json:"status"`
	SentAt       time.Time          `json:"sent_at"`
	Error        error              `json:"error,omitempty"`
	ErrorCode    string             `json:"error_code,omitempty"`
	ErrorMessage string             `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ProviderResponse interface{} `json:"provider_response,omitempty"`
}

// ProviderCapabilities describes what a provider can do
type ProviderCapabilities struct {
	SupportsBatch        bool     `json:"supports_batch"`
	SupportsScheduling   bool     `json:"supports_scheduling"`
	SupportsTracking     bool     `json:"supports_tracking"`
	SupportsTemplates    bool     `json:"supports_templates"`
	SupportsAttachments  bool     `json:"supports_attachments"`
	SupportsRichContent  bool     `json:"supports_rich_content"` // HTML, markdown, etc.
	MaxBatchSize         int      `json:"max_batch_size,omitempty"`
	RateLimitPerMinute   int      `json:"rate_limit_per_minute,omitempty"`
	SupportedPriorities  []string `json:"supported_priorities,omitempty"`
}

// NotificationService defines the main notification service interface
type NotificationService interface {
	// Send sends a single notification
	Send(ctx context.Context, req *SendRequest) (*Notification, error)

	// SendBatch sends multiple notifications
	SendBatch(ctx context.Context, batchReq *BatchSendRequest) (*NotificationBatch, error)

	// SendFromTemplate sends a notification using a template
	SendFromTemplate(ctx context.Context, req *TemplateSendRequest) (*Notification, error)

	// Schedule schedules a notification for later delivery
	Schedule(ctx context.Context, req *SendRequest, scheduledAt time.Time) (*Notification, error)

	// Cancel cancels a pending or scheduled notification
	Cancel(ctx context.Context, notificationID uuid.UUID) error

	// Retry retries a failed notification
	Retry(ctx context.Context, notificationID uuid.UUID) error

	// GetStatus gets the status of a notification
	GetStatus(ctx context.Context, notificationID uuid.UUID) (*Notification, error)

	// GetAuditTrail gets the audit trail for a notification
	GetAuditTrail(ctx context.Context, notificationID uuid.UUID) ([]NotificationAudit, error)

	// ProcessScheduled processes scheduled notifications that are due
	ProcessScheduled(ctx context.Context) error

	// ProcessRetries processes notifications that need to be retried
	ProcessRetries(ctx context.Context) error
}

// TemplateService defines the template management interface
type TemplateService interface {
	// Create creates a new template
	Create(ctx context.Context, template *NotificationTemplate) error

	// Update updates an existing template
	Update(ctx context.Context, template *NotificationTemplate) error

	// Get gets a template by ID
	Get(ctx context.Context, templateID uuid.UUID) (*NotificationTemplate, error)

	// GetByName gets a template by name
	GetByName(ctx context.Context, name string) (*NotificationTemplate, error)

	// List lists all templates with optional filters
	List(ctx context.Context, filters TemplateFilters) ([]*NotificationTemplate, error)

	// Delete deletes a template
	Delete(ctx context.Context, templateID uuid.UUID) error

	// Render renders a template with data
	Render(ctx context.Context, template *NotificationTemplate, data map[string]interface{}) (string, string, error)
}

// AuditService defines the audit trail interface
type AuditService interface {
	// Log logs an audit entry
	Log(ctx context.Context, audit *NotificationAudit) error

	// GetTrail gets the audit trail for a notification
	GetTrail(ctx context.Context, notificationID uuid.UUID) ([]NotificationAudit, error)

	// Search searches audit entries with filters
	Search(ctx context.Context, filters AuditFilters) ([]NotificationAudit, error)
}

// PreferenceService defines the user preference interface
type PreferenceService interface {
	// Get gets user preferences
	Get(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (*NotificationPreference, error)

	// Set sets user preferences
	Set(ctx context.Context, pref *NotificationPreference) error

	// CanSend checks if a notification can be sent based on user preferences
	CanSend(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (bool, string, error)

	// List lists all preferences for a user
	List(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)
}

// SendRequest represents a request to send a notification
type SendRequest struct {
	Type        NotificationType     `json:"type" binding:"required"`
	Channel     NotificationChannel  `json:"channel" binding:"required"`
	Priority    NotificationPriority `json:"priority,omitempty"`

	// Sender and Recipient
	SenderID    *uuid.UUID `json:"sender_id,omitempty"`
	RecipientID *uuid.UUID `json:"recipient_id,omitempty"`

	// Content
	Subject     string `json:"subject,omitempty"`
	Content     string `json:"content" binding:"required"`
	ContentHTML string `json:"content_html,omitempty"`

	// Delivery details
	RecipientEmail    string `json:"recipient_email,omitempty"`
	RecipientPhone    string `json:"recipient_phone,omitempty"`
	RecipientDeviceID string `json:"recipient_device_id,omitempty"`
	RecipientWebhook  string `json:"recipient_webhook,omitempty"`

	// Channel-specific data
	ChannelProvider string                 `json:"channel_provider,omitempty"`
	ChannelData     map[string]interface{} `json:"channel_data,omitempty"`

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	Tags       []string               `json:"tags,omitempty"`

	// Retry settings
	MaxRetries int `json:"max_retries,omitempty"`

	// Expiry
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// TemplateSendRequest represents a request to send a notification using a template
type TemplateSendRequest struct {
	TemplateName string                 `json:"template_name" binding:"required"`
	TemplateID   *uuid.UUID             `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data" binding:"required"`

	// Override template defaults
	Channel  *NotificationChannel  `json:"channel,omitempty"`
	Priority *NotificationPriority `json:"priority,omitempty"`

	// Sender and Recipient
	SenderID    *uuid.UUID `json:"sender_id,omitempty"`
	RecipientID *uuid.UUID `json:"recipient_id,omitempty"`

	// Delivery details
	RecipientEmail    string `json:"recipient_email,omitempty"`
	RecipientPhone    string `json:"recipient_phone,omitempty"`
	RecipientDeviceID string `json:"recipient_device_id,omitempty"`
	RecipientWebhook  string `json:"recipient_webhook,omitempty"`

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	Tags       []string               `json:"tags,omitempty"`

	// Schedule
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// BatchSendRequest represents a request to send multiple notifications
type BatchSendRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	Channel     NotificationChannel    `json:"channel" binding:"required"`
	Type        NotificationType       `json:"type" binding:"required"`
	Priority    NotificationPriority   `json:"priority,omitempty"`
	TemplateID  *uuid.UUID             `json:"template_id,omitempty"`
	Recipients  []RecipientData        `json:"recipients" binding:"required,min=1"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RecipientData represents recipient data for batch sending
type RecipientData struct {
	RecipientID       *uuid.UUID             `json:"recipient_id,omitempty"`
	RecipientEmail    string                 `json:"recipient_email,omitempty"`
	RecipientPhone    string                 `json:"recipient_phone,omitempty"`
	RecipientDeviceID string                 `json:"recipient_device_id,omitempty"`
	TemplateData      map[string]interface{} `json:"template_data,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateFilters represents filters for template queries
type TemplateFilters struct {
	Channel  *NotificationChannel  `json:"channel,omitempty"`
	Type     *NotificationType     `json:"type,omitempty"`
	IsActive *bool                 `json:"is_active,omitempty"`
	Locale   string                `json:"locale,omitempty"`
	Limit    int                   `json:"limit,omitempty"`
	Offset   int                   `json:"offset,omitempty"`
}

// AuditFilters represents filters for audit queries
type AuditFilters struct {
	NotificationID *uuid.UUID         `json:"notification_id,omitempty"`
	Status         *NotificationStatus `json:"status,omitempty"`
	Event          string              `json:"event,omitempty"`
	Provider       string              `json:"provider,omitempty"`
	IsError        *bool               `json:"is_error,omitempty"`
	StartTime      *time.Time          `json:"start_time,omitempty"`
	EndTime        *time.Time          `json:"end_time,omitempty"`
	Limit          int                 `json:"limit,omitempty"`
	Offset         int                 `json:"offset,omitempty"`
}

// NotificationFilters represents filters for notification queries
type NotificationFilters struct {
	Channel     *NotificationChannel  `json:"channel,omitempty"`
	Type        *NotificationType     `json:"type,omitempty"`
	Status      *NotificationStatus   `json:"status,omitempty"`
	Priority    *NotificationPriority `json:"priority,omitempty"`
	RecipientID *uuid.UUID            `json:"recipient_id,omitempty"`
	SenderID    *uuid.UUID            `json:"sender_id,omitempty"`
	EntityType  string                `json:"entity_type,omitempty"`
	EntityID    *uuid.UUID            `json:"entity_id,omitempty"`
	StartDate   *time.Time            `json:"start_date,omitempty"`
	EndDate     *time.Time            `json:"end_date,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Limit       int                   `json:"limit,omitempty"`
	Offset      int                   `json:"offset,omitempty"`
}
