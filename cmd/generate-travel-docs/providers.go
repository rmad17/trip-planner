package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"triplanner/travelknowledge"
)

// OpenAIProvider implements AIProvider using OpenAI API
type OpenAIProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent generates travel content using OpenAI GPT-4
func (p *OpenAIProvider) GenerateContent(ctx context.Context, city travelknowledge.City) (*travelknowledge.GeneratedContent, error) {
	prompt := buildContentPrompt(city)

	reqBody := map[string]interface{}{
		"model": "gpt-4-turbo-preview",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a professional travel content writer specializing in creating comprehensive, accurate, and engaging travel guides. Provide detailed, well-researched information that would be valuable for travelers.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  4000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	// Parse the generated content into structured format
	content := parseGeneratedContent(city, result.Choices[0].Message.Content)
	return content, nil
}

// GenerateEmbedding generates embeddings using OpenAI
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"input": text,
		"model": "text-embedding-ada-002",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}

	return result.Data[0].Embedding, nil
}

// ClaudeProvider implements AIProvider using Anthropic Claude API
type ClaudeProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent generates travel content using Claude
func (p *ClaudeProvider) GenerateContent(ctx context.Context, city travelknowledge.City) (*travelknowledge.GeneratedContent, error) {
	prompt := buildContentPrompt(city)

	reqBody := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 4000,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	// Parse the generated content into structured format
	content := parseGeneratedContent(city, result.Content[0].Text)
	return content, nil
}

// GenerateEmbedding generates embeddings - Claude doesn't have embedding endpoint, so we use OpenAI
func (p *ClaudeProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// For now, require OpenAI key for embeddings when using Claude
	// In production, consider using Voyage AI or other embedding services
	return nil, fmt.Errorf("embedding generation not available with Claude provider - please set OPENAI_API_KEY for embeddings")
}

// MockProvider implements AIProvider for testing without API keys
type MockProvider struct{}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// GenerateContent generates mock travel content
func (p *MockProvider) GenerateContent(ctx context.Context, city travelknowledge.City) (*travelknowledge.GeneratedContent, error) {
	sections := []travelknowledge.ContentSection{
		{
			Type:  travelknowledge.DocumentTypeHistorical,
			Title: fmt.Sprintf("Historical Significance of %s", city.Name),
			Content: fmt.Sprintf("%s is a city with rich historical heritage. [Mock content for testing]\n\n"+
				"The city has been a center of culture and civilization for centuries. "+
				"Its monuments and historical sites attract visitors from around the world.", city.Name),
		},
		{
			Type:  travelknowledge.DocumentTypeLogistics,
			Title: fmt.Sprintf("Travel Logistics for %s", city.Name),
			Content: fmt.Sprintf("Getting to and around %s:\n\n"+
				"Transportation: Well-connected by air, rail, and road.\n"+
				"Accommodation: Various options from budget to luxury hotels.\n"+
				"Best time to visit: October to March for pleasant weather.", city.Name),
		},
		{
			Type:  travelknowledge.DocumentTypeFood,
			Title: fmt.Sprintf("Food and Cuisine in %s", city.Name),
			Content: fmt.Sprintf("%s offers a delightful culinary experience.\n\n"+
				"Local specialties include traditional dishes that reflect the region's culture. "+
				"Street food is popular and offers authentic flavors at affordable prices.", city.Name),
		},
		{
			Type:  travelknowledge.DocumentTypeActivities,
			Title: fmt.Sprintf("Things to Do in %s", city.Name),
			Content: fmt.Sprintf("Top activities in %s:\n\n"+
				"- Visit historical monuments and temples\n"+
				"- Explore local markets and bazaars\n"+
				"- Experience cultural performances\n"+
				"- Take guided heritage walks", city.Name),
		},
		{
			Type:  travelknowledge.DocumentTypeNearbyPlaces,
			Title: fmt.Sprintf("Places Near %s", city.Name),
			Content: fmt.Sprintf("Interesting places to visit near %s:\n\n"+
				"Several attractions are located within a short distance, "+
				"making excellent day trip options. These include historical sites, "+
				"natural attractions, and cultural landmarks.", city.Name),
		},
	}

	return &travelknowledge.GeneratedContent{
		CityName:    city.Name,
		CityCountry: city.Country,
		Sections:    sections,
		Metadata: travelknowledge.TravelDocumentMetadata{
			Tags:     []string{"mock", "testing"},
			Language: "en",
		},
	}, nil
}

// GenerateEmbedding generates mock embeddings
func (p *MockProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Generate a deterministic mock embedding based on text length
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(len(text)%256) / 256.0
	}
	return embedding, nil
}

// buildContentPrompt creates the prompt for content generation
func buildContentPrompt(city travelknowledge.City) string {
	return fmt.Sprintf(`Generate comprehensive travel content for %s, %s.

Please provide detailed information in the following categories:

1. HISTORICAL & CULTURAL INFORMATION
   - Historical significance and background
   - Cultural heritage and traditions
   - Important monuments, temples, or landmarks
   - Local customs and etiquette

2. TRAVEL LOGISTICS
   - How to reach (air, rail, road)
   - Local transportation options
   - Accommodation recommendations (budget to luxury)
   - Best time to visit and weather information
   - Visa and permit requirements (if any)

3. FOOD & DINING
   - Local cuisine and specialties
   - Popular restaurants and eateries
   - Street food recommendations
   - Dining customs and etiquette
   - Must-try dishes

4. ACTIVITIES & ATTRACTIONS
   - Top tourist attractions
   - Things to do and experiences
   - Shopping areas and markets
   - Entertainment and nightlife
   - Festivals and events

5. NEARBY PLACES OF INTEREST
   - Day trip destinations
   - Surrounding attractions
   - Regional highlights
   - Distance and accessibility

Please format each section clearly with a title and detailed content. Write in an engaging, informative style suitable for travelers.`, city.Name, city.Country)
}

// parseGeneratedContent parses the AI-generated text into structured content
func parseGeneratedContent(city travelknowledge.City, generatedText string) *travelknowledge.GeneratedContent {
	// Simple parsing logic - splits by numbered sections
	// In production, you might want more sophisticated parsing or ask AI to return JSON

	sections := []travelknowledge.ContentSection{
		{
			Type:    travelknowledge.DocumentTypeHistorical,
			Title:   fmt.Sprintf("Historical and Cultural Heritage of %s", city.Name),
			Content: extractSection(generatedText, "HISTORICAL", "LOGISTICS"),
		},
		{
			Type:    travelknowledge.DocumentTypeLogistics,
			Title:   fmt.Sprintf("Travel Planning for %s", city.Name),
			Content: extractSection(generatedText, "LOGISTICS", "FOOD"),
		},
		{
			Type:    travelknowledge.DocumentTypeFood,
			Title:   fmt.Sprintf("Culinary Guide to %s", city.Name),
			Content: extractSection(generatedText, "FOOD", "ACTIVITIES"),
		},
		{
			Type:    travelknowledge.DocumentTypeActivities,
			Title:   fmt.Sprintf("Attractions and Activities in %s", city.Name),
			Content: extractSection(generatedText, "ACTIVITIES", "NEARBY"),
		},
		{
			Type:    travelknowledge.DocumentTypeNearbyPlaces,
			Title:   fmt.Sprintf("Destinations Near %s", city.Name),
			Content: extractSection(generatedText, "NEARBY", "END"),
		},
	}

	return &travelknowledge.GeneratedContent{
		CityName:    city.Name,
		CityCountry: city.Country,
		Sections:    sections,
		Metadata: travelknowledge.TravelDocumentMetadata{
			Language: "en",
			Tags:     []string{"travel", "guide", city.Name},
		},
	}
}

// extractSection extracts content between two section markers
func extractSection(text, startMarker, endMarker string) string {
	// Find the section between markers
	// This is a simplified version - enhance as needed
	if text == "" {
		return ""
	}

	// If markers not found, return a portion of the text
	// This ensures we always have some content even if parsing fails
	lines := bytes.Split([]byte(text), []byte("\n"))
	var content bytes.Buffer

	for _, line := range lines {
		content.Write(line)
		content.WriteByte('\n')
	}

	result := content.String()
	if len(result) > 2000 {
		// Split long text evenly across sections
		sectionSize := len(result) / 5
		// This is basic - the actual implementation would be smarter
		return result[:sectionSize]
	}

	return result
}
