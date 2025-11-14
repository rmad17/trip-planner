package trips

import (
	"context"
	"os"
	"testing"
)

// TestNewLLMProviderFactory tests the factory creation
func TestNewLLMProviderFactory(t *testing.T) {
	// Save original env var
	originalProvider := os.Getenv("DEFAULT_LLM_PROVIDER")
	defer os.Setenv("DEFAULT_LLM_PROVIDER", originalProvider)

	tests := []struct {
		name            string
		envValue        string
		expectedDefault LLMProviderType
	}{
		{
			name:            "Default to Gemini when env not set",
			envValue:        "",
			expectedDefault: ProviderGemini,
		},
		{
			name:            "Use Claude from env",
			envValue:        "claude",
			expectedDefault: ProviderClaude,
		},
		{
			name:            "Use GPT from env",
			envValue:        "gpt",
			expectedDefault: ProviderGPT,
		},
		{
			name:            "Use Gemini from env",
			envValue:        "gemini",
			expectedDefault: ProviderGemini,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DEFAULT_LLM_PROVIDER", tt.envValue)
			factory := NewLLMProviderFactory()

			if factory.GetDefaultProviderType() != tt.expectedDefault {
				t.Errorf("Expected default provider %s, got %s", tt.expectedDefault, factory.GetDefaultProviderType())
			}
		})
	}
}

// TestGetProvider tests getting specific providers
func TestGetProvider(t *testing.T) {
	factory := NewLLMProviderFactory()

	tests := []struct {
		name          string
		providerType  LLMProviderType
		apiKeyEnv     string
		apiKeyValue   string
		shouldFail    bool
		expectedError string
	}{
		{
			name:          "Get Gemini provider with valid key",
			providerType:  ProviderGemini,
			apiKeyEnv:     "GEMINI_API_KEY",
			apiKeyValue:   "test-gemini-key",
			shouldFail:    false,
			expectedError: "",
		},
		{
			name:          "Get Claude provider with valid key",
			providerType:  ProviderClaude,
			apiKeyEnv:     "ANTHROPIC_API_KEY",
			apiKeyValue:   "test-claude-key",
			shouldFail:    false,
			expectedError: "",
		},
		{
			name:          "Get GPT provider with valid key",
			providerType:  ProviderGPT,
			apiKeyEnv:     "OPENAI_API_KEY",
			apiKeyValue:   "test-gpt-key",
			shouldFail:    false,
			expectedError: "",
		},
		{
			name:          "Fail when Gemini API key not set",
			providerType:  ProviderGemini,
			apiKeyEnv:     "GEMINI_API_KEY",
			apiKeyValue:   "",
			shouldFail:    true,
			expectedError: "GEMINI_API_KEY not set in environment",
		},
		{
			name:          "Fail when Claude API key not set",
			providerType:  ProviderClaude,
			apiKeyEnv:     "ANTHROPIC_API_KEY",
			apiKeyValue:   "",
			shouldFail:    true,
			expectedError: "ANTHROPIC_API_KEY not set in environment",
		},
		{
			name:          "Fail when GPT API key not set",
			providerType:  ProviderGPT,
			apiKeyEnv:     "OPENAI_API_KEY",
			apiKeyValue:   "",
			shouldFail:    true,
			expectedError: "OPENAI_API_KEY not set in environment",
		},
		{
			name:          "Fail with unsupported provider",
			providerType:  LLMProviderType("unsupported"),
			apiKeyEnv:     "",
			apiKeyValue:   "",
			shouldFail:    true,
			expectedError: "unsupported LLM provider: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.apiKeyEnv != "" {
				originalKey := os.Getenv(tt.apiKeyEnv)
				defer os.Setenv(tt.apiKeyEnv, originalKey)
				os.Setenv(tt.apiKeyEnv, tt.apiKeyValue)
			}

			provider, err := factory.GetProvider(tt.providerType)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if provider == nil {
					t.Errorf("Expected provider instance, got nil")
				}
			}
		})
	}
}

// TestGetDefaultProvider tests getting the default provider
func TestGetDefaultProvider(t *testing.T) {
	// Save original env vars
	originalProvider := os.Getenv("DEFAULT_LLM_PROVIDER")
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	defer func() {
		os.Setenv("DEFAULT_LLM_PROVIDER", originalProvider)
		os.Setenv("GEMINI_API_KEY", originalGeminiKey)
	}()

	tests := []struct {
		name          string
		defaultType   string
		apiKey        string
		shouldFail    bool
		expectedError string
	}{
		{
			name:          "Get default Gemini provider",
			defaultType:   "gemini",
			apiKey:        "test-key",
			shouldFail:    false,
			expectedError: "",
		},
		{
			name:          "Fail when default provider API key not set",
			defaultType:   "gemini",
			apiKey:        "",
			shouldFail:    true,
			expectedError: "GEMINI_API_KEY not set in environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DEFAULT_LLM_PROVIDER", tt.defaultType)
			os.Setenv("GEMINI_API_KEY", tt.apiKey)

			factory := NewLLMProviderFactory()
			provider, err := factory.GetDefaultProvider()

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if provider == nil {
					t.Errorf("Expected provider instance, got nil")
				}
			}
		})
	}
}

// TestProviderInterface tests that all providers implement the interface correctly
func TestProviderInterface(t *testing.T) {
	ctx := context.Background()

	// Save original env vars
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	originalClaudeKey := os.Getenv("ANTHROPIC_API_KEY")
	originalGPTKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		os.Setenv("GEMINI_API_KEY", originalGeminiKey)
		os.Setenv("ANTHROPIC_API_KEY", originalClaudeKey)
		os.Setenv("OPENAI_API_KEY", originalGPTKey)
	}()

	// Set test API keys
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("ANTHROPIC_API_KEY", "test-claude-key")
	os.Setenv("OPENAI_API_KEY", "test-gpt-key")

	providers := []struct {
		name     string
		provider LLMProvider
	}{
		{
			name:     "Gemini Provider",
			provider: NewGeminiProvider("test-key"),
		},
		{
			name:     "Claude Provider",
			provider: NewClaudeProvider("test-key"),
		},
		{
			name:     "GPT Provider",
			provider: NewGPTProvider("test-key"),
		},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			// Test GetProviderName
			name := p.provider.GetProviderName()
			if name == "" {
				t.Errorf("GetProviderName() returned empty string")
			}

			// Test that provider implements the interface by checking it can be assigned
			var _ LLMProvider = p.provider

			// Verify provider is not nil
			if p.provider == nil {
				t.Errorf("Provider instance is nil")
			}
		})
	}

	// Test with invalid API keys (should fail gracefully)
	t.Run("Invalid API key handling", func(t *testing.T) {
		gemini := NewGeminiProvider("")
		_, err := gemini.GenerateTrip(ctx, TripGenerationRequest{})
		if err == nil {
			t.Errorf("Expected error with empty API key")
		}
	})
}

// TestInteractiveFeedback tests the feedback structure
func TestInteractiveFeedback(t *testing.T) {
	tests := []struct {
		name     string
		feedback InteractiveFeedback
		isValid  bool
	}{
		{
			name: "Valid feedback with all fields",
			feedback: InteractiveFeedback{
				TripPlanID:     "trip-123",
				CurrentPlan:    `{"trip_name":"Test Trip"}`,
				Feedback:       "Make it more adventurous",
				Interests:      []string{"hiking", "photography"},
				ModifiedBudget: 2000.0,
			},
			isValid: true,
		},
		{
			name: "Valid feedback with minimal fields",
			feedback: InteractiveFeedback{
				CurrentPlan: `{"trip_name":"Test Trip"}`,
				Feedback:    "Add more cultural activities",
			},
			isValid: true,
		},
		{
			name: "Invalid feedback - missing required fields",
			feedback: InteractiveFeedback{
				CurrentPlan: `{"trip_name":"Test Trip"}`,
				// Feedback is required but missing
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check if required field is present
			if tt.isValid && tt.feedback.Feedback == "" {
				t.Errorf("Expected valid feedback but Feedback field is empty")
			}
			if !tt.isValid && tt.feedback.Feedback != "" {
				t.Errorf("Expected invalid feedback but Feedback field is present")
			}
		})
	}
}

// TestCitySuggestion tests the city suggestion structure
func TestCitySuggestion(t *testing.T) {
	suggestion := CitySuggestion{
		City:          "Paris",
		Country:       "France",
		Reason:        "Rich culture and history",
		BestSeason:    "Spring",
		EstimatedDays: 3,
	}

	if suggestion.City == "" {
		t.Errorf("City should not be empty")
	}
	if suggestion.Country == "" {
		t.Errorf("Country should not be empty")
	}
	if suggestion.EstimatedDays <= 0 {
		t.Errorf("EstimatedDays should be positive, got %d", suggestion.EstimatedDays)
	}
}

// TestProviderTypeConstants tests provider type constants
func TestProviderTypeConstants(t *testing.T) {
	if ProviderGemini != "gemini" {
		t.Errorf("ProviderGemini constant incorrect: %s", ProviderGemini)
	}
	if ProviderClaude != "claude" {
		t.Errorf("ProviderClaude constant incorrect: %s", ProviderClaude)
	}
	if ProviderGPT != "gpt" {
		t.Errorf("ProviderGPT constant incorrect: %s", ProviderGPT)
	}
}
