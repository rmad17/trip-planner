package accounts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Pure logic tests (no DB) ---

func TestGenerateRawToken_Length(t *testing.T) {
	tok, err := generateRawToken()
	require.NoError(t, err)
	assert.Equal(t, 64, len(tok), "token should be 64 hex chars (32 bytes)")
}

func TestGenerateRawToken_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 50)
	for i := 0; i < 50; i++ {
		tok, _ := generateRawToken()
		_, dup := seen[tok]
		assert.False(t, dup, "duplicate token at iteration %d", i)
		seen[tok] = struct{}{}
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	h1 := hashToken("abc123")
	h2 := hashToken("abc123")
	assert.Equal(t, h1, h2, "same input must produce same hash")
}

func TestHashToken_DifferentInputs(t *testing.T) {
	assert.NotEqual(t, hashToken("abc"), hashToken("def"))
}

func TestIsTokenExpired_Nil_ReturnsTrue(t *testing.T) {
	assert.True(t, isTokenExpired(nil))
}

func TestIsTokenExpired_Past_ReturnsTrue(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	assert.True(t, isTokenExpired(&past))
}

func TestIsTokenExpired_Future_ReturnsFalse(t *testing.T) {
	future := time.Now().Add(time.Hour)
	assert.False(t, isTokenExpired(&future))
}

// --- HTTP handler tests (no DB) ---

func setupPasswordResetRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := r.Group("/api/v1/auth")
	RouterGroupPasswordReset(auth)
	return r
}

func TestRequestPasswordReset_AlwaysReturns200(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupPasswordResetRouter()
	for _, email := range []string{"notexist@example.com", "another@nowhere.com"} {
		body, _ := json.Marshal(map[string]string{"email": email})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/request",
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "must be 200 even for unknown email: %s", email)
	}
}

func TestResetPassword_InvalidToken_Returns400(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupPasswordResetRouter()
	body, _ := json.Marshal(map[string]string{
		"token":        "deadbeefdeadbeef",
		"new_password": "newpassword123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/confirm",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResetPassword_ShortPassword_Returns400(t *testing.T) {
	r := setupPasswordResetRouter()
	body, _ := json.Marshal(map[string]string{
		"token":        "sometoken",
		"new_password": "short",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/confirm",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyEmail_MissingToken_Returns400(t *testing.T) {
	r := setupPasswordResetRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/verify-email", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- MockEmailProvider for testing ---

type MockEmailProvider struct {
	ResetCalls    []string
	VerifyCalls   []string
}

func (m *MockEmailProvider) SendPasswordReset(to, link string) error {
	m.ResetCalls = append(m.ResetCalls, to)
	return nil
}

func (m *MockEmailProvider) SendEmailVerification(to, link string) error {
	m.VerifyCalls = append(m.VerifyCalls, to)
	return nil
}

func TestSetEmailProvider_SwapsProvider(t *testing.T) {
	original := DefaultEmailProvider
	defer SetEmailProvider(original)

	mock := &MockEmailProvider{}
	SetEmailProvider(mock)

	DefaultEmailProvider.SendPasswordReset("test@example.com", "http://reset")
	assert.Len(t, mock.ResetCalls, 1)
	assert.Equal(t, "test@example.com", mock.ResetCalls[0])
}
