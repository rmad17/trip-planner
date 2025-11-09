package trips

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ClaudeService handles interactions with Anthropic's Claude API
type ClaudeService struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewClaudeService creates a new Claude service instance
func NewClaudeService() *ClaudeService {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: ANTHROPIC_API_KEY environment variable not set")
	}

	return &ClaudeService{
		APIKey:  apiKey,
		BaseURL: "https://api.anthropic.com/v1",
		Model:   "claude-3-5-haiku-20241022", // Claude 3.5 Haiku
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeMessage represents a message in the Claude conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// TripGenerationRequest represents the input for trip generation
type TripGenerationRequest struct {
	Source          string    `json:"source" binding:"required"`
	SourcePlaceID   string    `json:"source_place_id,omitempty"`
	Destinations    []string  `json:"destinations" binding:"required,min=1"`
	DestinationIDs  []string  `json:"destination_ids,omitempty"`
	StartDate       time.Time `json:"start_date" binding:"required"`
	EndDate         time.Time `json:"end_date" binding:"required"`
	NumTravelers    int       `json:"num_travelers" binding:"required,min=1"`
	Budget          float64   `json:"budget,omitempty"`
	Currency        Currency  `json:"currency"`
	TripPreferences []string  `json:"trip_preferences,omitempty"` // e.g., ["adventure", "culture", "food"]
	PacePreference  string    `json:"pace_preference,omitempty"`  // "relaxed", "moderate", "fast"
}

// TripGenerationResponse represents Claude's structured trip plan
type TripGenerationResponse struct {
	TripName        string                    `json:"trip_name"`
	Description     string                    `json:"description"`
	RecommendedMode string                    `json:"recommended_mode"` // "flight", "train", "car", "bus", "mixed"
	TotalDays       int                       `json:"total_days"`
	EstimatedBudget float64                   `json:"estimated_budget"`
	BudgetBreakdown TripBudgetBreakdown       `json:"budget_breakdown"`
	Hops            []GeneratedHop            `json:"hops"`
	DailyItinerary  []GeneratedDay            `json:"daily_itinerary"`
	TravelTips      []string                  `json:"travel_tips"`
	BestTimeToVisit string                    `json:"best_time_to_visit"`
	Considerations  []string                  `json:"considerations"`
}

// TripBudgetBreakdown provides budget allocation
type TripBudgetBreakdown struct {
	Accommodation  float64 `json:"accommodation"`
	Transportation float64 `json:"transportation"`
	Food           float64 `json:"food"`
	Activities     float64 `json:"activities"`
	Shopping       float64 `json:"shopping"`
	Miscellaneous  float64 `json:"miscellaneous"`
}

// GeneratedHop represents a suggested destination/hop
type GeneratedHop struct {
	Name            string   `json:"name"`
	City            string   `json:"city"`
	Country         string   `json:"country"`
	Description     string   `json:"description"`
	StartDate       string   `json:"start_date"` // ISO format
	EndDate         string   `json:"end_date"`
	Duration        int      `json:"duration"` // days
	Transportation  string   `json:"transportation"`
	EstimatedBudget float64  `json:"estimated_budget"`
	POIs            []string `json:"pois"`
	Restaurants     []string `json:"restaurants"`
	Activities      []string `json:"activities"`
	HopOrder        int      `json:"hop_order"`
}

// GeneratedDay represents a daily itinerary
type GeneratedDay struct {
	DayNumber       int                `json:"day_number"`
	Date            string             `json:"date"` // ISO format
	Title           string             `json:"title"`
	DayType         string             `json:"day_type"` // "travel", "explore", etc.
	Location        string             `json:"location"`
	Description     string             `json:"description"`
	EstimatedBudget float64            `json:"estimated_budget"`
	Activities      []GeneratedActivity `json:"activities"`
	TravelTime      string             `json:"travel_time,omitempty"`
	Notes           string             `json:"notes,omitempty"`
}

// GeneratedActivity represents a suggested activity
type GeneratedActivity struct {
	Name          string  `json:"name"`
	ActivityType  string  `json:"activity_type"` // "sightseeing", "dining", etc.
	StartTime     string  `json:"start_time"`    // "09:00"
	EndTime       string  `json:"end_time"`      // "11:00"
	Duration      int     `json:"duration"`      // minutes
	Location      string  `json:"location"`
	Description   string  `json:"description"`
	EstimatedCost float64 `json:"estimated_cost"`
	Priority      int     `json:"priority"` // 1-5
	Tips          string  `json:"tips,omitempty"`
}

// GenerateTrip uses Claude AI to generate a comprehensive trip plan
func (cs *ClaudeService) GenerateTrip(req TripGenerationRequest) (*TripGenerationResponse, error) {
	if cs.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Build the prompt for Claude
	prompt := cs.buildTripGenerationPrompt(req)

	// Call Claude API
	claudeReq := ClaudeRequest{
		Model:     cs.Model,
		MaxTokens: 4096,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cs.BaseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", cs.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := cs.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse the JSON response from Claude
	var tripPlan TripGenerationResponse
	if err := json.Unmarshal([]byte(claudeResp.Content[0].Text), &tripPlan); err != nil {
		return nil, fmt.Errorf("failed to parse trip plan: %w", err)
	}

	return &tripPlan, nil
}

// buildTripGenerationPrompt creates the prompt for Claude
func (cs *ClaudeService) buildTripGenerationPrompt(req TripGenerationRequest) string {
	destinations := ""
	for i, dest := range req.Destinations {
		if i > 0 {
			destinations += ", "
		}
		destinations += dest
	}

	currency := string(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	preferences := "general sightseeing"
	if len(req.TripPreferences) > 0 {
		preferences = ""
		for i, pref := range req.TripPreferences {
			if i > 0 {
				preferences += ", "
			}
			preferences += pref
		}
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

	prompt := fmt.Sprintf(`You are an expert travel planner. Create a comprehensive, detailed trip itinerary based on the following requirements:

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

**Output Format:**
Return ONLY valid JSON (no markdown, no code blocks) in this exact structure:
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
		req.EndDate.Format("2006-01-02"), duration, req.NumTravelers, currency, budgetInfo, preferences, pace)

	return prompt
}

// SuggestMultiCityRoute suggests additional cities to include in a multi-city trip
func (cs *ClaudeService) SuggestMultiCityRoute(source, primaryDestination string, duration int, preferences []string) ([]string, error) {
	if cs.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	prefs := "general travel"
	if len(preferences) > 0 {
		prefs = ""
		for i, pref := range preferences {
			if i > 0 {
				prefs += ", "
			}
			prefs += pref
		}
	}

	prompt := fmt.Sprintf(`You are a travel expert. Suggest 2-4 additional cities to visit on a trip from %s to %s that would make an excellent multi-city itinerary.

**Requirements:**
- Trip duration: %d days
- Travel interests: %s
- Cities should be geographically logical (on the way or nearby)
- Consider travel time between cities
- Suggest cities that complement the main destination

Return ONLY a JSON array of city suggestions (no markdown, no explanations):
{
  "suggestions": [
    {
      "city": "City Name",
      "country": "Country",
      "reason": "Why this city fits well",
      "recommended_days": number,
      "travel_time_from_previous": "X hours by train/flight"
    }
  ]
}`, source, primaryDestination, duration, prefs)

	claudeReq := ClaudeRequest{
		Model:     cs.Model,
		MaxTokens: 1024,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cs.BaseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", cs.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := cs.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse the JSON response
	var suggestions struct {
		Suggestions []struct {
			City                   string `json:"city"`
			Country                string `json:"country"`
			Reason                 string `json:"reason"`
			RecommendedDays        int    `json:"recommended_days"`
			TravelTimeFromPrevious string `json:"travel_time_from_previous"`
		} `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(claudeResp.Content[0].Text), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	// Format the suggestions as strings
	result := make([]string, len(suggestions.Suggestions))
	for i, sugg := range suggestions.Suggestions {
		result[i] = fmt.Sprintf("%s, %s (%d days) - %s. Travel: %s",
			sugg.City, sugg.Country, sugg.RecommendedDays, sugg.Reason, sugg.TravelTimeFromPrevious)
	}

	return result, nil
}
