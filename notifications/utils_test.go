package notifications

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimePtr(t *testing.T) {
	now := time.Now()
	ptr := TimePtr(now)

	assert.NotNil(t, ptr)
	assert.Equal(t, now, *ptr)

	// Verify it's actually a pointer
	*ptr = now.Add(time.Hour)
	assert.NotEqual(t, now, *ptr)
}

func TestStringPtr(t *testing.T) {
	str := "test string"
	ptr := StringPtr(str)

	assert.NotNil(t, ptr)
	assert.Equal(t, str, *ptr)

	// Verify it's actually a pointer
	*ptr = "modified"
	assert.NotEqual(t, str, *ptr)
}

func TestIntPtr(t *testing.T) {
	num := 42
	ptr := IntPtr(num)

	assert.NotNil(t, ptr)
	assert.Equal(t, num, *ptr)

	// Verify it's actually a pointer
	*ptr = 100
	assert.NotEqual(t, num, *ptr)
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"true value", true},
		{"false value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := BoolPtr(tt.value)

			assert.NotNil(t, ptr)
			assert.Equal(t, tt.value, *ptr)

			// Verify it's actually a pointer
			*ptr = !tt.value
			assert.NotEqual(t, tt.value, *ptr)
		})
	}
}

func TestGetModels(t *testing.T) {
	models := GetModels()

	// Should return all 6 notification models
	expectedCount := 6
	assert.Len(t, models, expectedCount, "Should return %d models", expectedCount)

	// Verify model types
	expectedTypes := []interface{}{
		&Notification{},
		&NotificationTemplate{},
		&NotificationAudit{},
		&NotificationPreference{},
		&NotificationBatch{},
		&NotificationProvider{},
	}

	for i, expected := range expectedTypes {
		assert.IsType(t, expected, models[i], "Model %d should be correct type", i)
	}
}

func TestGetModels_NonNil(t *testing.T) {
	models := GetModels()

	// Verify all models are non-nil
	for i, model := range models {
		assert.NotNil(t, model, "Model %d should not be nil", i)
	}
}

func TestGetModels_Uniqueness(t *testing.T) {
	models := GetModels()

	// Verify all models are unique types
	types := make(map[string]bool)
	for _, model := range models {
		typeStr := fmt.Sprintf("%T", model)
		assert.False(t, types[typeStr], "Duplicate model type found")
		types[typeStr] = true
	}
}
