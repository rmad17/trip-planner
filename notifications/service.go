package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidRequest      = errors.New("invalid notification request")
	ErrRecipientRequired   = errors.New("recipient information required")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrCannotCancel        = errors.New("notification cannot be cancelled")
	ErrTemplateNotFound    = errors.New("template not found")
)

// Service implements the NotificationService interface
type Service struct {
	db              *gorm.DB
	providerManager *ProviderManager
	templateService *TemplateServiceImpl
	auditService    *AuditServiceImpl
	prefService     *PreferenceServiceImpl
	config          *NotificationConfig
}

// NewService creates a new notification service
func NewService(db *gorm.DB, config *NotificationConfig) *Service {
	pm := NewProviderManager(db)

	return &Service{
		db:              db,
		providerManager: pm,
		templateService: NewTemplateService(db),
		auditService:    NewAuditService(db),
		prefService:     NewPreferenceService(db),
		config:          config,
	}
}

// GetProviderManager returns the provider manager
func (s *Service) GetProviderManager() *ProviderManager {
	return s.providerManager
}

// GetTemplateService returns the template service
func (s *Service) GetTemplateService() TemplateService {
	return s.templateService
}

// GetAuditService returns the audit service
func (s *Service) GetAuditService() AuditService {
	return s.auditService
}

// GetPreferenceService returns the preference service
func (s *Service) GetPreferenceService() PreferenceService {
	return s.prefService
}

// Send sends a single notification
func (s *Service) Send(ctx context.Context, req *SendRequest) (*Notification, error) {
	// Validate request
	if err := s.validateSendRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	// Check user preferences if recipient ID is provided
	if req.RecipientID != nil {
		canSend, reason, err := s.prefService.CanSend(ctx, *req.RecipientID, req.Channel, req.Type)
		if err != nil {
			// Log but don't fail
			s.logError(ctx, "failed to check preferences", err)
		} else if !canSend {
			return nil, fmt.Errorf("notification blocked by user preference: %s", reason)
		}
	}

	// Create notification record
	notification := s.createNotificationFromRequest(req)

	// Save to database
	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Log creation
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         StatusPending,
		Event:          "created",
		Message:        "Notification created",
		Timestamp:      time.Now(),
		ActorType:      "system",
	})

	// Send immediately or queue
	go s.processNotification(context.Background(), notification)

	return notification, nil
}

// SendBatch sends multiple notifications
func (s *Service) SendBatch(ctx context.Context, batchReq *BatchSendRequest) (*NotificationBatch, error) {
	// Create batch record
	batch := &NotificationBatch{
		Name:        batchReq.Name,
		Description: batchReq.Description,
		Channel:     batchReq.Channel,
		Type:        batchReq.Type,
		Priority:    batchReq.Priority,
		TemplateID:  batchReq.TemplateID,
		TotalCount:  len(batchReq.Recipients),
		Status:      "processing",
		ScheduledAt: batchReq.ScheduledAt,
		StartedAt:   TimePtr(time.Now()),
		Metadata:    batchReq.Metadata,
	}

	if err := s.db.WithContext(ctx).Create(batch).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	// Create notifications for each recipient
	notifications := make([]*Notification, 0, len(batchReq.Recipients))

	for _, recipient := range batchReq.Recipients {
		notif := &Notification{
			Type:              batchReq.Type,
			Channel:           batchReq.Channel,
			Priority:          batchReq.Priority,
			Status:            StatusQueued,
			RecipientID:       recipient.RecipientID,
			RecipientEmail:    recipient.RecipientEmail,
			RecipientPhone:    recipient.RecipientPhone,
			RecipientDeviceID: recipient.RecipientDeviceID,
			TemplateID:        batchReq.TemplateID,
			TemplateData:      recipient.TemplateData,
			Metadata:          recipient.Metadata,
			ScheduledAt:       batchReq.ScheduledAt,
			MaxRetries:        s.config.DefaultRetries,
		}

		notifications = append(notifications, notif)
	}

	// Bulk insert
	if err := s.db.WithContext(ctx).CreateInBatches(notifications, s.config.BatchSize).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch notifications: %w", err)
	}

	// Process asynchronously
	go s.processBatch(context.Background(), batch, notifications)

	return batch, nil
}

// SendFromTemplate sends a notification using a template
func (s *Service) SendFromTemplate(ctx context.Context, req *TemplateSendRequest) (*Notification, error) {
	// Get template
	var template *NotificationTemplate
	var err error

	if req.TemplateID != nil {
		template, err = s.templateService.Get(ctx, *req.TemplateID)
	} else {
		template, err = s.templateService.GetByName(ctx, req.TemplateName)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTemplateNotFound, err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template %s is not active", template.Name)
	}

	// Render template
	content, contentHTML, err := s.templateService.Render(ctx, template, req.TemplateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Build send request
	sendReq := &SendRequest{
		Type:              template.Type,
		Channel:           template.Channel,
		Priority:          template.DefaultPriority,
		SenderID:          req.SenderID,
		RecipientID:       req.RecipientID,
		Subject:           template.Subject,
		Content:           content,
		ContentHTML:       contentHTML,
		RecipientEmail:    req.RecipientEmail,
		RecipientPhone:    req.RecipientPhone,
		RecipientDeviceID: req.RecipientDeviceID,
		RecipientWebhook:  req.RecipientWebhook,
		Metadata:          req.Metadata,
		EntityType:        req.EntityType,
		EntityID:          req.EntityID,
		Tags:              req.Tags,
	}

	// Override with request values
	if req.Channel != nil {
		sendReq.Channel = *req.Channel
	}
	if req.Priority != nil {
		sendReq.Priority = *req.Priority
	}

	// Schedule or send immediately
	if req.ScheduledAt != nil {
		return s.Schedule(ctx, sendReq, *req.ScheduledAt)
	}

	return s.Send(ctx, sendReq)
}

// Schedule schedules a notification for later delivery
func (s *Service) Schedule(ctx context.Context, req *SendRequest, scheduledAt time.Time) (*Notification, error) {
	// Validate request
	if err := s.validateSendRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	// Create notification record
	notification := s.createNotificationFromRequest(req)
	notification.Status = StatusQueued
	notification.ScheduledAt = &scheduledAt

	// Save to database
	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create scheduled notification: %w", err)
	}

	// Log creation
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         StatusQueued,
		Event:          "scheduled",
		Message:        fmt.Sprintf("Notification scheduled for %s", scheduledAt.Format(time.RFC3339)),
		Timestamp:      time.Now(),
		ActorType:      "system",
	})

	return notification, nil
}

// Cancel cancels a pending or scheduled notification
func (s *Service) Cancel(ctx context.Context, notificationID uuid.UUID) error {
	var notification Notification
	if err := s.db.WithContext(ctx).First(&notification, "id = ?", notificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to find notification: %w", err)
	}

	// Check if can be cancelled
	if notification.Status != StatusPending && notification.Status != StatusQueued {
		return fmt.Errorf("%w: status is %s", ErrCannotCancel, notification.Status)
	}

	// Update status
	notification.Status = StatusCancelled
	if err := s.db.WithContext(ctx).Save(&notification).Error; err != nil {
		return fmt.Errorf("failed to cancel notification: %w", err)
	}

	// Log cancellation
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         StatusCancelled,
		Event:          "cancelled",
		Message:        "Notification cancelled",
		Timestamp:      time.Now(),
		ActorType:      "system",
	})

	return nil
}

// Retry retries a failed notification
func (s *Service) Retry(ctx context.Context, notificationID uuid.UUID) error {
	var notification Notification
	if err := s.db.WithContext(ctx).First(&notification, "id = ?", notificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to find notification: %w", err)
	}

	// Reset for retry
	notification.Status = StatusRetrying
	notification.RetryCount = 0
	notification.NextRetryAt = nil
	notification.LastError = ""

	if err := s.db.WithContext(ctx).Save(&notification).Error; err != nil {
		return fmt.Errorf("failed to update notification for retry: %w", err)
	}

	// Log retry
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         StatusRetrying,
		Event:          "retry_initiated",
		Message:        "Manual retry initiated",
		Timestamp:      time.Now(),
		ActorType:      "system",
	})

	// Process immediately
	go s.processNotification(context.Background(), &notification)

	return nil
}

// GetStatus gets the status of a notification
func (s *Service) GetStatus(ctx context.Context, notificationID uuid.UUID) (*Notification, error) {
	var notification Notification
	if err := s.db.WithContext(ctx).
		Preload("Template").
		First(&notification, "id = ?", notificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

// GetAuditTrail gets the audit trail for a notification
func (s *Service) GetAuditTrail(ctx context.Context, notificationID uuid.UUID) ([]NotificationAudit, error) {
	return s.auditService.GetTrail(ctx, notificationID)
}

// ProcessScheduled processes scheduled notifications that are due
func (s *Service) ProcessScheduled(ctx context.Context) error {
	var notifications []Notification
	now := time.Now()

	if err := s.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", StatusQueued, now).
		Limit(s.config.BatchSize).
		Find(&notifications).Error; err != nil {
		return fmt.Errorf("failed to fetch scheduled notifications: %w", err)
	}

	for i := range notifications {
		go s.processNotification(context.Background(), &notifications[i])
	}

	return nil
}

// ProcessRetries processes notifications that need to be retried
func (s *Service) ProcessRetries(ctx context.Context) error {
	var notifications []Notification
	now := time.Now()

	if err := s.db.WithContext(ctx).
		Where("status = ? AND next_retry_at <= ? AND retry_count < max_retries", StatusRetrying, now).
		Limit(s.config.BatchSize).
		Find(&notifications).Error; err != nil {
		return fmt.Errorf("failed to fetch retry notifications: %w", err)
	}

	for i := range notifications {
		go s.processNotification(context.Background(), &notifications[i])
	}

	return nil
}

// Helper methods

func (s *Service) validateSendRequest(req *SendRequest) error {
	if req.Content == "" {
		return errors.New("content is required")
	}

	// Validate recipient based on channel
	switch req.Channel {
	case ChannelEmail:
		if req.RecipientEmail == "" {
			return ErrRecipientRequired
		}
	case ChannelSMS:
		if req.RecipientPhone == "" {
			return ErrRecipientRequired
		}
	case ChannelFirebase, ChannelPush:
		if req.RecipientDeviceID == "" {
			return ErrRecipientRequired
		}
	case ChannelWebhook:
		if req.RecipientWebhook == "" {
			return ErrRecipientRequired
		}
	}

	return nil
}

func (s *Service) createNotificationFromRequest(req *SendRequest) *Notification {
	return &Notification{
		Type:              req.Type,
		Channel:           req.Channel,
		Priority:          req.Priority,
		Status:            StatusPending,
		SenderID:          req.SenderID,
		RecipientID:       req.RecipientID,
		Subject:           req.Subject,
		Content:           req.Content,
		ContentHTML:       req.ContentHTML,
		ChannelProvider:   req.ChannelProvider,
		ChannelData:       req.ChannelData,
		RecipientEmail:    req.RecipientEmail,
		RecipientPhone:    req.RecipientPhone,
		RecipientDeviceID: req.RecipientDeviceID,
		RecipientWebhook:  req.RecipientWebhook,
		MaxRetries:        s.config.DefaultRetries,
		Metadata:          req.Metadata,
		EntityType:        req.EntityType,
		EntityID:          req.EntityID,
		Tags:              req.Tags,
		ExpiresAt:         req.ExpiresAt,
	}
}

func (s *Service) processNotification(ctx context.Context, notification *Notification) {
	// Update status to sending
	notification.Status = StatusSending
	s.db.WithContext(ctx).Model(notification).Update("status", StatusSending)

	// Log sending attempt
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         StatusSending,
		Event:          "sending",
		Message:        fmt.Sprintf("Attempt %d of %d", notification.RetryCount+1, notification.MaxRetries+1),
		Timestamp:      time.Now(),
		ActorType:      "system",
	})

	// Get provider
	provider, err := s.providerManager.GetProvider(notification.Channel, notification.ChannelProvider)
	if err != nil {
		s.handleSendError(ctx, notification, err)
		return
	}

	// Send notification
	result, err := provider.Send(ctx, notification)
	if err != nil || !result.Success {
		if err != nil {
			s.handleSendError(ctx, notification, err)
		} else {
			s.handleSendError(ctx, notification, errors.New(result.ErrorMessage))
		}
		return
	}

	// Update notification with result
	now := time.Now()
	notification.Status = result.Status
	notification.SentAt = &now
	notification.ExternalID = result.ExternalID
	if result.ProviderResponse != nil {
		if provResp, ok := result.ProviderResponse.(map[string]interface{}); ok {
			notification.ExternalResponse = provResp
		}
	}

	if result.Status == StatusDelivered {
		notification.DeliveredAt = &now
	}

	s.db.WithContext(ctx).Save(notification)

	// Log success
	auditEntry := &NotificationAudit{
		NotificationID: notification.ID,
		Status:         result.Status,
		Event:          "sent",
		Message:        "Notification sent successfully",
		Timestamp:      time.Now(),
		Provider:       provider.GetProviderName(),
		ActorType:      "system",
	}

	// Add provider response if available
	if result.ProviderResponse != nil {
		if provResp, ok := result.ProviderResponse.(map[string]interface{}); ok {
			auditEntry.ResponseData = provResp
		}
	}

	_ = s.auditService.Log(ctx, auditEntry)
}

func (s *Service) handleSendError(ctx context.Context, notification *Notification, err error) {
	notification.RetryCount++
	notification.LastError = err.Error()
	now := time.Now()
	notification.FailedAt = &now

	// Check if should retry
	if notification.RetryCount < notification.MaxRetries {
		notification.Status = StatusRetrying

		// Calculate next retry time with exponential backoff
		retryDelay := s.config.RetryDelay * (1 << (notification.RetryCount - 1))
		if retryDelay > s.config.MaxRetryDelay {
			retryDelay = s.config.MaxRetryDelay
		}
		nextRetry := now.Add(time.Duration(retryDelay) * time.Second)
		notification.NextRetryAt = &nextRetry
	} else {
		notification.Status = StatusFailed
	}

	s.db.WithContext(ctx).Save(notification)

	// Log error
	_ = s.auditService.Log(ctx, &NotificationAudit{
		NotificationID: notification.ID,
		Status:         notification.Status,
		Event:          "send_failed",
		Message:        fmt.Sprintf("Send failed: %v", err),
		Timestamp:      time.Now(),
		IsError:        true,
		ErrorMessage:   err.Error(),
		ActorType:      "system",
	})
}

func (s *Service) processBatch(ctx context.Context, batch *NotificationBatch, notifications []*Notification) {
	for _, notif := range notifications {
		s.processNotification(ctx, notif)

		// Update batch counts
		var status string
		s.db.WithContext(ctx).Model(notif).Pluck("status", &status)

		switch NotificationStatus(status) {
		case StatusSent, StatusDelivered:
			batch.SentCount++
			if status == string(StatusDelivered) {
				batch.DeliveredCount++
			}
		case StatusFailed:
			batch.FailedCount++
		}

		s.db.WithContext(ctx).Model(batch).Updates(map[string]interface{}{
			"sent_count":      batch.SentCount,
			"delivered_count": batch.DeliveredCount,
			"failed_count":    batch.FailedCount,
		})
	}

	// Mark batch as completed
	now := time.Now()
	batch.Status = "completed"
	batch.CompletedAt = &now
	s.db.WithContext(ctx).Save(batch)
}

func (s *Service) logError(ctx context.Context, message string, err error) {
	// In a production system, use proper logging
	fmt.Printf("ERROR: %s: %v\n", message, err)
}
