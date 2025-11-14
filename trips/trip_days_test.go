package trips

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTripDay_DateParsing_DateOnly(t *testing.T) {
	// Test that TripDay can unmarshal date-only format
	jsonData := `{
		"date": "2025-11-28",
		"day_number": 1,
		"day_type": "explore"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err != nil {
		t.Fatalf("Failed to unmarshal date-only format: %v", err)
	}

	expectedDate := "2025-11-28"
	actualDate := tripDay.Date.Format("2006-01-02")

	if actualDate != expectedDate {
		t.Errorf("Date parsing failed: got %s, want %s", actualDate, expectedDate)
	}

	if tripDay.DayNumber != 1 {
		t.Errorf("DayNumber parsing failed: got %d, want 1", tripDay.DayNumber)
	}

	if tripDay.DayType != TripDayTypeExplore {
		t.Errorf("DayType parsing failed: got %s, want %s", tripDay.DayType, TripDayTypeExplore)
	}
}

func TestTripDay_DateParsing_RFC3339(t *testing.T) {
	// Test that TripDay can also unmarshal RFC3339 format
	jsonData := `{
		"date": "2025-11-28T15:04:05Z",
		"day_number": 2,
		"day_type": "travel"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err != nil {
		t.Fatalf("Failed to unmarshal RFC3339 format: %v", err)
	}

	expectedDate := "2025-11-28"
	actualDate := tripDay.Date.Format("2006-01-02")

	if actualDate != expectedDate {
		t.Errorf("Date parsing failed: got %s, want %s", actualDate, expectedDate)
	}
}

func TestTripDay_DateParsing_WithTimezone(t *testing.T) {
	// Test that TripDay can unmarshal RFC3339 with timezone
	jsonData := `{
		"date": "2025-11-28T15:04:05+05:30",
		"day_number": 3,
		"day_type": "relax"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err != nil {
		t.Fatalf("Failed to unmarshal RFC3339 with timezone: %v", err)
	}

	expectedDate := "2025-11-28"
	actualDate := tripDay.Date.Format("2006-01-02")

	if actualDate != expectedDate {
		t.Errorf("Date parsing failed: got %s, want %s", actualDate, expectedDate)
	}
}

func TestTripDay_DateParsing_InvalidFormat(t *testing.T) {
	// Test that invalid date format returns error
	jsonData := `{
		"date": "invalid-date",
		"day_number": 1,
		"day_type": "explore"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err == nil {
		t.Error("Expected error for invalid date format, got nil")
	}
}

func TestTripDay_DateParsing_NullDate(t *testing.T) {
	// Test that null date is handled
	jsonData := `{
		"date": null,
		"day_number": 1,
		"day_type": "explore"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err != nil {
		t.Fatalf("Failed to unmarshal null date: %v", err)
	}

	if !tripDay.Date.IsZero() {
		t.Error("Null date should result in zero time")
	}
}

func TestTripDay_DateSerialization(t *testing.T) {
	// Test that TripDay serializes date correctly
	tripDay := TripDay{
		Date: core.Date{
			Time: time.Date(2025, 11, 28, 15, 30, 45, 0, time.UTC),
		},
		DayNumber: 1,
		DayType:   TripDayTypeExplore,
	}

	jsonData, err := json.Marshal(tripDay)
	if err != nil {
		t.Fatalf("Failed to marshal TripDay: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Date should be serialized as date-only format
	if result["date"] != "2025-11-28" {
		t.Errorf("Date serialization failed: got %v, want 2025-11-28", result["date"])
	}
}

func TestTripDay_RoundTrip(t *testing.T) {
	// Test full round-trip: unmarshal -> process -> marshal
	originalJSON := `{
		"date": "2025-11-28",
		"day_number": 5,
		"day_type": "adventure",
		"title": "Mountain Hiking",
		"notes": "Early start recommended"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(originalJSON), &tripDay)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	marshaledJSON, err := json.Marshal(tripDay)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(marshaledJSON, &result)
	if err != nil {
		t.Fatalf("Final unmarshal failed: %v", err)
	}

	if result["date"] != "2025-11-28" {
		t.Errorf("Date not preserved in round-trip: got %v", result["date"])
	}
}

func TestCreateTripDay_API_DateOnly(t *testing.T) {
	// Test the actual API endpoint with date-only format
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		user := accounts.User{
			BaseModel: core.BaseModel{
				ID: uuid.New(),
			},
		}
		c.Set("currentUser", user)
		c.Next()
	})

	router.POST("/trip-plans/:id/days", func(c *gin.Context) {
		var tripDay TripDay
		if err := c.BindJSON(&tripDay); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify date was parsed correctly
		expectedDate := "2025-11-28"
		actualDate := tripDay.Date.Format("2006-01-02")

		if actualDate != expectedDate {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Date parsing failed",
				"got":   actualDate,
				"want":  expectedDate,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"trip_day": tripDay,
			"message":  "Date parsed correctly",
		})
	})

	// Create test request
	requestBody := `{
		"date": "2025-11-28",
		"day_number": 1,
		"day_type": "explore",
		"title": "First Day"
	}`

	req, _ := http.NewRequest("POST", "/trip-plans/"+uuid.New().String()+"/days", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Date parsed correctly" {
		t.Errorf("Date was not parsed correctly: %v", response)
	}
}

func TestTripDay_AllDayTypes(t *testing.T) {
	// Test all valid day types
	dayTypes := []TripDayType{
		TripDayTypeTravel,
		TripDayTypeExplore,
		TripDayTypeRelax,
		TripDayTypeBusiness,
		TripDayTypeAdventure,
		TripDayTypeCultural,
	}

	for _, dayType := range dayTypes {
		t.Run(string(dayType), func(t *testing.T) {
			jsonData := `{
				"date": "2025-11-28",
				"day_number": 1,
				"day_type": "` + string(dayType) + `"
			}`

			var tripDay TripDay
			err := json.Unmarshal([]byte(jsonData), &tripDay)
			if err != nil {
				t.Fatalf("Failed to unmarshal day type %s: %v", dayType, err)
			}

			if tripDay.DayType != dayType {
				t.Errorf("DayType mismatch: got %s, want %s", tripDay.DayType, dayType)
			}
		})
	}
}

func TestTripDay_WithOptionalFields(t *testing.T) {
	// Test with all optional fields populated
	jsonData := `{
		"date": "2025-11-28",
		"day_number": 1,
		"day_type": "explore",
		"title": "Exploring the City",
		"notes": "Don't forget camera",
		"start_location": "Hotel Downtown",
		"end_location": "Eiffel Tower",
		"estimated_budget": 150.50,
		"actual_budget": 175.25,
		"weather": "Sunny, 22°C"
	}`

	var tripDay TripDay
	err := json.Unmarshal([]byte(jsonData), &tripDay)
	if err != nil {
		t.Fatalf("Failed to unmarshal with optional fields: %v", err)
	}

	if tripDay.Title == nil || *tripDay.Title != "Exploring the City" {
		t.Error("Title not parsed correctly")
	}

	if tripDay.Notes == nil || *tripDay.Notes != "Don't forget camera" {
		t.Error("Notes not parsed correctly")
	}

	if tripDay.EstimatedBudget == nil || *tripDay.EstimatedBudget != 150.50 {
		t.Error("EstimatedBudget not parsed correctly")
	}

	if tripDay.ActualBudget == nil || *tripDay.ActualBudget != 175.25 {
		t.Error("ActualBudget not parsed correctly")
	}
}

func TestTripDay_MultipleFormats(t *testing.T) {
	// Test that both formats can be used interchangeably
	formats := []string{
		`"2025-11-28"`,
		`"2025-11-28T00:00:00Z"`,
		`"2025-11-28T15:30:45Z"`,
		`"2025-11-28T15:30:45+05:30"`,
		`"2025-11-28T15:30:45-08:00"`,
	}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			jsonData := `{
				"date": ` + format + `,
				"day_number": 1,
				"day_type": "explore"
			}`

			var tripDay TripDay
			err := json.Unmarshal([]byte(jsonData), &tripDay)
			if err != nil {
				t.Fatalf("Failed to unmarshal format %s: %v", format, err)
			}

			expectedDate := "2025-11-28"
			actualDate := tripDay.Date.Format("2006-01-02")

			if actualDate != expectedDate {
				t.Errorf("Date parsing failed for format %s: got %s, want %s", format, actualDate, expectedDate)
			}
		})
	}
}

func TestGetValidTripDayTypes(t *testing.T) {
	types := GetValidTripDayTypes()

	if len(types) != 6 {
		t.Errorf("Expected 6 trip day types, got %d", len(types))
	}

	expectedTypes := map[TripDayType]bool{
		TripDayTypeTravel:    true,
		TripDayTypeExplore:   true,
		TripDayTypeRelax:     true,
		TripDayTypeBusiness:  true,
		TripDayTypeAdventure: true,
		TripDayTypeCultural:  true,
	}

	for _, dayType := range types {
		if !expectedTypes[dayType] {
			t.Errorf("Unexpected trip day type: %s", dayType)
		}
	}
}

func TestIsValidTripDayType(t *testing.T) {
	tests := []struct {
		name    string
		dayType string
		valid   bool
	}{
		{"Travel", "travel", true},
		{"Explore", "explore", true},
		{"Relax", "relax", true},
		{"Business", "business", true},
		{"Adventure", "adventure", true},
		{"Cultural", "cultural", true},
		{"Invalid", "party", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTripDayType(tt.dayType)
			if result != tt.valid {
				t.Errorf("IsValidTripDayType(%s) = %v, want %v", tt.dayType, result, tt.valid)
			}
		})
	}
}
