package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PreferenceServiceImpl implements the PreferenceService interface
type PreferenceServiceImpl struct {
	db *gorm.DB
}

// NewPreferenceService creates a new preference service
func NewPreferenceService(db *gorm.DB) *PreferenceServiceImpl {
	return &PreferenceServiceImpl{db: db}
}

// Get gets user preferences
func (ps *PreferenceServiceImpl) Get(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (*NotificationPreference, error) {
	var pref NotificationPreference

	err := ps.db.WithContext(ctx).
		Where("user_id = ? AND channel = ? AND type = ?", userID, channel, notifType).
		First(&pref).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default preference
			return &NotificationPreference{
				UserID:    userID,
				Channel:   channel,
				Type:      notifType,
				IsEnabled: true, // Default to enabled
			}, nil
		}
		return nil, fmt.Errorf("failed to get preference: %w", err)
	}

	return &pref, nil
}

// Set sets user preferences
func (ps *PreferenceServiceImpl) Set(ctx context.Context, pref *NotificationPreference) error {
	// Check if preference exists
	var existing NotificationPreference
	err := ps.db.WithContext(ctx).
		Where("user_id = ? AND channel = ? AND type = ?", pref.UserID, pref.Channel, pref.Type).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new preference
			if err := ps.db.WithContext(ctx).Create(pref).Error; err != nil {
				return fmt.Errorf("failed to create preference: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to check existing preference: %w", err)
	}

	// Update existing preference
	pref.ID = existing.ID
	if err := ps.db.WithContext(ctx).Save(pref).Error; err != nil {
		return fmt.Errorf("failed to update preference: %w", err)
	}

	return nil
}

// CanSend checks if a notification can be sent based on user preferences
func (ps *PreferenceServiceImpl) CanSend(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (bool, string, error) {
	pref, err := ps.Get(ctx, userID, channel, notifType)
	if err != nil {
		return false, "", err
	}

	// Check if enabled
	if !pref.IsEnabled {
		return false, "notifications disabled by user", nil
	}

	// Check quiet hours
	if pref.QuietHoursStart != nil && pref.QuietHoursEnd != nil {
		now := time.Now()

		// Load timezone if specified
		var loc *time.Location
		if pref.Timezone != "" {
			loc, err = time.LoadLocation(pref.Timezone)
			if err != nil {
				// Fall back to UTC
				loc = time.UTC
			}
			now = now.In(loc)
		}

		currentTime := now.Format("15:04")
		startTime := pref.QuietHoursStart.Format("15:04")
		endTime := pref.QuietHoursEnd.Format("15:04")

		if startTime <= endTime {
			// Normal case: start < end
			if currentTime >= startTime && currentTime < endTime {
				return false, "quiet hours active", nil
			}
		} else {
			// Overnight case: start > end (e.g., 22:00 to 08:00)
			if currentTime >= startTime || currentTime < endTime {
				return false, "quiet hours active", nil
			}
		}
	}

	// Check frequency limits
	if pref.MaxPerDay > 0 {
		count, err := ps.countSentToday(ctx, userID, channel, notifType)
		if err != nil {
			return false, "", err
		}
		if count >= pref.MaxPerDay {
			return false, "daily limit reached", nil
		}
	}

	if pref.MaxPerWeek > 0 {
		count, err := ps.countSentThisWeek(ctx, userID, channel, notifType)
		if err != nil {
			return false, "", err
		}
		if count >= pref.MaxPerWeek {
			return false, "weekly limit reached", nil
		}
	}

	if pref.MaxPerMonth > 0 {
		count, err := ps.countSentThisMonth(ctx, userID, channel, notifType)
		if err != nil {
			return false, "", err
		}
		if count >= pref.MaxPerMonth {
			return false, "monthly limit reached", nil
		}
	}

	return true, "", nil
}

// List lists all preferences for a user
func (ps *PreferenceServiceImpl) List(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	var prefs []*NotificationPreference

	if err := ps.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&prefs).Error; err != nil {
		return nil, fmt.Errorf("failed to list preferences: %w", err)
	}

	return prefs, nil
}

// Helper methods

func (ps *PreferenceServiceImpl) countSentToday(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (int, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)

	var count int64
	err := ps.db.WithContext(ctx).Model(&Notification{}).
		Where("recipient_id = ? AND channel = ? AND type = ? AND sent_at >= ? AND status IN (?)",
			userID, channel, notifType, startOfDay, []NotificationStatus{StatusSent, StatusDelivered}).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count sent today: %w", err)
	}

	return int(count), nil
}

func (ps *PreferenceServiceImpl) countSentThisWeek(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (int, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	startOfWeek := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)

	var count int64
	err := ps.db.WithContext(ctx).Model(&Notification{}).
		Where("recipient_id = ? AND channel = ? AND type = ? AND sent_at >= ? AND status IN (?)",
			userID, channel, notifType, startOfWeek, []NotificationStatus{StatusSent, StatusDelivered}).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count sent this week: %w", err)
	}

	return int(count), nil
}

func (ps *PreferenceServiceImpl) countSentThisMonth(ctx context.Context, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := ps.db.WithContext(ctx).Model(&Notification{}).
		Where("recipient_id = ? AND channel = ? AND type = ? AND sent_at >= ? AND status IN (?)",
			userID, channel, notifType, startOfMonth, []NotificationStatus{StatusSent, StatusDelivered}).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count sent this month: %w", err)
	}

	return int(count), nil
}

// BulkSetPreferences sets multiple preferences at once
func (ps *PreferenceServiceImpl) BulkSetPreferences(ctx context.Context, userID uuid.UUID, prefs []*NotificationPreference) error {
	for _, pref := range prefs {
		pref.UserID = userID
		if err := ps.Set(ctx, pref); err != nil {
			return fmt.Errorf("failed to set preference: %w", err)
		}
	}

	return nil
}

// ResetToDefaults resets user preferences to defaults
func (ps *PreferenceServiceImpl) ResetToDefaults(ctx context.Context, userID uuid.UUID) error {
	// Delete all existing preferences
	if err := ps.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&NotificationPreference{}).Error; err != nil {
		return fmt.Errorf("failed to reset preferences: %w", err)
	}

	return nil
}

// DisableAll disables all notifications for a user
func (ps *PreferenceServiceImpl) DisableAll(ctx context.Context, userID uuid.UUID) error {
	// Update all preferences to disabled
	if err := ps.db.WithContext(ctx).
		Model(&NotificationPreference{}).
		Where("user_id = ?", userID).
		Update("is_enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable all notifications: %w", err)
	}

	return nil
}

// EnableAll enables all notifications for a user
func (ps *PreferenceServiceImpl) EnableAll(ctx context.Context, userID uuid.UUID) error {
	// Update all preferences to enabled
	if err := ps.db.WithContext(ctx).
		Model(&NotificationPreference{}).
		Where("user_id = ?", userID).
		Update("is_enabled", true).Error; err != nil {
		return fmt.Errorf("failed to enable all notifications: %w", err)
	}

	return nil
}
