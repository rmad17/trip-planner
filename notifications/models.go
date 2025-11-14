package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"triplanner/core"
)

// NotificationChannel represents different delivery channels
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelFirebase NotificationChannel = "firebase"
	ChannelPush     NotificationChannel = "push"
	ChannelWebhook  NotificationChannel = "webhook"
	ChannelInApp    NotificationChannel = "in_app"
)

// NotificationStatus represents the current status of a notification
type NotificationStatus string

const (
	StatusPending    NotificationStatus = "pending"
	StatusQueued     NotificationStatus = "queued"
	StatusSending    NotificationStatus = "sending"
	StatusSent       NotificationStatus = "sent"
	StatusDelivered  NotificationStatus = "delivered"
	StatusFailed     NotificationStatus = "failed"
	StatusRetrying   NotificationStatus = "retrying"
	StatusCancelled  NotificationStatus = "cancelled"
	StatusRead       NotificationStatus = "read"
	StatusArchived   NotificationStatus = "archived"
)

// NotificationPriority represents the priority level
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	TypeTransactional NotificationType = "transactional"
	TypeMarketing     NotificationType = "marketing"
	TypeAlert         NotificationType = "alert"
	TypeReminder      NotificationType = "reminder"
	TypeSystem        NotificationType = "system"
)

// JSONB type for storing JSON data in PostgreSQL
type JSONB map[string]interface{}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB: expected []byte, got %T", value)
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

// Notification represents a notification record
type Notification struct {
	core.BaseModel

	// Core fields
	Type        NotificationType   `gorm:"type:varchar(50);not null;index" json:"type"`
	Channel     NotificationChannel `gorm:"type:varchar(50);not null;index" json:"channel"`
	Priority    NotificationPriority `gorm:"type:varchar(20);default:'normal';index" json:"priority"`
	Status      NotificationStatus  `gorm:"type:varchar(50);default:'pending';index" json:"status"`

	// Sender and Recipient
	SenderID    *uuid.UUID `gorm:"type:uuid;index" json:"sender_id,omitempty"`
	RecipientID *uuid.UUID `gorm:"type:uuid;index" json:"recipient_id,omitempty"`

	// Content
	Subject     string `gorm:"type:text" json:"subject,omitempty"`
	Content     string `gorm:"type:text;not null" json:"content"`
	ContentHTML string `gorm:"type:text" json:"content_html,omitempty"`

	// Template information
	TemplateID   *uuid.UUID `gorm:"type:uuid;index" json:"template_id,omitempty"`
	TemplateData JSONB      `gorm:"type:jsonb" json:"template_data,omitempty"`

	// Channel-specific data
	ChannelProvider string `gorm:"type:varchar(100)" json:"channel_provider,omitempty"` // e.g., "sendgrid", "twilio", "firebase"
	ChannelData     JSONB  `gorm:"type:jsonb" json:"channel_data,omitempty"`

	// Delivery information
	RecipientEmail    string `gorm:"type:varchar(255);index" json:"recipient_email,omitempty"`
	RecipientPhone    string `gorm:"type:varchar(50);index" json:"recipient_phone,omitempty"`
	RecipientDeviceID string `gorm:"type:varchar(255);index" json:"recipient_device_id,omitempty"`
	RecipientWebhook  string `gorm:"type:varchar(500)" json:"recipient_webhook,omitempty"`

	// Tracking
	ScheduledAt *time.Time `gorm:"index" json:"scheduled_at,omitempty"`
	SentAt      *time.Time `gorm:"index" json:"sent_at,omitempty"`
	DeliveredAt *time.Time `gorm:"index" json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`

	// Retry mechanism
	RetryCount    int    `gorm:"default:0" json:"retry_count"`
	MaxRetries    int    `gorm:"default:3" json:"max_retries"`
	NextRetryAt   *time.Time `gorm:"index" json:"next_retry_at,omitempty"`
	LastError     string `gorm:"type:text" json:"last_error,omitempty"`

	// Metadata
	Metadata        JSONB  `gorm:"type:jsonb" json:"metadata,omitempty"`
	EntityType      string `gorm:"type:varchar(100);index" json:"entity_type,omitempty"` // e.g., "trip", "expense"
	EntityID        *uuid.UUID `gorm:"type:uuid;index" json:"entity_id,omitempty"`
	Tags            []string `gorm:"type:text[];index:,type:gin" json:"tags,omitempty"`

	// External tracking
	ExternalID       string `gorm:"type:varchar(255);index" json:"external_id,omitempty"` // Provider's message ID
	ExternalResponse JSONB  `gorm:"type:jsonb" json:"external_response,omitempty"`

	// Expiry and archival
	ExpiresAt  *time.Time `gorm:"index" json:"expires_at,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`

	// Relationships
	Template *NotificationTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Audits   []NotificationAudit   `gorm:"foreignKey:NotificationID" json:"audits,omitempty"`
}

// NotificationTemplate represents reusable notification templates
type NotificationTemplate struct {
	core.BaseModel

	Name        string              `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description string              `gorm:"type:text" json:"description,omitempty"`
	Type        NotificationType    `gorm:"type:varchar(50);not null;index" json:"type"`
	Channel     NotificationChannel `gorm:"type:varchar(50);not null;index" json:"channel"`

	// Template content (supports Go template syntax)
	Subject     string `gorm:"type:text" json:"subject,omitempty"`
	Content     string `gorm:"type:text;not null" json:"content"`
	ContentHTML string `gorm:"type:text" json:"content_html,omitempty"`

	// Default values
	DefaultPriority NotificationPriority `gorm:"type:varchar(20);default:'normal'" json:"default_priority"`
	DefaultMetadata JSONB                `gorm:"type:jsonb" json:"default_metadata,omitempty"`

	// Template variables documentation
	Variables JSONB `gorm:"type:jsonb" json:"variables,omitempty"` // Schema/docs for template variables

	// Template settings
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`
	Version     int    `gorm:"default:1" json:"version"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`

	// Localization
	Locale      string `gorm:"type:varchar(10);default:'en';index" json:"locale"`

	// Relationships
	Notifications []Notification `gorm:"foreignKey:TemplateID" json:"notifications,omitempty"`
}

// NotificationAudit represents audit trail for notifications
type NotificationAudit struct {
	core.BaseModel

	NotificationID uuid.UUID          `gorm:"type:uuid;not null;index" json:"notification_id"`
	Status         NotificationStatus `gorm:"type:varchar(50);not null" json:"status"`

	// Event details
	Event       string    `gorm:"type:varchar(100);not null" json:"event"` // e.g., "created", "sent", "delivered", "failed"
	Message     string    `gorm:"type:text" json:"message,omitempty"`
	Details     JSONB     `gorm:"type:jsonb" json:"details,omitempty"`

	// Timing
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`

	// Request/Response tracking
	RequestData  JSONB `gorm:"type:jsonb" json:"request_data,omitempty"`
	ResponseData JSONB `gorm:"type:jsonb" json:"response_data,omitempty"`

	// Provider information
	Provider         string `gorm:"type:varchar(100)" json:"provider,omitempty"`
	ProviderResponse string `gorm:"type:text" json:"provider_response,omitempty"`

	// Error tracking
	IsError      bool   `gorm:"default:false;index" json:"is_error"`
	ErrorCode    string `gorm:"type:varchar(100)" json:"error_code,omitempty"`
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// Actor
	ActorID   *uuid.UUID `gorm:"type:uuid" json:"actor_id,omitempty"`
	ActorType string     `gorm:"type:varchar(50)" json:"actor_type,omitempty"` // "system", "user", "api"

	// Additional metadata
	Metadata JSONB `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Relationships
	Notification *Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
}

// NotificationPreference represents user preferences for notifications
type NotificationPreference struct {
	core.BaseModel

	UserID  uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Channel NotificationChannel `gorm:"type:varchar(50);not null" json:"channel"`
	Type    NotificationType    `gorm:"type:varchar(50);not null" json:"type"`

	// Preference settings
	IsEnabled   bool  `gorm:"default:true" json:"is_enabled"`
	Priority    NotificationPriority `gorm:"type:varchar(20)" json:"priority,omitempty"`

	// Quiet hours
	QuietHoursStart *time.Time `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *time.Time `json:"quiet_hours_end,omitempty"`
	Timezone        string     `gorm:"type:varchar(50)" json:"timezone,omitempty"`

	// Frequency control
	MaxPerDay   int `json:"max_per_day,omitempty"`
	MaxPerWeek  int `json:"max_per_week,omitempty"`
	MaxPerMonth int `json:"max_per_month,omitempty"`

	// Metadata
	Metadata JSONB `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Unique constraint on user + channel + type
	// This ensures one preference record per user per channel per type
}

// NotificationBatch represents a batch of notifications sent together
type NotificationBatch struct {
	core.BaseModel

	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Batch properties
	Channel     NotificationChannel  `gorm:"type:varchar(50);not null;index" json:"channel"`
	Type        NotificationType     `gorm:"type:varchar(50);not null;index" json:"type"`
	Priority    NotificationPriority `gorm:"type:varchar(20);default:'normal'" json:"priority"`

	// Template
	TemplateID *uuid.UUID `gorm:"type:uuid" json:"template_id,omitempty"`

	// Tracking
	TotalCount     int `gorm:"default:0" json:"total_count"`
	SentCount      int `gorm:"default:0" json:"sent_count"`
	DeliveredCount int `gorm:"default:0" json:"delivered_count"`
	FailedCount    int `gorm:"default:0" json:"failed_count"`

	// Status
	Status      string     `gorm:"type:varchar(50);default:'draft';index" json:"status"` // draft, scheduled, processing, completed, failed
	ScheduledAt *time.Time `gorm:"index" json:"scheduled_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Creator
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`

	// Metadata
	Metadata JSONB `gorm:"type:jsonb" json:"metadata,omitempty"`
}

// NotificationProvider represents configuration for notification providers
type NotificationProvider struct {
	core.BaseModel

	Name        string              `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Channel     NotificationChannel `gorm:"type:varchar(50);not null;index" json:"channel"`
	Provider    string              `gorm:"type:varchar(100);not null" json:"provider"` // e.g., "sendgrid", "twilio", "firebase"
	Description string              `gorm:"type:text" json:"description,omitempty"`

	// Configuration (encrypted in production)
	Config JSONB `gorm:"type:jsonb;not null" json:"config"` // API keys, endpoints, etc.

	// Status
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
	Priority    int    `gorm:"default:0" json:"priority"` // For fallback ordering

	// Rate limiting
	RateLimit      int `gorm:"default:0" json:"rate_limit,omitempty"` // Max per minute, 0 = unlimited
	CurrentUsage   int `gorm:"default:0" json:"current_usage"`
	UsageResetAt   *time.Time `json:"usage_reset_at,omitempty"`

	// Health tracking
	HealthStatus   string     `gorm:"type:varchar(50);default:'healthy'" json:"health_status"` // healthy, degraded, down
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
	LastError      string     `gorm:"type:text" json:"last_error,omitempty"`

	// Metadata
	Metadata JSONB `gorm:"type:jsonb" json:"metadata,omitempty"`
}

// TableName overrides
func (Notification) TableName() string {
	return "notifications"
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

func (NotificationAudit) TableName() string {
	return "notification_audits"
}

func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

func (NotificationBatch) TableName() string {
	return "notification_batches"
}

func (NotificationProvider) TableName() string {
	return "notification_providers"
}
