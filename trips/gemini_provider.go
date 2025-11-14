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

// GeminiProvider implements LLMProvider for Google's Gemini AI
type GeminiProvider struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		APIKey:  apiKey,
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
		Model:   "gemini-1.5-pro", // Using Gemini 1.5 Pro for best results
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents the content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
		Index        int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	PromptFeedback struct {
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"promptFeedback,omitempty"`
}

// AITripResponse is an alias for TripGenerationResponse to satisfy the interface
type AITripResponse = TripGenerationResponse

// GenerateTrip generates a complete trip plan using Gemini AI
func (gp *GeminiProvider) GenerateTrip(ctx context.Context, request TripGenerationRequest) (*AITripResponse, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
	}

	// Build the prompt
	prompt := gp.buildTripGenerationPrompt(request)

	// Create Gemini request
	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
				Role: "user",
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 4096,
			TopP:            0.95,
			TopK:            40,
		},
	}

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the URL with API key
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", gp.BaseURL, gp.Model, gp.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Extract the text response
	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Clean up the response (remove markdown code blocks if present)
	responseText = cleanJSONResponse(responseText)

	// Parse the JSON response
	var tripPlan AITripResponse
	if err := json.Unmarshal([]byte(responseText), &tripPlan); err != nil {
		return nil, fmt.Errorf("failed to parse trip plan: %w. Response: %s", err, responseText)
	}

	return &tripPlan, nil
}

// RefineTrip refines an existing trip based on user feedback
func (gp *GeminiProvider) RefineTrip(ctx context.Context, currentTrip *AITripResponse, feedback string) (*AITripResponse, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
	}

	// Serialize current trip to JSON
	currentTripJSON, err := json.Marshal(currentTrip)
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

	// Create Gemini request
	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
				Role: "user",
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 4096,
			TopP:            0.95,
			TopK:            40,
		},
	}

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", gp.BaseURL, gp.Model, gp.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	responseText = cleanJSONResponse(responseText)

	var refinedPlan AITripResponse
	if err := json.Unmarshal([]byte(responseText), &refinedPlan); err != nil {
		return nil, fmt.Errorf("failed to parse refined trip plan: %w", err)
	}

	return &refinedPlan, nil
}

// SuggestCities suggests additional cities for a multi-city trip
func (gp *GeminiProvider) SuggestCities(ctx context.Context, source, destinations string, preferences []string) ([]CitySuggestion, error) {
	if gp.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
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

	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
				Role: "user",
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 1024,
			TopP:            0.95,
			TopK:            40,
		},
	}

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", gp.BaseURL, gp.Model, gp.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gp.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	responseText = cleanJSONResponse(responseText)

	var result struct {
		Suggestions []CitySuggestion `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse city suggestions: %w", err)
	}

	return result.Suggestions, nil
}

// GetProviderName returns the provider name
func (gp *GeminiProvider) GetProviderName() string {
	return "gemini"
}

// buildTripGenerationPrompt creates the prompt for Gemini
func (gp *GeminiProvider) buildTripGenerationPrompt(req TripGenerationRequest) string {
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

	prompt := fmt.Sprintf(`You are an expert travel planner specializing in creating comprehensive, personalized trip itineraries. Create a detailed trip plan based on the following requirements:

**Trip Details:**
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

**Instructions:**
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

**Output Format:**
Return ONLY valid JSON (no markdown, no code blocks, no explanations) in this exact structure:
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
}

Generate the trip plan now:`, req.Source, destinations, req.StartDate.Format("2006-01-02"),
		req.EndDate.Format("2006-01-02"), duration, req.NumTravelers, currency, budgetInfo, preferences, pace, preferences)

	return prompt
}

// cleanJSONResponse removes markdown code blocks and extra whitespace from JSON responses
func cleanJSONResponse(text string) string {
	// Remove markdown code blocks
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	return text
}
