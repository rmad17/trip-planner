package trips

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var claudeService = NewClaudeService()

// GenerateTripWithAI handles the AI-powered trip generation
// @Summary Generate trip using AI
// @Description Uses Claude AI to generate a complete trip plan with hops, days, and activities
// @Tags trips
// @Accept json
// @Produce json
// @Param request body TripGenerationRequest true "Trip generation parameters"
// @Success 200 {object} map[string]interface{} "Generated trip plan with preview"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/trip/generate [post]
func GenerateTripWithAI(c *gin.Context) {
	var req TripGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	// Set default currency if not provided
	if req.Currency == "" {
		req.Currency = CurrencyUSD
	}

	// Generate trip plan using Claude
	tripPlan, err := claudeService.GenerateTrip(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate trip plan: " + err.Error()})
		return
	}

	// Return the generated plan for preview (not saved yet)
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"trip_plan": tripPlan,
		"message":   "Trip plan generated successfully. Review and confirm to save.",
	})
}

// CreateTripFromAIGeneration creates a complete trip in the database from AI-generated plan
// @Summary Create trip from AI-generated plan
// @Description Saves an AI-generated trip plan to the database with all hops, days, and activities
// @Tags trips
// @Accept json
// @Produce json
// @Param request body TripGenerationResponse true "AI-generated trip plan to save"
// @Success 201 {object} map[string]interface{} "Created trip with ID"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/trip/generate/confirm [post]
func CreateTripFromAIGeneration(c *gin.Context) {
	var generatedPlan TripGenerationResponse
	if err := c.ShouldBindJSON(&generatedPlan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := currentUser.(uuid.UUID)

	// Start a transaction
	tx := core.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the main trip plan
	tripName := generatedPlan.TripName
	tripDesc := generatedPlan.Description
	travelMode := generatedPlan.RecommendedMode
	status := "planning"
	tripType := "leisure"

	// Parse start and end dates from first and last hop
	var startDate, endDate *time.Time
	if len(generatedPlan.Hops) > 0 {
		if t, err := time.Parse("2006-01-02", generatedPlan.Hops[0].StartDate); err == nil {
			startDate = &t
		}
		if t, err := time.Parse("2006-01-02", generatedPlan.Hops[len(generatedPlan.Hops)-1].EndDate); err == nil {
			endDate = &t
		}
	}

	totalDays := int8(generatedPlan.TotalDays)
	budget := generatedPlan.EstimatedBudget

	trip := TripPlan{
		BaseModel:   core.BaseModel{ID: uuid.New()},
		Name:        &tripName,
		Description: &tripDesc,
		StartDate:   startDate,
		EndDate:     endDate,
		MinDays:     &totalDays,
		MaxDays:     &totalDays,
		TravelMode:  &travelMode,
		TripType:    &tripType,
		Budget:      &budget,
		Currency:    CurrencyUSD, // Default, should get from request
		Status:      &status,
		UserID:      userID,
	}

	if err := tx.Create(&trip).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip plan"})
		return
	}

	// Map to store hop UUIDs for linking
	hopUUIDs := make(map[int]uuid.UUID)

	// Create hops
	for _, generatedHop := range generatedPlan.Hops {
		hopID := uuid.New()
		hopUUIDs[generatedHop.HopOrder] = hopID

		hopName := generatedHop.Name
		hopDesc := generatedHop.Description
		city := generatedHop.City
		country := generatedHop.Country
		transportation := generatedHop.Transportation
		hopOrder := generatedHop.HopOrder
		estimatedBudget := generatedHop.EstimatedBudget

		var hopStartDate, hopEndDate *time.Time
		if t, err := time.Parse("2006-01-02", generatedHop.StartDate); err == nil {
			hopStartDate = &t
		}
		if t, err := time.Parse("2006-01-02", generatedHop.EndDate); err == nil {
			hopEndDate = &t
		}

		hop := TripHop{
			BaseModel:       core.BaseModel{ID: hopID},
			Name:            &hopName,
			Description:     &hopDesc,
			City:            &city,
			Country:         &country,
			StartDate:       hopStartDate,
			EndDate:         hopEndDate,
			EstimatedBudget: &estimatedBudget,
			Transportation:  &transportation,
			POIs:            pq.StringArray(generatedHop.POIs),
			Restaurants:     pq.StringArray(generatedHop.Restaurants),
			Activities:      pq.StringArray(generatedHop.Activities),
			HopOrder:        &hopOrder,
			TripPlan:        trip.ID,
		}

		// Link to previous hop
		if generatedHop.HopOrder > 1 {
			if prevHopID, exists := hopUUIDs[generatedHop.HopOrder-1]; exists {
				hop.PreviousHop = &prevHopID
			}
		}

		if err := tx.Create(&hop).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hop"})
			return
		}

		// Update previous hop's NextHop reference
		if generatedHop.HopOrder > 1 {
			if prevHopID, exists := hopUUIDs[generatedHop.HopOrder-1]; exists {
				tx.Model(&TripHop{}).Where("id = ?", prevHopID).Update("next_hop", hopID)
			}
		}
	}

	// Map to store day UUIDs
	dayUUIDs := make(map[int]uuid.UUID)

	// Create trip days with activities
	for _, generatedDay := range generatedPlan.DailyItinerary {
		dayID := uuid.New()
		dayUUIDs[generatedDay.DayNumber] = dayID

		dayTitle := generatedDay.Title
		location := generatedDay.Location
		estimatedBudget := generatedDay.EstimatedBudget
		notes := generatedDay.Notes

		// Parse date
		dayDate, err := time.Parse("2006-01-02", generatedDay.Date)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format in day " + generatedDay.Date})
			return
		}

		// Validate day type
		dayType := TripDayType(generatedDay.DayType)
		if !IsValidTripDayType(string(dayType)) {
			dayType = TripDayTypeExplore // Default
		}

		// Find which hop this day belongs to
		var fromHopID *uuid.UUID
		for _, hop := range generatedPlan.Hops {
			hopStart, _ := time.Parse("2006-01-02", hop.StartDate)
			hopEnd, _ := time.Parse("2006-01-02", hop.EndDate)
			if (dayDate.Equal(hopStart) || dayDate.After(hopStart)) && (dayDate.Equal(hopEnd) || dayDate.Before(hopEnd)) {
				if hopUUID, exists := hopUUIDs[hop.HopOrder]; exists {
					fromHopID = &hopUUID
					break
				}
			}
		}

		day := TripDay{
			BaseModel:       core.BaseModel{ID: dayID},
			Date:            dayDate,
			DayNumber:       generatedDay.DayNumber,
			Title:           &dayTitle,
			DayType:         dayType,
			Notes:           &notes,
			StartLocation:   &location,
			EndLocation:     &location,
			EstimatedBudget: &estimatedBudget,
			TripPlan:        trip.ID,
			FromTripHop:     fromHopID,
			ToTripHop:       fromHopID,
		}

		if err := tx.Create(&day).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip day"})
			return
		}

		// Create activities for this day
		for _, generatedActivity := range generatedDay.Activities {
			activityName := generatedActivity.Name
			activityDesc := generatedActivity.Description
			activityLocation := generatedActivity.Location
			estimatedCost := generatedActivity.EstimatedCost
			duration := generatedActivity.Duration
			priority := int8(generatedActivity.Priority)
			tips := generatedActivity.Tips
			status := "planned"

			// Validate activity type
			activityType := ActivityType(generatedActivity.ActivityType)
			if !IsValidActivityType(string(activityType)) {
				activityType = ActivityTypeOther // Default
			}

			// Parse times
			var startTime, endTime *time.Time
			if generatedActivity.StartTime != "" {
				if t, err := time.Parse("15:04", generatedActivity.StartTime); err == nil {
					combinedStart := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(),
						t.Hour(), t.Minute(), 0, 0, dayDate.Location())
					startTime = &combinedStart
				}
			}
			if generatedActivity.EndTime != "" {
				if t, err := time.Parse("15:04", generatedActivity.EndTime); err == nil {
					combinedEnd := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(),
						t.Hour(), t.Minute(), 0, 0, dayDate.Location())
					endTime = &combinedEnd
				}
			}

			activity := Activity{
				BaseModel:     core.BaseModel{ID: uuid.New()},
				Name:          activityName,
				Description:   &activityDesc,
				ActivityType:  activityType,
				StartTime:     startTime,
				EndTime:       endTime,
				Duration:      &duration,
				Location:      &activityLocation,
				EstimatedCost: &estimatedCost,
				Priority:      &priority,
				Status:        &status,
				Notes:         &tips,
				TripDay:       dayID,
				TripHop:       fromHopID,
			}

			if err := tx.Create(&activity).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create activity"})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"trip_id": trip.ID,
		"message": "Trip created successfully from AI-generated plan",
	})
}

// GetMultiCitySuggestions provides AI-powered multi-city route suggestions
// @Summary Get multi-city suggestions
// @Description Get AI-powered suggestions for additional cities to include in a multi-city trip
// @Tags trips
// @Accept json
// @Produce json
// @Param source query string true "Source city"
// @Param destination query string true "Primary destination"
// @Param duration query int true "Trip duration in days"
// @Param preferences query string false "Comma-separated preferences (e.g., 'adventure,culture')"
// @Success 200 {object} map[string]interface{} "City suggestions"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/trip/suggest-cities [get]
func GetMultiCitySuggestions(c *gin.Context) {
	source := c.Query("source")
	destination := c.Query("destination")
	durationStr := c.Query("duration")
	preferencesStr := c.Query("preferences")

	if source == "" || destination == "" || durationStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source, destination, and duration are required"})
		return
	}

	duration := 0
	if _, err := fmt.Sscanf(durationStr, "%d", &duration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration must be a valid number"})
		return
	}

	var preferences []string
	if preferencesStr != "" {
		preferences = strings.Split(preferencesStr, ",")
	}

	suggestions, err := claudeService.SuggestMultiCityRoute(source, destination, duration, preferences)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get suggestions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"suggestions": suggestions,
	})
}
