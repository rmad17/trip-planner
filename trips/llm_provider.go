package trips

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// LLMProvider defines the interface that all LLM providers must implement
type LLMProvider interface {
	// GenerateTrip generates a complete trip plan based on the request
	GenerateTrip(ctx context.Context, request TripGenerationRequest) (*AITripResponse, error)

	// RefineTrip refines an existing trip based on user feedback
	RefineTrip(ctx context.Context, currentTrip *AITripResponse, feedback string) (*AITripResponse, error)

	// SuggestCities suggests additional cities for a multi-city trip
	SuggestCities(ctx context.Context, source, destinations string, preferences []string) ([]CitySuggestion, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string
}

// LLMProviderType represents the type of LLM provider
type LLMProviderType string

const (
	ProviderClaude LLMProviderType = "claude"
	ProviderGPT    LLMProviderType = "gpt"
	ProviderGemini LLMProviderType = "gemini"
)

// InteractiveFeedback represents user feedback for trip refinement
type InteractiveFeedback struct {
	TripPlanID     string   `json:"trip_plan_id,omitempty"`
	CurrentPlan    string   `json:"current_plan,omitempty"`
	Feedback       string   `json:"feedback" binding:"required"`
	Interests      []string `json:"interests,omitempty"`
	ModifiedBudget float64  `json:"modified_budget,omitempty"`
	ModifiedDates  struct {
		StartDate string `json:"start_date,omitempty"`
		EndDate   string `json:"end_date,omitempty"`
	} `json:"modified_dates,omitempty"`
}

// CitySuggestion represents a suggested city for multi-city trips
type CitySuggestion struct {
	City        string `json:"city"`
	Country     string `json:"country"`
	Reason      string `json:"reason"`
	BestSeason  string `json:"best_season,omitempty"`
	EstimatedDays int  `json:"estimated_days,omitempty"`
}

// LLMProviderFactory creates LLM provider instances
type LLMProviderFactory struct {
	defaultProvider LLMProviderType
}

// NewLLMProviderFactory creates a new factory with the default provider from environment
func NewLLMProviderFactory() *LLMProviderFactory {
	defaultProvider := os.Getenv("DEFAULT_LLM_PROVIDER")
	if defaultProvider == "" {
		defaultProvider = string(ProviderGemini) // Default to Gemini
	}

	return &LLMProviderFactory{
		defaultProvider: LLMProviderType(defaultProvider),
	}
}

// GetProvider returns an LLM provider instance based on the provider type
func (f *LLMProviderFactory) GetProvider(providerType LLMProviderType) (LLMProvider, error) {
	if providerType == "" {
		providerType = f.defaultProvider
	}

	switch providerType {
	case ProviderClaude:
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, errors.New("ANTHROPIC_API_KEY not set in environment")
		}
		return NewClaudeProvider(apiKey), nil

	case ProviderGPT:
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, errors.New("OPENAI_API_KEY not set in environment")
		}
		return NewGPTProvider(apiKey), nil

	case ProviderGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, errors.New("GEMINI_API_KEY not set in environment")
		}
		return NewGeminiProvider(apiKey), nil

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", providerType)
	}
}

// GetDefaultProvider returns the default provider instance
func (f *LLMProviderFactory) GetDefaultProvider() (LLMProvider, error) {
	return f.GetProvider(f.defaultProvider)
}

// GetDefaultProviderType returns the configured default provider type
func (f *LLMProviderFactory) GetDefaultProviderType() LLMProviderType {
	return f.defaultProvider
}
