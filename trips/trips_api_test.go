package trips

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"triplanner/accounts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func setupTestDB() {
	// Set test environment
	os.Setenv("APP_ENV", "test")

	// For now, skip actual database connection in tests
	// This would require a real test database to be set up
	// TODO: Set up proper test database or use in-memory database for tests
}

func createTestUser() accounts.User {
	user := accounts.User{
		Email:    stringPtr("test@example.com"),
		Username: "testuser",
		Password: "testpassword",
	}
	user.BaseModel.ID = uuid.New()
	// TODO: Create user in test database when database is properly set up
	// core.DB.Create(&user)
	return user
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func int16Ptr(i int16) *int16 {
	return &i
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set test user
	router.Use(func(c *gin.Context) {
		user := createTestUser()
		c.Set("currentUser", user)
		c.Next()
	})

	v1 := router.Group("/api/v1")
	RouterGroupCreateTrip(v1.Group("/trips"))

	return router
}

// Positive Test Cases

func TestCreateTrip_Success(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	tripRequest := CreateTripRequest{
		Name:       stringPtr("Trip to Paris"),
		StartDate:  timePtr(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:    timePtr(time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)),
		MinDays:    int16Ptr(7),
		TravelMode: stringPtr("flight"),
		Notes:      stringPtr("Romantic getaway"),
		Hotels:     pq.StringArray{"Hotel de Paris", "Le Bristol"},
		Tags:       pq.StringArray{"romantic", "europe", "culture"},
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	trip, exists := response["trip"]
	if !exists {
		t.Error("Expected 'trip' field in response")
	}

	tripData := trip.(map[string]interface{})
	if tripData["name"] != "Trip to Paris" {
		t.Errorf("Expected trip name 'Trip to Paris', got %v", tripData["name"])
	}

	// TODO: Verify default hop was created when database is properly set up
	// var hopCount int64
	// core.DB.Model(&TripHop{}).Count(&hopCount)
	// if hopCount != 1 {
	// 	t.Errorf("Expected 1 hop to be created, got %d", hopCount)
	// }

	// TODO: Verify default stay was created when database is properly set up
	// var stayCount int64
	// core.DB.Model(&Stay{}).Count(&stayCount)
	// if stayCount != 1 {
	// 	t.Errorf("Expected 1 stay to be created, got %d", stayCount)
	// }
}

func TestCreateTrip_MinimalData(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	tripRequest := CreateTripRequest{
		Name: stringPtr("Quick Trip"),
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// TODO: Verify hop and stay were still created with minimal data when database is properly set up
	// var hopCount, stayCount int64
	// core.DB.Model(&TripHop{}).Count(&hopCount)
	// core.DB.Model(&Stay{}).Count(&stayCount)

	// if hopCount != 1 || stayCount != 1 {
	// 	t.Error("Expected default hop and stay to be created even with minimal trip data")
	// }
}

func TestCreateTrip_WithAllFields(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	tripRequest := CreateTripRequest{
		Name:       stringPtr("Complete Trip"),
		StartDate:  timePtr(time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:    timePtr(time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)),
		MinDays:    int16Ptr(10),
		TravelMode: stringPtr("car"),
		Notes:      stringPtr("Road trip across the country"),
		Hotels:     pq.StringArray{"Motel 6", "Holiday Inn", "Best Western"},
		Tags:       pq.StringArray{"adventure", "road-trip", "family"},
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	tripData := response["trip"].(map[string]interface{})
	if tripData["travel_mode"] != "car" {
		t.Errorf("Expected travel_mode 'car', got %v", tripData["travel_mode"])
	}
	if tripData["notes"] != "Road trip across the country" {
		t.Errorf("Expected specific notes, got %v", tripData["notes"])
	}
}

// Negative Test Cases

func TestCreateTrip_MissingName(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	tripRequest := CreateTripRequest{
		StartDate: timePtr(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:   timePtr(time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)),
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if _, exists := response["error"]; !exists {
		t.Error("Expected 'error' field in response for missing required field")
	}

	// TODO: Verify no trip, hop, or stay was created when database is properly set up
	// var tripCount, hopCount, stayCount int64
	// core.DB.Model(&TripPlan{}).Count(&tripCount)
	// core.DB.Model(&TripHop{}).Count(&hopCount)
	// core.DB.Model(&Stay{}).Count(&stayCount)

	// if tripCount != 0 || hopCount != 0 || stayCount != 0 {
	// 	t.Error("Expected no records to be created when validation fails")
	// }
}

func TestCreateTrip_InvalidJSON(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	invalidJSON := `{"name": "Test Trip", "start_date": "invalid-date"}`

	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if _, exists := response["error"]; !exists {
		t.Error("Expected 'error' field in response for invalid JSON")
	}
}

func TestCreateTrip_MalformedJSON(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	malformedJSON := `{"name": "Test Trip", "start_date":}`

	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBufferString(malformedJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateTrip_EmptyBody(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateTrip_NoAuthUser(t *testing.T) {
	setupTestDB()

	// Create router without auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RouterGroupCreateTrip(v1.Group("/trips"))

	tripRequest := CreateTripRequest{
		Name: stringPtr("Unauthorized Trip"),
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "User not found" {
		t.Errorf("Expected 'User not found' error, got %v", response["error"])
	}
}

func TestCreateTrip_LargeMinDays(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	tripRequest := CreateTripRequest{
		Name:    stringPtr("Long Trip"),
		MinDays: int16Ptr(32767), // Max int16 value
	}

	jsonData, _ := json.Marshal(tripRequest)
	req, _ := http.NewRequest("POST", "/api/v1/trips/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// TODO: Verify the conversion from int16 to int8 works when database is properly set up
	// var trip TripPlan
	// core.DB.First(&trip)

	// if trip.MinDays == nil {
	// 	t.Error("Expected MinDays to be set")
	// } else if *trip.MinDays != 127 { // int8 max value after truncation
	// 	t.Errorf("Expected MinDays to be 127 (truncated), got %d", *trip.MinDays)
	// }
}
