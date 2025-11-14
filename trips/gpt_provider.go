package trips

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GPTProvider implements LLMProvider for OpenAI's GPT models
type GPTProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewGPTProvider creates a new GPT provider instance
func NewGPTProvider(apiKey string) *GPTProvider {
	return &GPTProvider{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-4-turbo-preview", // Using GPT-4 Turbo for best results
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GPTRequest represents the request structure for OpenAI API
type GPTRequest struct {
	Model            string        `json:"model"`
	Messages         []GPTMessage  `json:"messages"`
	Temperature      float64       `json:"temperature,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	ResponseFormat   *GPTResponseFormat `json:"response_format,omitempty"`
}

// GPTMessage represents a message in the GPT conversation
type GPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GPTResponseFormat specifies the response format
type GPTResponseFormat struct {
	Type string `json:"type"` // "json_object" or "text"
}

// GPTResponse represents the response from OpenAI API
type GPTResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateTrip generates a complete trip plan using OpenAI GPT
func (gp *GPTProvider) GenerateTrip(ctx context.Context, request TripGenerationRequest) (*AITripResponse, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	// Build the prompt
	prompt := gp.buildTripGenerationPrompt(request)

	// Create GPT request
	gptReq := GPTRequest{
		Model: gp.Model,
		Messages: []GPTMessage{
			{
				Role:    "system",
				Content: "You are an expert travel planner who creates comprehensive, personalized trip itineraries. Always respond with valid JSON only, no markdown formatting.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   4096,
		ResponseFormat: &GPTResponseFormat{
			Type: "json_object",
		},
	}

	reqBody, err := json.Marshal(gptReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", gp.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+gp.APIKey)

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gptResp GPTResponse
	if err := json.Unmarshal(body, &gptResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gptResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	// Extract the text response
	responseText := gptResp.Choices[0].Message.Content

	// Parse the JSON response
	var tripPlan AITripResponse
	if err := json.Unmarshal([]byte(responseText), &tripPlan); err != nil {
		return nil, fmt.Errorf("failed to parse trip plan: %w. Response: %s", err, responseText)
	}

	return &tripPlan, nil
}

// RefineTrip refines an existing trip based on user feedback
func (gp *GPTProvider) RefineTrip(ctx context.Context, currentTrip *AITripResponse, feedback string) (*AITripResponse, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	// Serialize current trip to JSON
	currentTripJSON, err := json.Marshal(currentTrip)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current trip: %w", err)
	}

	// Build refinement prompt
	prompt := fmt.Sprintf(`A trip plan has been generated, but the user wants modifications based on their feedback.

Current Trip Plan:
%s

User Feedback:
%s

Instructions:
Based on the user's feedback, modify the trip plan accordingly. Keep the same overall structure but update:
- Activities, times, and locations based on interests
- Budget adjustments if mentioned
- Pace changes (more relaxed or more packed)
- Add or remove activities based on preferences
- Adjust day types and themes
- Keep all required fields and maintain the JSON structure

Return the modified trip plan as valid JSON with the same structure.`, string(currentTripJSON), feedback)

	// Create GPT request
	gptReq := GPTRequest{
		Model: gp.Model,
		Messages: []GPTMessage{
			{
				Role:    "system",
				Content: "You are an expert travel planner who refines trip itineraries based on user feedback. Always respond with valid JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   4096,
		ResponseFormat: &GPTResponseFormat{
			Type: "json_object",
		},
	}

	reqBody, err := json.Marshal(gptReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", gp.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+gp.APIKey)

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gptResp GPTResponse
	if err := json.Unmarshal(body, &gptResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gptResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	responseText := gptResp.Choices[0].Message.Content

	var refinedPlan AITripResponse
	if err := json.Unmarshal([]byte(responseText), &refinedPlan); err != nil {
		return nil, fmt.Errorf("failed to parse refined trip plan: %w", err)
	}

	return &refinedPlan, nil
}

// SuggestCities suggests additional cities for a multi-city trip
func (gp *GPTProvider) SuggestCities(ctx context.Context, source, destinations string, preferences []string) ([]CitySuggestion, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	prefs := "general travel"
	if len(preferences) > 0 {
		prefs = strings.Join(preferences, ", ")
	}

	prompt := fmt.Sprintf(`Suggest 2-4 additional cities to visit on a trip from %s to %s that would make an excellent multi-city itinerary.

Requirements:
- Travel interests: %s
- Cities should be geographically logical (on the way or nearby)
- Consider travel time between cities
- Suggest cities that complement the main destination

Return valid JSON in this exact structure:
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

	gptReq := GPTRequest{
		Model: gp.Model,
		Messages: []GPTMessage{
			{
				Role:    "system",
				Content: "You are a travel expert who suggests destinations. Always respond with valid JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
		ResponseFormat: &GPTResponseFormat{
			Type: "json_object",
		},
	}

	reqBody, err := json.Marshal(gptReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", gp.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+gp.APIKey)

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gptResp GPTResponse
	if err := json.Unmarshal(body, &gptResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gptResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	responseText := gptResp.Choices[0].Message.Content

	var result struct {
		Suggestions []CitySuggestion `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse city suggestions: %w", err)
	}

	return result.Suggestions, nil
}

// GetProviderName returns the provider name
func (gp *GPTProvider) GetProviderName() string {
	return "gpt"
}

// buildTripGenerationPrompt creates the prompt for GPT
func (gp *GPTProvider) buildTripGenerationPrompt(req TripGenerationRequest) string {
	destinations := strings.Join(req.Destinations, ", ")

	currency := string(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	preferences := "general sightseeing"
	if len(req.TripPreferences) > 0 {
		preferences = strings.Join(req.TripPreferences, ", ")
	}

	pace := req.PacePreference
	if pace == "" {
		pace = "moderate"
	}

	duration := int(req.EndDate.Sub(req.StartDate).Hours() / 24)

	budgetInfo := ""
	if req.Budget > 0 {
		budgetInfo = fmt.Sprintf(" with a budget of %.2f %s per person", req.Budget, currency)
	}

	prompt := fmt.Sprintf(`Create a comprehensive, detailed trip itinerary based on the following requirements:

Trip Details:
- Source: %s
- Destinations: %s
- Start Date: %s
- End Date: %s
- Duration: %d days
- Number of Travelers: %d
- Currency: %s
%s
- Trip Preferences: %s
- Pace: %s

Instructions:
1. Analyze the route and recommend the BEST travel mode (flight, train, car, bus, or mixed)
2. Create a multi-city itinerary with optimal sequence if multiple destinations
3. Suggest additional cities/stops that make sense geographically (max 2 extra suggestions)
4. For each destination (hop), provide:
   - Name, city, country, description
   - Recommended duration (days)
   - Transportation method to reach it
   - Top POIs (points of interest)
   - Restaurant recommendations
   - Local activities
   - Budget estimate
5. Create a DAY-BY-DAY itinerary with:
   - Daily theme/title
   - Type of day (travel/explore/relax/adventure/cultural)
   - 4-6 activities per day with times, locations, costs
   - Realistic time estimates including travel time
   - Meal suggestions
6. Provide budget breakdown by category
7. Include travel tips and best time considerations
8. Consider the user's interests (%s) throughout the planning

Return valid JSON in this exact structure:
{
  "trip_name": "string",
  "description": "string",
  "recommended_mode": "flight|train|car|bus|mixed",
  "total_days": number,
  "estimated_budget": number,
  "budget_breakdown": {
    "accommodation": number,
    "transportation": number,
    "food": number,
    "activities": number,
    "shopping": number,
    "miscellaneous": number
  },
  "hops": [
    {
      "name": "string",
      "city": "string",
      "country": "string",
      "description": "string",
      "start_date": "YYYY-MM-DD",
      "end_date": "YYYY-MM-DD",
      "duration": number,
      "transportation": "string",
      "estimated_budget": number,
      "pois": ["string"],
      "restaurants": ["string"],
      "activities": ["string"],
      "hop_order": number
    }
  ],
  "daily_itinerary": [
    {
      "day_number": number,
      "date": "YYYY-MM-DD",
      "title": "string",
      "day_type": "travel|explore|relax|adventure|cultural|business",
      "location": "string",
      "description": "string",
      "estimated_budget": number,
      "activities": [
        {
          "name": "string",
          "activity_type": "transport|sightseeing|dining|shopping|entertainment|adventure|cultural|other",
          "start_time": "HH:MM",
          "end_time": "HH:MM",
          "duration": number,
          "location": "string",
          "description": "string",
          "estimated_cost": number,
          "priority": number,
          "tips": "string"
        }
      ],
      "travel_time": "string",
      "notes": "string"
    }
  ],
  "travel_tips": ["string"],
  "best_time_to_visit": "string",
  "considerations": ["string"]
}`, req.Source, destinations, req.StartDate.Format("2006-01-02"),
		req.EndDate.Format("2006-01-02"), duration, req.NumTravelers, currency, budgetInfo, preferences, pace, preferences)

	return prompt
}
