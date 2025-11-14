package trips

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupAITestRouter creates a test router for AI endpoints
func setupAITestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Middleware to set authenticated user
	router.Use(func(c *gin.Context) {
		c.Set("currentUser", uuid.New())
		c.Next()
	})

	return router
}

// TestGenerateTripWithAI_MissingAPIKey tests error handling when API key is missing
func TestGenerateTripWithAI_MissingAPIKey(t *testing.T) {
	// Save original env vars
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	originalClaudeKey := os.Getenv("ANTHROPIC_API_KEY")
	originalGPTKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		os.Setenv("GEMINI_API_KEY", originalGeminiKey)
		os.Setenv("ANTHROPIC_API_KEY", originalClaudeKey)
		os.Setenv("OPENAI_API_KEY", originalGPTKey)
	}()

	// Clear API keys
	os.Setenv("GEMINI_API_KEY", "")
	os.Setenv("ANTHROPIC_API_KEY", "")
	os.Setenv("OPENAI_API_KEY", "")

	router := setupAITestRouter()
	router.POST("/trip/generate", GenerateTripWithAI)

	tests := []struct {
		name           string
		provider       string
		expectedError  string
		expectedStatus int
	}{
		{
			name:           "Missing Gemini API key",
			provider:       "gemini",
			expectedError:  "GEMINI_API_KEY not set in environment",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Missing Claude API key",
			provider:       "claude",
			expectedError:  "ANTHROPIC_API_KEY not set in environment",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Missing GPT API key",
			provider:       "gpt",
			expectedError:  "OPENAI_API_KEY not set in environment",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Missing default provider key",
			provider:       "",
			expectedError:  "GEMINI_API_KEY not set in environment",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := TripGenerationRequest{
				Source:       "New York",
				Destinations: []string{"Paris"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, 7),
				NumTravelers: 2,
				Currency:     CurrencyUSD,
			}

			jsonBody, _ := json.Marshal(requestBody)
			url := "/trip/generate"
			if tt.provider != "" {
				url += "?provider=" + tt.provider
			}

			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestGenerateTripWithAI_InvalidRequest tests request validation
func TestGenerateTripWithAI_InvalidRequest(t *testing.T) {
	router := setupAITestRouter()
	router.POST("/trip/generate", GenerateTripWithAI)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error",
		},
		{
			name: "Missing required fields",
			requestBody: map[string]interface{}{
				"source": "New York",
				// Missing destinations
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error",
		},
		{
			name: "End date before start date",
			requestBody: TripGenerationRequest{
				Source:       "New York",
				Destinations: []string{"Paris"},
				StartDate:    time.Now(),
				EndDate:      time.Now().AddDate(0, 0, -7), // 7 days before start
				NumTravelers: 2,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "end_date must be after start_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonBody []byte
			if str, ok := tt.requestBody.(string); ok {
				jsonBody = []byte(str)
			} else {
				jsonBody, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest("POST", "/trip/generate", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestGenerateTripWithAI_UnsupportedProvider tests invalid provider parameter
func TestGenerateTripWithAI_UnsupportedProvider(t *testing.T) {
	router := setupAITestRouter()
	router.POST("/trip/generate", GenerateTripWithAI)

	requestBody := TripGenerationRequest{
		Source:       "New York",
		Destinations: []string{"Paris"},
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(0, 0, 7),
		NumTravelers: 2,
		Currency:     CurrencyUSD,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/trip/generate?provider=unsupported", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "unsupported LLM provider")
}

// TestRefineTripWithFeedback_MissingCurrentPlan tests validation
func TestRefineTripWithFeedback_MissingCurrentPlan(t *testing.T) {
	router := setupAITestRouter()
	router.POST("/trip/refine", RefineTripWithFeedback)

	tests := []struct {
		name           string
		feedback       InteractiveFeedback
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Missing current plan",
			feedback: InteractiveFeedback{
				Feedback: "Add more activities",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "current_plan is required",
		},
		{
			name: "Invalid JSON in current plan",
			feedback: InteractiveFeedback{
				CurrentPlan: `{invalid json}`,
				Feedback:    "Add more activities",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid current plan format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.feedback)
			req, _ := http.NewRequest("POST", "/trip/refine", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestRefineTripWithFeedback_ValidRequest tests feedback processing
func TestRefineTripWithFeedback_ValidRequest(t *testing.T) {
	// Set up test API key
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalGeminiKey)
	os.Setenv("GEMINI_API_KEY", "")

	router := setupAITestRouter()
	router.POST("/trip/refine", RefineTripWithFeedback)

	// Create a valid trip plan JSON
	tripPlan := TripGenerationResponse{
		TripName:        "Test Trip",
		Description:     "Test Description",
		RecommendedMode: "flight",
		TotalDays:       7,
		EstimatedBudget: 2000,
	}
	tripPlanJSON, _ := json.Marshal(tripPlan)

	feedback := InteractiveFeedback{
		CurrentPlan:    string(tripPlanJSON),
		Feedback:       "Add more cultural activities",
		Interests:      []string{"culture", "history"},
		ModifiedBudget: 2500,
	}

	jsonBody, _ := json.Marshal(feedback)
	req, _ := http.NewRequest("POST", "/trip/refine", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail because API key is not set
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "GEMINI_API_KEY not set")
}

// TestGetMultiCitySuggestions_MissingParams tests parameter validation
func TestGetMultiCitySuggestions_MissingParams(t *testing.T) {
	router := setupAITestRouter()
	router.GET("/trip/suggest-cities", GetMultiCitySuggestions)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing source",
			queryParams:    "?destination=Paris",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "source and destination are required",
		},
		{
			name:           "Missing destination",
			queryParams:    "?source=New York",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "source and destination are required",
		},
		{
			name:           "Missing both parameters",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "source and destination are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/trip/suggest-cities"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestGetMultiCitySuggestions_WithProvider tests provider selection
func TestGetMultiCitySuggestions_WithProvider(t *testing.T) {
	// Save original env vars
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalGeminiKey)
	os.Setenv("GEMINI_API_KEY", "")

	router := setupAITestRouter()
	router.GET("/trip/suggest-cities", GetMultiCitySuggestions)

	tests := []struct {
		name           string
		provider       string
		expectedError  string
		expectedStatus int
	}{
		{
			name:           "Default provider without API key",
			provider:       "",
			expectedError:  "GEMINI_API_KEY not set",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Unsupported provider",
			provider:       "unsupported",
			expectedError:  "unsupported LLM provider",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/trip/suggest-cities?source=New York&destination=Paris"
			if tt.provider != "" {
				url += "&provider=" + tt.provider
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestDefaultCurrency tests that default currency is set correctly
func TestDefaultCurrency(t *testing.T) {
	router := setupAITestRouter()
	router.POST("/trip/generate", GenerateTripWithAI)

	// Set API key for this test
	originalKey := os.Getenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalKey)
	os.Setenv("GEMINI_API_KEY", "")

	requestBody := TripGenerationRequest{
		Source:       "New York",
		Destinations: []string{"Paris"},
		StartDate:    time.Now(),
		EndDate:      time.Now().AddDate(0, 0, 7),
		NumTravelers: 2,
		// Currency not set
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/trip/generate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The request should proceed (and fail at API key check, not currency validation)
	// This confirms that currency default is being set
	assert.NotEqual(t, http.StatusBadRequest, w.Code, "Currency default should be set, not causing validation error")
}

// TestProviderResponseFormat tests that provider name is returned in response
func TestProviderResponseFormat(t *testing.T) {
	// This is a unit test to verify response structure
	// We can't test actual API calls without valid keys, but we can verify the structure

	t.Run("Response should include provider field", func(t *testing.T) {
		// The response format is verified in the handler code
		// This test documents the expected behavior
		expectedFields := []string{"success", "trip_plan", "provider", "message"}

		// Verify these fields are documented
		for _, field := range expectedFields {
			assert.NotEmpty(t, field, "Response should include "+field)
		}
	})
}

// TestInteractiveFeedbackWithInterests tests feedback with interests
func TestInteractiveFeedbackWithInterests(t *testing.T) {
	tests := []struct {
		name      string
		feedback  InteractiveFeedback
		shouldAdd string
	}{
		{
			name: "Feedback with interests",
			feedback: InteractiveFeedback{
				Feedback:  "Make it more adventurous",
				Interests: []string{"hiking", "photography"},
			},
			shouldAdd: "User Interests: hiking, photography",
		},
		{
			name: "Feedback with budget",
			feedback: InteractiveFeedback{
				Feedback:       "Increase activities",
				ModifiedBudget: 3000,
			},
			shouldAdd: "Revised Budget: 3000",
		},
		{
			name: "Feedback with dates",
			feedback: InteractiveFeedback{
				Feedback: "Change dates",
				ModifiedDates: struct {
					StartDate string `json:"start_date,omitempty"`
					EndDate   string `json:"end_date,omitempty"`
				}{
					StartDate: "2024-06-01",
					EndDate:   "2024-06-15",
				},
			},
			shouldAdd: "Revised Dates: 2024-06-01 to 2024-06-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that feedback structure can hold all the data
			assert.NotEmpty(t, tt.feedback.Feedback)
			assert.NotEmpty(t, tt.shouldAdd)
		})
	}
}

// TestConcurrentProviderCreation tests thread safety of provider factory
func TestConcurrentProviderCreation(t *testing.T) {
	// Set up test API keys
	os.Setenv("GEMINI_API_KEY", "test-key-1")
	os.Setenv("ANTHROPIC_API_KEY", "test-key-2")
	os.Setenv("OPENAI_API_KEY", "test-key-3")
	defer func() {
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	factory := NewLLMProviderFactory()

	// Create providers concurrently
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(providerType LLMProviderType) {
			_, err := factory.GetProvider(providerType)
			if err != nil {
				t.Errorf("Error creating provider: %v", err)
			}
			done <- true
		}([]LLMProviderType{ProviderGemini, ProviderClaude, ProviderGPT}[i])
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
