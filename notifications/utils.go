package notifications

import (
	"time"
)

// TimePtr returns a pointer to the given time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}

// GetModels returns all notification models for database migration
func GetModels() []interface{} {
	return []interface{}{
		&Notification{},
		&NotificationTemplate{},
		&NotificationAudit{},
		&NotificationPreference{},
		&NotificationBatch{},
		&NotificationProvider{},
	}
}
