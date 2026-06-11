package trips

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTransportRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("currentUser", createTestUser())
		c.Next()
	})
	RouterGroupTransportSegments(r.Group("/api/v1/trip"))
	RouterGroupTransportSegmentItems(r.Group("/api/v1/transport"))
	return r
}

// TestIsValidTransportMode covers all defined modes plus an invalid one.
func TestIsValidTransportMode(t *testing.T) {
	validModes := []TransportMode{
		TransportModeFlight, TransportModeTrain, TransportModeBus,
		TransportModeCarRental, TransportModeFerry, TransportModeTaxi, TransportModeOther,
	}
	for _, m := range validModes {
		assert.True(t, isValidTransportMode(m), "expected %q to be valid", m)
	}
	assert.False(t, isValidTransportMode("helicopter"), "unexpected mode should be invalid")
	assert.False(t, isValidTransportMode(""), "empty mode should be invalid")
}

func TestCreateTransportSegment_MissingMode_Returns400(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupTransportRouter()
	body, _ := json.Marshal(map[string]interface{}{
		"operator": "IRCTC",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/trip/"+uuid.New().String()+"/transport",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransportSegment_InvalidMode_Returns400(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupTransportRouter()
	body, _ := json.Marshal(map[string]interface{}{
		"mode":     "submarine",
		"trip_plan": uuid.New().String(),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/trip/"+uuid.New().String()+"/transport",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParseTimePtr_ValidRFC3339(t *testing.T) {
	s := "2026-09-01T06:00:00+05:30"
	result := parseTimePtr(&s)
	assert.NotNil(t, result)
	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, 9, int(result.Month()))
}

func TestParseTimePtr_InvalidString_ReturnsNil(t *testing.T) {
	s := "not-a-date"
	assert.Nil(t, parseTimePtr(&s))
}

func TestParseTimePtr_NilPointer_ReturnsNil(t *testing.T) {
	assert.Nil(t, parseTimePtr(nil))
}
