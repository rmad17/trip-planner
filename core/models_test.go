package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDate_UnmarshalJSON_DateOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid date-only format",
			input:    `"2025-11-28"`,
			expected: "2025-11-28",
			wantErr:  false,
		},
		{
			name:     "Valid RFC3339 format",
			input:    `"2025-11-28T15:04:05Z"`,
			expected: "2025-11-28",
			wantErr:  false,
		},
		{
			name:     "Valid RFC3339 with timezone",
			input:    `"2025-11-28T15:04:05+05:30"`,
			expected: "2025-11-28",
			wantErr:  false,
		},
		{
			name:     "Null value",
			input:    `null`,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Empty string",
			input:    `""`,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			input:    `"invalid-date"`,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid year",
			input:    `"99-11-28"`,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tt.input), &d)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.expected != "" {
				got := d.Time.Format("2006-01-02")
				if got != tt.expected {
					t.Errorf("UnmarshalJSON() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		date     Date
		expected string
	}{
		{
			name:     "Valid date",
			date:     Date{Time: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)},
			expected: `"2025-11-28"`,
		},
		{
			name:     "Zero date",
			date:     Date{Time: time.Time{}},
			expected: `null`,
		},
		{
			name:     "Date with time component",
			date:     Date{Time: time.Date(2025, 11, 28, 15, 30, 45, 0, time.UTC)},
			expected: `"2025-11-28"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.date)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}

			if string(got) != tt.expected {
				t.Errorf("MarshalJSON() got = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestDate_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid time.Time",
			input:    time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
			expected: "2025-11-28",
			wantErr:  false,
		},
		{
			name:     "Nil value",
			input:    nil,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Invalid type",
			input:    "2025-11-28",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid type - int",
			input:    123456,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := d.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.expected != "" {
				got := d.Time.Format("2006-01-02")
				if got != tt.expected {
					t.Errorf("Scan() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestDate_Value(t *testing.T) {
	tests := []struct {
		name     string
		date     Date
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "Valid date",
			date:     Date{Time: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)},
			expected: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Zero date",
			date:     Date{Time: time.Time{}},
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.Value()

			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got != tt.expected {
					t.Errorf("Value() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestDate_RoundTrip(t *testing.T) {
	// Test JSON round-trip
	t.Run("JSON round-trip", func(t *testing.T) {
		original := Date{Time: time.Date(2025, 11, 28, 15, 30, 45, 0, time.UTC)}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// Unmarshal back
		var restored Date
		err = json.Unmarshal(jsonData, &restored)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Compare dates (ignoring time component)
		originalDate := original.Time.Format("2006-01-02")
		restoredDate := restored.Time.Format("2006-01-02")

		if originalDate != restoredDate {
			t.Errorf("Round-trip failed: got %v, want %v", restoredDate, originalDate)
		}
	})

	// Test database round-trip
	t.Run("Database round-trip", func(t *testing.T) {
		original := Date{Time: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)}

		// Convert to driver value
		val, err := original.Value()
		if err != nil {
			t.Fatalf("Value() failed: %v", err)
		}

		// Scan back
		var restored Date
		err = restored.Scan(val)
		if err != nil {
			t.Fatalf("Scan() failed: %v", err)
		}

		if !original.Time.Equal(restored.Time) {
			t.Errorf("Round-trip failed: got %v, want %v", restored.Time, original.Time)
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("StringPtr", func(t *testing.T) {
		str := "test"
		ptr := StringPtr(str)
		if ptr == nil {
			t.Error("StringPtr returned nil")
		}
		if *ptr != str {
			t.Errorf("StringPtr got = %v, want %v", *ptr, str)
		}
	})

	t.Run("IntPtr", func(t *testing.T) {
		val := 42
		ptr := IntPtr(val)
		if ptr == nil {
			t.Error("IntPtr returned nil")
		}
		if *ptr != val {
			t.Errorf("IntPtr got = %v, want %v", *ptr, val)
		}
	})

	t.Run("Float64Ptr", func(t *testing.T) {
		val := 3.14
		ptr := Float64Ptr(val)
		if ptr == nil {
			t.Error("Float64Ptr returned nil")
		}
		if *ptr != val {
			t.Errorf("Float64Ptr got = %v, want %v", *ptr, val)
		}
	})
}
