package trips

import (
	"context"
	"fmt"
	"strings"
)

// ClaudeProvider implements LLMProvider for Anthropic's Claude AI
type ClaudeProvider struct {
	service *ClaudeService
}

// NewClaudeProvider creates a new Claude provider instance
func NewClaudeProvider(apiKey string) *ClaudeProvider {
	service := &ClaudeService{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com/v1",
		Model:      "claude-3-5-haiku-20241022",
		HTTPClient: NewClaudeService().HTTPClient,
	}
	return &ClaudeProvider{
		service: service,
	}
}

// GenerateTrip generates a complete trip plan using Claude AI
func (cp *ClaudeProvider) GenerateTrip(ctx context.Context, request TripGenerationRequest) (*AITripResponse, error) {
	// Convert the generic TripGenerationRequest to the format expected by ClaudeService
	// The request types are already compatible, so we can use it directly
	tripPlan, err := cp.service.GenerateTrip(request)
	if err != nil {
		return nil, err
	}
	return tripPlan, nil
}

// RefineTrip refines an existing trip based on user feedback
func (cp *ClaudeProvider) RefineTrip(ctx context.Context, currentTrip *AITripResponse, feedback string) (*AITripResponse, error) {
	if cp.service.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Serialize current trip to JSON
	currentTripJSON, err := serializeTripToJSON(currentTrip)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current trip: %w", err)
	}

	// Build refinement prompt
	prompt := fmt.Sprintf(`You are an expert travel planner. A trip plan has been generated, but the user wants modifications based on their feedback.

**Current Trip Plan:**
%s

**User Feedback:**
%s

**Instructions:**
Based on the user's feedback, modify the trip plan accordingly. Keep the same overall structure but update:
- Activities, times, and locations based on interests
- Budget adjustments if mentioned
- Pace changes (more relaxed or more packed)
- Add or remove activities based on preferences
- Adjust day types and themes
- Keep all required fields and maintain the JSON structure

Return ONLY valid JSON (no markdown, no code blocks) in the exact same structure as the input trip plan.

Generate the refined trip plan now:`, string(currentTripJSON), feedback)

	// Call Claude API with the refinement prompt
	claudeReq := ClaudeRequest{
		Model:     cp.service.Model,
		MaxTokens: 4096,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Execute the request using the existing Claude service infrastructure
	response, err := cp.service.executeRequest(claudeReq)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SuggestCities suggests additional cities for a multi-city trip
func (cp *ClaudeProvider) SuggestCities(ctx context.Context, source, destinations string, preferences []string) ([]CitySuggestion, error) {
	if cp.service.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	prefs := "general travel"
	if len(preferences) > 0 {
		prefs = strings.Join(preferences, ", ")
	}

	prompt := fmt.Sprintf(`You are a travel expert. Suggest 2-4 additional cities to visit on a trip from %s to %s that would make an excellent multi-city itinerary.

**Requirements:**
- Travel interests: %s
- Cities should be geographically logical (on the way or nearby)
- Consider travel time between cities
- Suggest cities that complement the main destination

Return ONLY valid JSON (no markdown, no code blocks) in this exact structure:
{
  "suggestions": [
    {
      "city": "City Name",
      "country": "Country",
      "reason": "Why this city fits well",
      "best_season": "Best time to visit",
      "estimated_days": 2
    }
  ]
}`, source, destinations, prefs)

	claudeReq := ClaudeRequest{
		Model:     cp.service.Model,
		MaxTokens: 1024,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Execute the request
	responseText, err := cp.service.executeRawRequest(claudeReq)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var result struct {
		Suggestions []CitySuggestion `json:"suggestions"`
	}

	if err := parseJSONResponse(responseText, &result); err != nil {
		return nil, fmt.Errorf("failed to parse city suggestions: %w", err)
	}

	return result.Suggestions, nil
}

// GetProviderName returns the provider name
func (cp *ClaudeProvider) GetProviderName() string {
	return "claude"
}
