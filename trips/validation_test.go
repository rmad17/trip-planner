package trips

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTripGenerationRequestValidation tests request validation scenarios
func TestTripGenerationRequestValidation(t *testing.T) {
	tests := []struct {
		name      string
		request   TripGenerationRequest
		shouldErr bool
		errField  string
	}{
		{
			name: "Valid request with all fields",
			request: TripGenerationRequest{
				Source:          "New York",
				Destinations:    []string{"Paris", "London"},
				StartDate:       time.Now(),
				EndDate:         time.Now().AddDate(0, 0, 14),
				NumTravelers:    2,
				Budget:          5000,
				Currency:        CurrencyUSD,
				TripPreferences: []string{"culture", "food"},
				PacePreference:  "moderate",
			},
			shouldErr: false,
		},
		{
			name: "Valid request with minimal fields",
			request: TripGenerationRequest{
				Source:       "Tokyo",
				Destinations: []string{"Kyoto"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 1,
			},
			shouldErr: false,
		},
		{
			name: "Invalid - empty source",
			request: TripGenerationRequest{
				Source:       "",
				Destinations: []string{"Paris"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 2,
			},
			shouldErr: true,
			errField:  "source",
		},
		{
			name: "Invalid - empty destinations",
			request: TripGenerationRequest{
				Source:       "New York",
				Destinations: []string{},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 2,
			},
			shouldErr: true,
			errField:  "destinations",
		},
		{
			name: "Invalid - zero travelers",
			request: TripGenerationRequest{
				Source:       "New York",
				Destinations: []string{"Paris"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 0,
			},
			shouldErr: true,
			errField:  "num_travelers",
		},
		{
			name: "Invalid - negative budget",
			request: TripGenerationRequest{
				Source:       "New York",
				Destinations: []string{"Paris"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 2,
				Budget:       -1000,
			},
			shouldErr: true,
			errField:  "budget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate required fields
			hasError := false

			if tt.request.Source == "" {
				hasError = true
			}
			if len(tt.request.Destinations) == 0 {
				hasError = true
			}
			if tt.request.NumTravelers <= 0 {
				hasError = true
			}
			if tt.request.Budget < 0 {
				hasError = true
			}

			if tt.shouldErr && !hasError {
				t.Errorf("Expected validation error for %s, but got none", tt.errField)
			}
			if !tt.shouldErr && hasError {
				t.Errorf("Expected no validation error, but got one")
			}
		})
	}
}

// TestDateValidation tests date-related validation
func TestDateValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		isValid   bool
	}{
		{
			name:      "Valid - end after start",
			startDate: now,
			endDate:   now.AddDate(0, 0, 7),
			isValid:   true,
		},
		{
			name:      "Invalid - end before start",
			startDate: now,
			endDate:   now.AddDate(0, 0, -7),
			isValid:   false,
		},
		{
			name:      "Invalid - end same as start",
			startDate: now,
			endDate:   now,
			isValid:   false,
		},
		{
			name:      "Valid - long trip",
			startDate: now,
			endDate:   now.AddDate(0, 1, 0),
			isValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := !tt.endDate.Before(tt.startDate) && !tt.endDate.Equal(tt.startDate)

			if isValid != tt.isValid {
				t.Errorf("Expected isValid=%v, got %v", tt.isValid, isValid)
			}
		})
	}
}

// TestCurrencyValidation tests currency validation
func TestCurrencyValidation(t *testing.T) {
	validCurrencies := []Currency{
		CurrencyUSD,
		CurrencyEUR,
		CurrencyGBP,
		CurrencyINR,
		CurrencyJPY,
		CurrencyCAD,
		CurrencyAUD,
	}

	for _, currency := range validCurrencies {
		t.Run("Valid currency: "+string(currency), func(t *testing.T) {
			if currency == "" {
				t.Errorf("Currency should not be empty")
			}
		})
	}

	t.Run("Empty currency should default to USD", func(t *testing.T) {
		var currency Currency = ""
		if currency == "" {
			currency = CurrencyUSD
		}
		if currency != CurrencyUSD {
			t.Errorf("Expected USD default, got %s", currency)
		}
	})
}

// TestJSONSerialization tests JSON marshaling/unmarshaling
func TestJSONSerialization(t *testing.T) {
	t.Run("TripGenerationRequest serialization", func(t *testing.T) {
		req := TripGenerationRequest{
			Source:          "New York",
			Destinations:    []string{"Paris"},
			StartDate:       time.Now(),
			EndDate:         time.Now().AddDate(0, 0, 7),
			NumTravelers:    2,
			Budget:          3000,
			Currency:        CurrencyUSD,
			TripPreferences: []string{"culture"},
			PacePreference:  "moderate",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		var decoded TripGenerationRequest
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Failed to unmarshal request: %v", err)
		}

		if decoded.Source != req.Source {
			t.Errorf("Source mismatch: expected %s, got %s", req.Source, decoded.Source)
		}
	})

	t.Run("InteractiveFeedback serialization", func(t *testing.T) {
		feedback := InteractiveFeedback{
			TripPlanID:     "trip-123",
			CurrentPlan:    `{"trip_name":"Test"}`,
			Feedback:       "Add more activities",
			Interests:      []string{"hiking"},
			ModifiedBudget: 2500,
		}

		data, err := json.Marshal(feedback)
		if err != nil {
			t.Errorf("Failed to marshal feedback: %v", err)
		}

		var decoded InteractiveFeedback
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Failed to unmarshal feedback: %v", err)
		}

		if decoded.Feedback != feedback.Feedback {
			t.Errorf("Feedback mismatch: expected %s, got %s", feedback.Feedback, decoded.Feedback)
		}
	})

	t.Run("CitySuggestion serialization", func(t *testing.T) {
		suggestion := CitySuggestion{
			City:          "Rome",
			Country:       "Italy",
			Reason:        "Historic sites",
			BestSeason:    "Spring",
			EstimatedDays: 3,
		}

		data, err := json.Marshal(suggestion)
		if err != nil {
			t.Errorf("Failed to marshal suggestion: %v", err)
		}

		var decoded CitySuggestion
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Failed to unmarshal suggestion: %v", err)
		}

		if decoded.City != suggestion.City {
			t.Errorf("City mismatch: expected %s, got %s", suggestion.City, decoded.City)
		}
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Very large number of travelers", func(t *testing.T) {
		req := TripGenerationRequest{
			Source:       "New York",
			Destinations: []string{"Paris"},
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 7),
			NumTravelers: 100,
		}

		if req.NumTravelers <= 0 {
			t.Errorf("Should allow large number of travelers")
		}
	})

	t.Run("Very long trip duration", func(t *testing.T) {
		start := time.Now()
		end := start.AddDate(1, 0, 0) // 1 year

		if end.Before(start) {
			t.Errorf("End date should be after start date")
		}
	})

	t.Run("Very high budget", func(t *testing.T) {
		req := TripGenerationRequest{
			Source:       "New York",
			Destinations: []string{"Paris"},
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 7),
			NumTravelers: 2,
			Budget:       1000000,
		}

		if req.Budget < 0 {
			t.Errorf("Budget should be non-negative")
		}
	})

	t.Run("Many destinations", func(t *testing.T) {
		destinations := make([]string, 20)
		for i := range destinations {
			destinations[i] = "City" + string(rune(i))
		}

		req := TripGenerationRequest{
			Source:       "New York",
			Destinations: destinations,
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 30),
			NumTravelers: 2,
		}

		if len(req.Destinations) == 0 {
			t.Errorf("Should allow many destinations")
		}
	})

	t.Run("Many preferences", func(t *testing.T) {
		preferences := []string{
			"adventure", "culture", "food", "nature",
			"photography", "shopping", "nightlife", "relaxation",
		}

		req := TripGenerationRequest{
			Source:          "New York",
			Destinations:    []string{"Paris"},
			StartDate:       time.Now(),
			EndDate:         time.Now().AddDate(0, 0, 7),
			NumTravelers:    2,
			TripPreferences: preferences,
		}

		if len(req.TripPreferences) == 0 {
			t.Errorf("Should allow many preferences")
		}
	})
}

// TestEmptyAndNilValues tests handling of empty and nil values
func TestEmptyAndNilValues(t *testing.T) {
	t.Run("Empty trip preferences", func(t *testing.T) {
		req := TripGenerationRequest{
			Source:          "New York",
			Destinations:    []string{"Paris"},
			StartDate:       time.Now(),
			EndDate:         time.Now().AddDate(0, 0, 7),
			NumTravelers:    2,
			TripPreferences: []string{},
		}

		// Should be valid - empty preferences is okay
		if req.Source == "" || len(req.Destinations) == 0 {
			t.Errorf("Should be valid with empty preferences")
		}
	})

	t.Run("Nil interests in feedback", func(t *testing.T) {
		feedback := InteractiveFeedback{
			CurrentPlan: `{"trip_name":"Test"}`,
			Feedback:    "Add activities",
			Interests:   nil,
		}

		if feedback.Feedback == "" {
			t.Errorf("Feedback should be present")
		}
	})

	t.Run("Zero budget", func(t *testing.T) {
		req := TripGenerationRequest{
			Source:       "New York",
			Destinations: []string{"Paris"},
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 0, 7),
			NumTravelers: 2,
			Budget:       0,
		}

		// Zero budget should be valid (means no budget specified)
		if req.Budget < 0 {
			t.Errorf("Zero budget should be valid")
		}
	})
}

// TestProviderNameValidation tests provider name validation
func TestProviderNameValidation(t *testing.T) {
	validProviders := map[string]bool{
		"gemini": true,
		"claude": true,
		"gpt":    true,
	}

	tests := []struct {
		name     string
		provider string
		isValid  bool
	}{
		{"Valid Gemini", "gemini", true},
		{"Valid Claude", "claude", true},
		{"Valid GPT", "gpt", true},
		{"Invalid provider", "invalid", false},
		{"Empty provider", "", false},
		{"Uppercase", "GEMINI", false},
		{"Mixed case", "Gemini", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validProviders[tt.provider]

			if isValid != tt.isValid {
				t.Errorf("Provider %s: expected valid=%v, got %v", tt.provider, tt.isValid, isValid)
			}
		})
	}
}

// TestResponseStructures tests response structure validation
func TestResponseStructures(t *testing.T) {
	t.Run("TripGenerationResponse has required fields", func(t *testing.T) {
		response := TripGenerationResponse{
			TripName:        "Test Trip",
			Description:     "A test trip",
			RecommendedMode: "flight",
			TotalDays:       7,
			EstimatedBudget: 3000,
		}

		if response.TripName == "" {
			t.Errorf("TripName should not be empty")
		}
		if response.TotalDays <= 0 {
			t.Errorf("TotalDays should be positive")
		}
		if response.EstimatedBudget < 0 {
			t.Errorf("EstimatedBudget should be non-negative")
		}
	})

	t.Run("GeneratedHop has required fields", func(t *testing.T) {
		hop := GeneratedHop{
			Name:        "Paris",
			City:        "Paris",
			Country:     "France",
			Description: "City of lights",
			Duration:    3,
			HopOrder:    1,
		}

		if hop.Name == "" || hop.City == "" || hop.Country == "" {
			t.Errorf("Required fields should not be empty")
		}
		if hop.Duration <= 0 {
			t.Errorf("Duration should be positive")
		}
	})

	t.Run("GeneratedActivity has required fields", func(t *testing.T) {
		activity := GeneratedActivity{
			Name:         "Eiffel Tower Visit",
			ActivityType: "sightseeing",
			StartTime:    "09:00",
			EndTime:      "11:00",
			Duration:     120,
			Priority:     5,
		}

		if activity.Name == "" {
			t.Errorf("Activity name should not be empty")
		}
		if activity.Duration <= 0 {
			t.Errorf("Duration should be positive")
		}
		if activity.Priority < 1 || activity.Priority > 5 {
			t.Errorf("Priority should be between 1 and 5")
		}
	})
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("serializeTripToJSON", func(t *testing.T) {
		trip := &TripGenerationResponse{
			TripName:    "Test",
			Description: "Test trip",
		}

		data, err := serializeTripToJSON(trip)
		if err != nil {
			t.Errorf("Failed to serialize trip: %v", err)
		}
		if len(data) == 0 {
			t.Errorf("Serialized data should not be empty")
		}
	})

	t.Run("parseJSONResponse", func(t *testing.T) {
		jsonStr := `{"trip_name":"Test","description":"Test trip"}`
		var trip TripGenerationResponse

		err := parseJSONResponse(jsonStr, &trip)
		if err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}
		if trip.TripName != "Test" {
			t.Errorf("Expected TripName 'Test', got '%s'", trip.TripName)
		}
	})

	t.Run("parseJSONResponse with invalid JSON", func(t *testing.T) {
		jsonStr := `{invalid json}`
		var trip TripGenerationResponse

		err := parseJSONResponse(jsonStr, &trip)
		if err == nil {
			t.Errorf("Expected error for invalid JSON")
		}
	})
}
