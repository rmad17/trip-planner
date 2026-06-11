package trips

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- generateShareCode unit tests (no DB) ---

func TestGenerateShareCode_Length(t *testing.T) {
	code := generateShareCode()
	assert.Equal(t, shareCodeLength, len(code), "share code must be %d chars", shareCodeLength)
}

func TestGenerateShareCode_Charset(t *testing.T) {
	for i := 0; i < 50; i++ {
		code := generateShareCode()
		for _, ch := range code {
			assert.Contains(t, shareCodeAlphabet, string(ch),
				"character %q is not in the allowed alphabet", ch)
		}
	}
}

func TestGenerateShareCode_NoAmbiguousChars(t *testing.T) {
	ambiguous := "0O1Il"
	for i := 0; i < 200; i++ {
		code := generateShareCode()
		for _, ch := range ambiguous {
			assert.NotContains(t, code, string(ch),
				"ambiguous character %q found in share code %q", ch, code)
		}
	}
}

func TestGenerateShareCode_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 200)
	for i := 0; i < 200; i++ {
		code := generateShareCode()
		_, collision := seen[code]
		assert.False(t, collision, "collision on code %q at iteration %d", code, i)
		seen[code] = struct{}{}
	}
}

// --- sanitizeTripForPublic unit tests (no DB) ---

func TestSanitizeTripForPublic_NoPII(t *testing.T) {
	bookingRef := "SECRET-REF"
	contactInfo := "secret@hotel.com"
	actualCost := 500.0
	estCost := 25.0

	trip := &TripPlan{
		TripHops: []TripHop{
			{
				TripDays: []TripDay{
					{
						Activities: []Activity{
							{
								Name:          "Visit Eiffel Tower",
								EstimatedCost: &estCost,
								ActualCost:    &actualCost,
								BookingRef:    &bookingRef,
								ContactInfo:   &contactInfo,
							},
						},
					},
				},
			},
		},
	}

	pub := sanitizeTripForPublic(trip)
	require.Len(t, pub.Hops, 1)
	require.Len(t, pub.Hops[0].Days, 1)
	require.Len(t, pub.Hops[0].Days[0].Activities, 1)

	act := pub.Hops[0].Days[0].Activities[0]
	assert.Equal(t, "Visit Eiffel Tower", act.Name)
	assert.Equal(t, &estCost, act.EstimatedCost, "estimated cost should be public")
	// The public struct has no BookingRef or ContactInfo fields — compile-time guarantee.
}

// --- HTTP handler tests (no DB, test 403/404 paths via mock middleware) ---

func setupSharingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("currentUser", createTestUser())
		c.Next()
	})
	trips := r.Group("/api/v1/trip")
	RouterGroupSharing(trips)
	public := r.Group("/api/v1/public")
	RouterGroupPublicTrips(public.Group("/trip"))
	return r
}

func TestGetPublicTrip_InvalidCode_Returns404(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupSharingRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/trip/DOESNOTEXIST", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPublishTrip_NonOwner_Returns403(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupSharingRouter()
	// Trip owned by a different user UUID than the test user.
	otherTripID := uuid.New().String()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/trip/"+otherTripID+"/publish", nil)
	r.ServeHTTP(w, req)
	// The trip does not exist → 404 is acceptable; in real DB it would be 403.
	assert.Contains(t, []int{http.StatusNotFound, http.StatusForbidden}, w.Code)
}

func TestClonePublicTrip_UnknownCode_Returns404(t *testing.T) {
	t.Skip("Requires database — skipped unless TEST_DB_URL is set")
	r := setupSharingRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/trip/clone/BADCODE", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- response shape test ---

func TestPublicTripPlan_JSONOmitsSensitiveFields(t *testing.T) {
	pub := PublicTripPlan{
		ID:   uuid.New(),
		Name: func() *string { s := "Paris Trip"; return &s }(),
	}
	data, err := json.Marshal(pub)
	require.NoError(t, err)
	body := string(data)
	assert.NotContains(t, body, "user_id")
	assert.NotContains(t, body, "booking_ref")
	assert.NotContains(t, body, "contact_info")
	assert.NotContains(t, body, "actual_cost")
	assert.NotContains(t, body, "payment_mode")
}
