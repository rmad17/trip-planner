package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	amadeusTestBaseURL = "https://test.api.amadeus.com"
	amadeusProdBaseURL = "https://api.amadeus.com"
)

// AmadeusClient is a shared OAuth2 token cache + HTTP client used by both
// the hotels and flights providers (they share the same client_id/secret).
type AmadeusClient struct {
	APIKey    string
	APISecret string
	BaseURL   string

	HTTPClient *http.Client

	mu         sync.Mutex
	token      string
	tokenExpAt time.Time
}

// AmadeusTokenResponse is the OAuth2 token response shape.
type AmadeusTokenResponse struct {
	Type            string `json:"type"`
	Username        string `json:"username"`
	ApplicationName string `json:"application_name"`
	ClientID        string `json:"client_id"`
	TokenType       string `json:"token_type"`
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in"`
	State           string `json:"state"`
	Scope           string `json:"scope"`
}

// AmadeusErrorResponse is the standard error envelope returned by Amadeus.
type AmadeusErrorResponse struct {
	Errors []struct {
		Status int    `json:"status"`
		Code   int    `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

var (
	defaultAmadeusClient *AmadeusClient
	amadeusClientOnce    sync.Once
)

// DefaultAmadeusClient returns a process-wide singleton client, configured from env.
func DefaultAmadeusClient() *AmadeusClient {
	amadeusClientOnce.Do(func() {
		defaultAmadeusClient = NewAmadeusClient(
			os.Getenv("AMADEUS_API_KEY"),
			os.Getenv("AMADEUS_API_SECRET"),
			os.Getenv("AMADEUS_ENV"),
		)
	})
	return defaultAmadeusClient
}

// NewAmadeusClient constructs a client. env is "test" (default) or "production".
func NewAmadeusClient(apiKey, apiSecret, env string) *AmadeusClient {
	base := amadeusTestBaseURL
	if strings.ToLower(env) == "production" {
		base = amadeusProdBaseURL
	}
	return &AmadeusClient{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   base,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Configured reports whether credentials were provided.
func (c *AmadeusClient) Configured() bool {
	return c != nil && c.APIKey != "" && c.APISecret != ""
}

// Token returns a valid bearer token, refreshing if expired.
func (c *AmadeusClient) Token(ctx context.Context) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("amadeus: API key/secret not configured")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpAt.Add(-30*time.Second)) {
		return c.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.APIKey)
	form.Set("client_secret", c.APISecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/v1/security/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("amadeus: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("amadeus: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("amadeus: read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("amadeus: token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokResp AmadeusTokenResponse
	if err := json.Unmarshal(body, &tokResp); err != nil {
		return "", fmt.Errorf("amadeus: parse token response: %w", err)
	}

	c.token = tokResp.AccessToken
	c.tokenExpAt = time.Now().Add(time.Duration(tokResp.ExpiresIn) * time.Second)
	return c.token, nil
}

// DoJSON performs an authenticated GET request and decodes the JSON body into out.
// Path is the API path (e.g. "/v3/shopping/flight-offers"), q are query params.
func (c *AmadeusClient) DoJSON(ctx context.Context, path string, q url.Values, out interface{}) error {
	tok, err := c.Token(ctx)
	if err != nil {
		return err
	}

	endpoint := c.BaseURL + path
	if len(q) > 0 {
		endpoint += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("amadeus: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/vnd.amadeus+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("amadeus: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("amadeus: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr AmadeusErrorResponse
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && len(apiErr.Errors) > 0 {
			return fmt.Errorf("amadeus: %s (status %d): %s", apiErr.Errors[0].Title, resp.StatusCode, apiErr.Errors[0].Detail)
		}
		return fmt.Errorf("amadeus: request failed (status %d): %s", resp.StatusCode, string(body))
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("amadeus: parse response: %w", err)
		}
	}
	return nil
}
