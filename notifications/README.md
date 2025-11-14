# Notification System

A comprehensive, cloud-agnostic notification system for the Trip Planner application. This module provides multi-channel notification delivery with audit tracking, template management, and user preferences.

## Features

- **Multi-Channel Support**: Email, SMS, Firebase FCM, Push Notifications, Webhooks, In-App
- **Cloud-Agnostic**: Support for multiple providers per channel (SendGrid, SES, Twilio, etc.)
- **Template Management**: Reusable templates with variable substitution
- **Audit Trail**: Complete tracking of notification lifecycle
- **User Preferences**: Fine-grained control over notification delivery
- **Retry Mechanism**: Automatic retry with exponential backoff
- **Batch Operations**: Send notifications to multiple recipients efficiently
- **Priority Levels**: Low, Normal, High, Critical
- **Scheduling**: Schedule notifications for future delivery
- **Independent Module**: Designed to be extracted as a microservice if needed

## Architecture

### Core Components

```
notifications/
├── models.go           # Database models
├── interfaces.go       # Service interfaces
├── service.go         # Core notification service
├── config.go          # Provider management
├── templates.go       # Template service
├── audit.go           # Audit service
├── preferences.go     # Preference service
├── controllers.go     # HTTP controllers
├── routers.go         # Route definitions
├── utils.go           # Helper functions
└── channels/          # Channel providers
    ├── email.go       # Email providers (SMTP, SendGrid, SES, Mailgun)
    ├── sms.go         # SMS providers (Twilio, AWS SNS, Vonage, Plivo)
    ├── firebase.go    # Firebase Cloud Messaging
    ├── push.go        # Generic push notifications (APNs, Web Push)
    └── webhook.go     # HTTP webhooks
```

### Database Models

#### Notification
- Core notification record with status tracking
- Supports all delivery channels
- Tracks delivery timestamps and external IDs

#### NotificationTemplate
- Reusable notification templates
- Supports Go template syntax
- Version controlled

#### NotificationAudit
- Complete audit trail for each notification
- Tracks state changes, errors, and provider responses

#### NotificationPreference
- User-specific notification preferences
- Quiet hours, frequency limits
- Channel and type-specific controls

#### NotificationBatch
- Batch notification management
- Progress tracking

#### NotificationProvider
- Provider configuration storage
- Health status monitoring
- Rate limiting

## Installation

### 1. Add to Database Migrations

The notification models are automatically included in database migrations through the `database/models.go` file.

Run migrations:
```bash
go run cmd/migrate/main.go
```

Or use Atlas:
```bash
atlas migrate apply --env local
```

### 2. Initialize Notification Service

```go
import (
    "triplanner/notifications"
    "triplanner/notifications/channels"
)

// Create notification service
config := notifications.DefaultConfig()
notifService := notifications.NewService(db, config)

// Register channel providers
providerManager := notifService.GetProviderManager()

// Register email provider
emailConfig := channels.EmailConfig{
    FromEmail:    "noreply@tripplanner.com",
    FromName:     "Trip Planner",
    SMTPHost:     "smtp.gmail.com",
    SMTPPort:     587,
    SMTPUsername: "your-email@gmail.com",
    SMTPPassword: "your-password",
    SMTPUseTLS:   true,
}
emailProvider := channels.NewEmailProvider(channels.EmailProviderSMTP, emailConfig)
providerManager.Register(emailProvider)

// Register SMS provider (Twilio example)
smsConfig := channels.SMSConfig{
    TwilioAccountSID: "your-account-sid",
    TwilioAuthToken:  "your-auth-token",
    TwilioFromNumber: "+1234567890",
}
smsProvider := channels.NewSMSProvider(channels.SMSProviderTwilio, smsConfig)
providerManager.Register(smsProvider)

// Register Firebase FCM provider
firebaseConfig := channels.FirebaseConfig{
    ProjectID:          "your-project-id",
    ServiceAccountJSON: []byte("..."),
}
firebaseProvider := channels.NewFirebaseProvider(firebaseConfig)
providerManager.Register(firebaseProvider)

// Register webhook provider
webhookConfig := channels.WebhookConfig{
    SigningSecret: "your-secret",
    Timeout:       30 * time.Second,
}
webhookProvider := channels.NewWebhookProvider(webhookConfig)
providerManager.Register(webhookProvider)
```

### 3. Register Routes

```go
import (
    "github.com/gin-gonic/gin"
    "triplanner/notifications"
)

// In your main application setup
router := gin.Default()
api := router.Group("/api/v1")

// Create controller
controller := notifications.NewController(notifService)

// Register routes
notifications.RegisterRoutes(api, controller)

// Register public routes for webhooks
public := router.Group("/public")
notifications.RegisterPublicRoutes(public, controller)
```

## Usage Examples

### 1. Send a Simple Email

```go
req := &notifications.SendRequest{
    Type:           notifications.TypeTransactional,
    Channel:        notifications.ChannelEmail,
    Priority:       notifications.PriorityNormal,
    RecipientEmail: "user@example.com",
    Subject:        "Welcome to Trip Planner",
    Content:        "Thank you for signing up!",
    ContentHTML:    "<h1>Thank you for signing up!</h1>",
}

notification, err := notifService.Send(ctx, req)
if err != nil {
    log.Printf("Failed to send notification: %v", err)
}
```

### 2. Send SMS

```go
req := &notifications.SendRequest{
    Type:           notifications.TypeAlert,
    Channel:        notifications.ChannelSMS,
    Priority:       notifications.PriorityHigh,
    RecipientPhone: "+1234567890",
    Content:        "Your trip to Paris starts tomorrow!",
}

notification, err := notifService.Send(ctx, req)
```

### 3. Send Push Notification via Firebase

```go
req := &notifications.SendRequest{
    Type:              notifications.TypeReminder,
    Channel:           notifications.ChannelFirebase,
    Priority:          notifications.PriorityHigh,
    RecipientDeviceID: "user-device-token",
    Subject:           "Trip Reminder",
    Content:           "Don't forget to pack your passport!",
    EntityType:        "trip",
    EntityID:          &tripID,
}

notification, err := notifService.Send(ctx, req)
```

### 4. Use Templates

```go
// Create a template
template := &notifications.NotificationTemplate{
    Name:        "trip_invitation",
    Type:        notifications.TypeTransactional,
    Channel:     notifications.ChannelEmail,
    Subject:     "You've been invited to {{.TripName}}",
    Content:     "{{.InviterName}} has invited you to join their trip to {{.Destination}}.",
    ContentHTML: "<h2>Trip Invitation</h2><p>{{.InviterName}} has invited you to join their trip to {{.Destination}}.</p>",
    IsActive:    true,
}

err := templateService.Create(ctx, template)

// Send using template
req := &notifications.TemplateSendRequest{
    TemplateName: "trip_invitation",
    TemplateData: map[string]interface{}{
        "TripName":     "Summer Adventure",
        "InviterName":  "John Doe",
        "Destination":  "Paris, France",
    },
    RecipientEmail: "friend@example.com",
}

notification, err := notifService.SendFromTemplate(ctx, req)
```

### 5. Send Batch Notifications

```go
recipients := []notifications.RecipientData{
    {
        RecipientEmail: "user1@example.com",
        TemplateData:   map[string]interface{}{"Name": "Alice"},
    },
    {
        RecipientEmail: "user2@example.com",
        TemplateData:   map[string]interface{}{"Name": "Bob"},
    },
}

batchReq := &notifications.BatchSendRequest{
    Name:       "Weekly Newsletter",
    Channel:    notifications.ChannelEmail,
    Type:       notifications.TypeMarketing,
    TemplateID: &templateID,
    Recipients: recipients,
}

batch, err := notifService.SendBatch(ctx, batchReq)
```

### 6. Schedule Notifications

```go
req := &notifications.SendRequest{
    Type:           notifications.TypeReminder,
    Channel:        notifications.ChannelEmail,
    RecipientEmail: "user@example.com",
    Subject:        "Trip starts tomorrow!",
    Content:        "Don't forget to check in!",
}

// Schedule for 24 hours before trip
scheduledTime := tripStartDate.Add(-24 * time.Hour)
notification, err := notifService.Schedule(ctx, req, scheduledTime)
```

### 7. User Preferences

```go
// Set user preference
pref := &notifications.NotificationPreference{
    UserID:    userID,
    Channel:   notifications.ChannelEmail,
    Type:      notifications.TypeMarketing,
    IsEnabled: false, // Opt out of marketing emails
}

err := prefService.Set(ctx, pref)

// Set quiet hours
pref := &notifications.NotificationPreference{
    UserID:          userID,
    Channel:         notifications.ChannelSMS,
    Type:            notifications.TypeAlert,
    IsEnabled:       true,
    QuietHoursStart: time.Parse("15:04", "22:00"),
    QuietHoursEnd:   time.Parse("15:04", "08:00"),
    Timezone:        "America/New_York",
}

err := prefService.Set(ctx, pref)
```

### 8. Webhook Notifications

```go
req := &notifications.SendRequest{
    Type:             notifications.TypeSystem,
    Channel:          notifications.ChannelWebhook,
    RecipientWebhook: "https://your-app.com/webhooks/notifications",
    Content:          "Payment received for trip booking",
    Metadata: map[string]interface{}{
        "event": "payment.received",
        "amount": 500.00,
        "currency": "USD",
    },
}

notification, err := notifService.Send(ctx, req)
```

## API Endpoints

### Notifications

```
POST   /api/v1/notifications/send                    # Send notification
POST   /api/v1/notifications/send/batch              # Send batch
POST   /api/v1/notifications/send/template           # Send from template
GET    /api/v1/notifications                         # List notifications
GET    /api/v1/notifications/:id                     # Get notification
GET    /api/v1/notifications/:id/audit               # Get audit trail
POST   /api/v1/notifications/:id/cancel              # Cancel notification
POST   /api/v1/notifications/:id/retry               # Retry failed notification
```

### Templates

```
POST   /api/v1/notifications/templates               # Create template
GET    /api/v1/notifications/templates               # List templates
GET    /api/v1/notifications/templates/:id           # Get template
PUT    /api/v1/notifications/templates/:id           # Update template
DELETE /api/v1/notifications/templates/:id           # Delete template
```

### Preferences

```
GET    /api/v1/notifications/preferences/:user_id    # Get user preferences
POST   /api/v1/notifications/preferences             # Set preference
```

## Channel Providers

### Email Providers

- **SMTP**: Direct SMTP connection (Gmail, custom servers)
- **SendGrid**: High-volume email delivery (requires SDK)
- **AWS SES**: Amazon Simple Email Service (requires SDK)
- **Mailgun**: Transactional email service (requires SDK)

### SMS Providers

- **Twilio**: Global SMS delivery (requires SDK)
- **AWS SNS**: Amazon Simple Notification Service (requires SDK)
- **Vonage**: Messaging platform (requires SDK)
- **Plivo**: SMS API platform (requires SDK)

### Push Notification Providers

- **Firebase FCM**: Cross-platform push notifications (requires SDK)
- **APNs**: Apple Push Notification service (requires SDK)
- **Web Push**: Browser push notifications (requires SDK)

### Webhook Provider

- Generic HTTP webhook support with signature verification
- Configurable authentication (Basic, Bearer, API Key)
- Automatic retries with exponential backoff

## Configuration

### Notification Config

```go
config := &notifications.NotificationConfig{
    EnableEmail:          true,
    EnableSMS:            true,
    EnableFirebase:       true,
    EnablePush:           true,
    EnableWebhook:        true,
    EnableInApp:          true,
    DefaultRetries:       3,
    RetryDelay:           60,      // seconds
    MaxRetryDelay:        3600,    // seconds
    ProcessInterval:      60,      // seconds
    BatchSize:            100,
    EnableQueue:          true,
    QueueType:            "memory",
    EnableAudit:          true,
    AuditRetentionDays:   90,
    GlobalRateLimit:      1000,    // per minute
    HealthCheckInterval:  300,     // seconds
}
```

## Background Processing

The notification system includes background processors for:

1. **Scheduled Notifications**: Processes notifications scheduled for delivery
2. **Retry Processing**: Handles failed notifications with retry logic
3. **Health Checks**: Monitors provider health

Example background worker:

```go
func StartNotificationWorker(ctx context.Context, service *notifications.Service) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Process scheduled notifications
            if err := service.ProcessScheduled(ctx); err != nil {
                log.Printf("Error processing scheduled: %v", err)
            }

            // Process retries
            if err := service.ProcessRetries(ctx); err != nil {
                log.Printf("Error processing retries: %v", err)
            }
        }
    }
}
```

## Security Considerations

1. **Credential Management**: Store provider credentials securely (environment variables, secret manager)
2. **Webhook Signatures**: Always verify webhook signatures
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Input Validation**: Validate all notification inputs
5. **User Consent**: Respect user preferences and quiet hours
6. **Data Privacy**: Handle PII according to privacy regulations

## Monitoring and Observability

### Audit Trail

Every notification has a complete audit trail:

```go
audits, err := service.GetAuditTrail(ctx, notificationID)
for _, audit := range audits {
    log.Printf("%s: %s - %s", audit.Timestamp, audit.Event, audit.Message)
}
```

### Provider Health

Monitor provider health status:

```go
providers := providerManager.GetProviders(notifications.ChannelEmail)
for _, provider := range providers {
    if err := provider.HealthCheck(ctx); err != nil {
        log.Printf("Provider %s is unhealthy: %v", provider.GetProviderName(), err)
    }
}
```

## Extending the System

### Adding a New Channel Provider

1. Implement the `ChannelProvider` interface
2. Create provider-specific configuration
3. Register with the provider manager

Example:

```go
type CustomProvider struct {
    config CustomConfig
}

func (p *CustomProvider) Send(ctx context.Context, notification *notifications.Notification) (*notifications.SendResult, error) {
    // Implementation
}

func (p *CustomProvider) GetChannel() notifications.NotificationChannel {
    return notifications.ChannelCustom
}

// Implement other interface methods...

// Register
provider := NewCustomProvider(config)
providerManager.Register(provider)
```

### Adding New Notification Types

Define new types in `models.go`:

```go
const (
    TypeBookingConfirmation NotificationType = "booking_confirmation"
    TypePaymentReminder     NotificationType = "payment_reminder"
)
```

## Testing

Run tests:

```bash
go test ./notifications/... -v
```

## Migration to Microservice

The notification system is designed to be extracted as an independent microservice:

1. Move the `notifications` package to a separate repository
2. Expose the service via gRPC or REST API
3. Update the main application to call the notification service
4. Use message queue (RabbitMQ, Kafka) for async processing

## License

MIT License - see LICENSE file for details
