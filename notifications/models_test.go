package notifications

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected JSONB
		wantErr  bool
	}{
		{
			name:     "valid JSON bytes",
			value:    []byte(`{"key": "value", "number": 123}`),
			expected: JSONB{"key": "value", "number": float64(123)},
			wantErr:  false,
		},
		{
			name:     "empty JSON",
			value:    []byte(`{}`),
			expected: JSONB{},
			wantErr:  false,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: JSONB{},
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			value:    []byte(`{invalid json}`),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "wrong type",
			value:    "string",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONB
			err := j.Scan(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, j)
			}
		})
	}
}

func TestJSONB_Value(t *testing.T) {
	tests := []struct {
		name     string
		jsonb    JSONB
		expected string
	}{
		{
			name:     "simple object",
			jsonb:    JSONB{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "nested object",
			jsonb:    JSONB{"outer": map[string]interface{}{"inner": "value"}},
			expected: `{"outer":{"inner":"value"}}`,
		},
		{
			name:     "with array",
			jsonb:    JSONB{"items": []interface{}{"a", "b", "c"}},
			expected: `{"items":["a","b","c"]}`,
		},
		{
			name:     "empty object",
			jsonb:    JSONB{},
			expected: `{}`,
		},
		{
			name:     "nil JSONB",
			jsonb:    nil,
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.jsonb.Value()
			require.NoError(t, err)

			bytes, ok := value.([]byte)
			require.True(t, ok)
			assert.JSONEq(t, tt.expected, string(bytes))
		})
	}
}

func TestNotificationChannel_Constants(t *testing.T) {
	channels := []NotificationChannel{
		ChannelEmail,
		ChannelSMS,
		ChannelFirebase,
		ChannelPush,
		ChannelWebhook,
		ChannelInApp,
	}

	// Verify all channels are unique
	seen := make(map[NotificationChannel]bool)
	for _, ch := range channels {
		assert.False(t, seen[ch], "Duplicate channel: %s", ch)
		seen[ch] = true
		assert.NotEmpty(t, string(ch), "Channel should not be empty")
	}
}

func TestNotificationStatus_Constants(t *testing.T) {
	statuses := []NotificationStatus{
		StatusPending,
		StatusQueued,
		StatusSending,
		StatusSent,
		StatusDelivered,
		StatusFailed,
		StatusRetrying,
		StatusCancelled,
		StatusRead,
		StatusArchived,
	}

	// Verify all statuses are unique
	seen := make(map[NotificationStatus]bool)
	for _, st := range statuses {
		assert.False(t, seen[st], "Duplicate status: %s", st)
		seen[st] = true
		assert.NotEmpty(t, string(st), "Status should not be empty")
	}
}

func TestNotificationPriority_Constants(t *testing.T) {
	priorities := []NotificationPriority{
		PriorityLow,
		PriorityNormal,
		PriorityHigh,
		PriorityCritical,
	}

	// Verify all priorities are unique
	seen := make(map[NotificationPriority]bool)
	for _, p := range priorities {
		assert.False(t, seen[p], "Duplicate priority: %s", p)
		seen[p] = true
		assert.NotEmpty(t, string(p), "Priority should not be empty")
	}
}

func TestNotificationType_Constants(t *testing.T) {
	types := []NotificationType{
		TypeTransactional,
		TypeMarketing,
		TypeAlert,
		TypeReminder,
		TypeSystem,
	}

	// Verify all types are unique
	seen := make(map[NotificationType]bool)
	for _, t := range types {
		assert.False(t, seen[t], "Duplicate type: %s", t)
		seen[t] = true
		assert.NotEmpty(t, string(t), "Type should not be empty")
	}
}

func TestNotification_TableName(t *testing.T) {
	notif := Notification{}
	assert.Equal(t, "notifications", notif.TableName())
}

func TestNotificationTemplate_TableName(t *testing.T) {
	template := NotificationTemplate{}
	assert.Equal(t, "notification_templates", template.TableName())
}

func TestNotificationAudit_TableName(t *testing.T) {
	audit := NotificationAudit{}
	assert.Equal(t, "notification_audits", audit.TableName())
}

func TestNotificationPreference_TableName(t *testing.T) {
	pref := NotificationPreference{}
	assert.Equal(t, "notification_preferences", pref.TableName())
}

func TestNotificationBatch_TableName(t *testing.T) {
	batch := NotificationBatch{}
	assert.Equal(t, "notification_batches", batch.TableName())
}

func TestNotificationProvider_TableName(t *testing.T) {
	provider := NotificationProvider{}
	assert.Equal(t, "notification_providers", provider.TableName())
}

func TestJSONB_MarshalUnmarshal(t *testing.T) {
	original := JSONB{
		"string": "value",
		"number": float64(123),
		"bool":   true,
		"nested": map[string]interface{}{
			"key": "value",
		},
		"array": []interface{}{"a", "b", "c"},
	}

	// Marshal to JSON
	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var result JSONB
	err = json.Unmarshal(bytes, &result)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original["string"], result["string"])
	assert.Equal(t, original["number"], result["number"])
	assert.Equal(t, original["bool"], result["bool"])

	// Check nested
	originalNested := original["nested"].(map[string]interface{})
	resultNested := result["nested"].(map[string]interface{})
	assert.Equal(t, originalNested["key"], resultNested["key"])

	// Check array
	originalArray := original["array"].([]interface{})
	resultArray := result["array"].([]interface{})
	assert.Equal(t, len(originalArray), len(resultArray))
}

func TestNotificationModel_Fields(t *testing.T) {
	notif := Notification{
		Type:     TypeTransactional,
		Channel:  ChannelEmail,
		Priority: PriorityNormal,
		Status:   StatusPending,
		Subject:  "Test",
		Content:  "Content",
	}

	assert.Equal(t, TypeTransactional, notif.Type)
	assert.Equal(t, ChannelEmail, notif.Channel)
	assert.Equal(t, PriorityNormal, notif.Priority)
	assert.Equal(t, StatusPending, notif.Status)
	assert.Equal(t, "Test", notif.Subject)
	assert.Equal(t, "Content", notif.Content)
}

func TestNotificationTemplate_Fields(t *testing.T) {
	template := NotificationTemplate{
		Name:            "test_template",
		Type:            TypeTransactional,
		Channel:         ChannelEmail,
		Subject:         "{{.Subject}}",
		Content:         "{{.Content}}",
		DefaultPriority: PriorityNormal,
		IsActive:        true,
		Version:         1,
	}

	assert.Equal(t, "test_template", template.Name)
	assert.Equal(t, TypeTransactional, template.Type)
	assert.Equal(t, ChannelEmail, template.Channel)
	assert.True(t, template.IsActive)
	assert.Equal(t, 1, template.Version)
}

func TestNotificationAudit_Fields(t *testing.T) {
	audit := NotificationAudit{
		Status:    StatusSent,
		Event:     "sent",
		Message:   "Notification sent",
		IsError:   false,
		ActorType: "system",
	}

	assert.Equal(t, StatusSent, audit.Status)
	assert.Equal(t, "sent", audit.Event)
	assert.Equal(t, "Notification sent", audit.Message)
	assert.False(t, audit.IsError)
	assert.Equal(t, "system", audit.ActorType)
}

func TestNotificationPreference_Fields(t *testing.T) {
	pref := NotificationPreference{
		Channel:    ChannelEmail,
		Type:       TypeMarketing,
		IsEnabled:  true,
		MaxPerDay:  10,
		MaxPerWeek: 50,
	}

	assert.Equal(t, ChannelEmail, pref.Channel)
	assert.Equal(t, TypeMarketing, pref.Type)
	assert.True(t, pref.IsEnabled)
	assert.Equal(t, 10, pref.MaxPerDay)
	assert.Equal(t, 50, pref.MaxPerWeek)
}
