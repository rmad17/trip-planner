package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditServiceImpl implements the AuditService interface
type AuditServiceImpl struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditServiceImpl {
	return &AuditServiceImpl{db: db}
}

// Log logs an audit entry
func (as *AuditServiceImpl) Log(ctx context.Context, audit *NotificationAudit) error {
	// Set timestamp if not set
	if audit.Timestamp.IsZero() {
		audit.Timestamp = time.Now()
	}

	// Create audit entry
	if err := as.db.WithContext(ctx).Create(audit).Error; err != nil {
		// Don't fail the main operation if audit logging fails
		// In production, you'd want to log this error
		fmt.Printf("WARNING: Failed to log audit entry: %v\n", err)
		return nil
	}

	return nil
}

// GetTrail gets the audit trail for a notification
func (as *AuditServiceImpl) GetTrail(ctx context.Context, notificationID uuid.UUID) ([]NotificationAudit, error) {
	var audits []NotificationAudit

	if err := as.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("timestamp ASC").
		Find(&audits).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}

	return audits, nil
}

// Search searches audit entries with filters
func (as *AuditServiceImpl) Search(ctx context.Context, filters AuditFilters) ([]NotificationAudit, error) {
	query := as.db.WithContext(ctx).Model(&NotificationAudit{})

	if filters.NotificationID != nil {
		query = query.Where("notification_id = ?", *filters.NotificationID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Event != "" {
		query = query.Where("event = ?", filters.Event)
	}

	if filters.Provider != "" {
		query = query.Where("provider = ?", filters.Provider)
	}

	if filters.IsError != nil {
		query = query.Where("is_error = ?", *filters.IsError)
	}

	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}

	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var audits []NotificationAudit
	if err := query.Order("timestamp DESC").Find(&audits).Error; err != nil {
		return nil, fmt.Errorf("failed to search audits: %w", err)
	}

	return audits, nil
}

// CleanupOldAudits removes audit entries older than retention period
func (as *AuditServiceImpl) CleanupOldAudits(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := as.db.WithContext(ctx).
		Where("timestamp < ?", cutoffDate).
		Delete(&NotificationAudit{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old audits: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetStatistics gets audit statistics
func (as *AuditServiceImpl) GetStatistics(ctx context.Context, notificationID *uuid.UUID, startTime, endTime *time.Time) (map[string]interface{}, error) {
	query := as.db.WithContext(ctx).Model(&NotificationAudit{})

	if notificationID != nil {
		query = query.Where("notification_id = ?", *notificationID)
	}

	if startTime != nil {
		query = query.Where("timestamp >= ?", *startTime)
	}

	if endTime != nil {
		query = query.Where("timestamp <= ?", *endTime)
	}

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}

	if err := query.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	// Count errors
	var errorCount int64
	if err := query.Where("is_error = ?", true).Count(&errorCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count errors: %w", err)
	}

	// Count by provider
	var providerCounts []struct {
		Provider string
		Count    int64
	}

	if err := query.Select("provider, COUNT(*) as count").
		Where("provider IS NOT NULL AND provider != ''").
		Group("provider").
		Scan(&providerCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get provider counts: %w", err)
	}

	stats := map[string]interface{}{
		"status_counts":   statusCounts,
		"error_count":     errorCount,
		"provider_counts": providerCounts,
	}

	return stats, nil
}
