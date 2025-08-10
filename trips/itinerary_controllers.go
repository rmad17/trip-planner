package trips

import (
	"net/http"
	"strconv"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

// GetDailyItinerary godoc
// @Summary Get daily itinerary for a trip
// @Description Get activities and itinerary for all days of a trip, organized by date
// @Tags itinerary
// @Produce json
// @Param trip_id path string true "Trip Plan ID"
// @Param day query int false "Specific day number to get (optional)"
// @Param date query string false "Specific date to get (YYYY-MM-DD format, optional)"
// @Success 200 {object} map[string]interface{} "Daily itinerary"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_id}/itinerary [get]
func GetDailyItinerary(c *gin.Context) {
	tripID := c.Param("trip_id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan ownership
	var tripPlan TripPlan
	result := core.DB.Where("id = ? AND user_id = ?", tripID, user.BaseModel.ID).First(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	// Parse optional filters
	dayParam := c.Query("day")
	dateParam := c.Query("date")

	// Build query for trip days
	query := core.DB.Preload("Activities").Where("trip_plan = ?", tripID)

	// Apply filters
	if dayParam != "" {
		if dayNumber, err := strconv.Atoi(dayParam); err == nil {
			query = query.Where("day_number = ?", dayNumber)
		}
	}

	if dateParam != "" {
		if date, err := time.Parse("2006-01-02", dateParam); err == nil {
			query = query.Where("date = ?", date)
		}
	}

	// Get trip days with activities
	var tripDays []TripDay
	result = query.Order("day_number ASC, date ASC").Find(&tripDays)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Organize the response
	itinerary := make([]map[string]interface{}, 0)
	
	for _, day := range tripDays {
		// Sort activities by start time
		activities := day.Activities
		// Simple bubble sort by start time (for small arrays this is fine)
		for i := 0; i < len(activities); i++ {
			for j := 0; j < len(activities)-1-i; j++ {
				if activities[j].StartTime != nil && activities[j+1].StartTime != nil {
					if activities[j].StartTime.After(*activities[j+1].StartTime) {
						activities[j], activities[j+1] = activities[j+1], activities[j]
					}
				}
			}
		}

		dayItinerary := map[string]interface{}{
			"day_number":       day.DayNumber,
			"date":            day.Date.Format("2006-01-02"),
			"title":           day.Title,
			"day_type":        day.DayType,
			"notes":           day.Notes,
			"start_location":  day.StartLocation,
			"end_location":    day.EndLocation,
			"estimated_budget": day.EstimatedBudget,
			"actual_budget":   day.ActualBudget,
			"weather":         day.Weather,
			"activities":      activities,
			"activity_count":  len(activities),
		}

		// Calculate total estimated and actual costs for the day
		var totalEstimated, totalActual float64
		for _, activity := range activities {
			if activity.EstimatedCost != nil {
				totalEstimated += *activity.EstimatedCost
			}
			if activity.ActualCost != nil {
				totalActual += *activity.ActualCost
			}
		}

		dayItinerary["total_estimated_activity_cost"] = totalEstimated
		dayItinerary["total_actual_activity_cost"] = totalActual

		itinerary = append(itinerary, dayItinerary)
	}

	// Calculate summary
	summary := map[string]interface{}{
		"total_days": len(itinerary),
		"trip_id":    tripID,
		"trip_name":  tripPlan.Name,
	}

	// Add date range if we have days
	if len(tripDays) > 0 {
		summary["start_date"] = tripDays[0].Date.Format("2006-01-02")
		summary["end_date"] = tripDays[len(tripDays)-1].Date.Format("2006-01-02")
	}

	response := gin.H{
		"itinerary": itinerary,
		"summary":   summary,
	}

	c.JSON(http.StatusOK, response)
}

// GetDayItinerary godoc
// @Summary Get itinerary for a specific day
// @Description Get detailed itinerary for a specific trip day including all activities
// @Tags itinerary
// @Produce json
// @Param trip_id path string true "Trip Plan ID"
// @Param day_number path int true "Day number"
// @Success 200 {object} map[string]interface{} "Day itinerary"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_id}/itinerary/day/{day_number} [get]
func GetDayItinerary(c *gin.Context) {
	tripID := c.Param("trip_id")
	dayNumberParam := c.Param("day_number")
	
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Parse day number
	dayNumber, err := strconv.Atoi(dayNumberParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day number"})
		return
	}

	// Verify trip plan ownership
	var tripPlan TripPlan
	result := core.DB.Where("id = ? AND user_id = ?", tripID, user.BaseModel.ID).First(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	// Get specific trip day with activities
	var tripDay TripDay
	result = core.DB.Preload("Activities").
		Where("trip_plan = ? AND day_number = ?", tripID, dayNumber).
		First(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	// Sort activities by start time
	activities := tripDay.Activities
	for i := 0; i < len(activities); i++ {
		for j := 0; j < len(activities)-1-i; j++ {
			if activities[j].StartTime != nil && activities[j+1].StartTime != nil {
				if activities[j].StartTime.After(*activities[j+1].StartTime) {
					activities[j], activities[j+1] = activities[j+1], activities[j]
				}
			}
		}
	}

	// Calculate total costs
	var totalEstimated, totalActual float64
	for _, activity := range activities {
		if activity.EstimatedCost != nil {
			totalEstimated += *activity.EstimatedCost
		}
		if activity.ActualCost != nil {
			totalActual += *activity.ActualCost
		}
	}

	response := gin.H{
		"day_number":       tripDay.DayNumber,
		"date":            tripDay.Date.Format("2006-01-02"),
		"title":           tripDay.Title,
		"day_type":        tripDay.DayType,
		"notes":           tripDay.Notes,
		"start_location":  tripDay.StartLocation,
		"end_location":    tripDay.EndLocation,
		"estimated_budget": tripDay.EstimatedBudget,
		"actual_budget":   tripDay.ActualBudget,
		"weather":         tripDay.Weather,
		"activities":      activities,
		"summary": gin.H{
			"activity_count":              len(activities),
			"total_estimated_activity_cost": totalEstimated,
			"total_actual_activity_cost":   totalActual,
			"trip_id":                     tripID,
			"trip_name":                   tripPlan.Name,
		},
	}

	c.JSON(http.StatusOK, response)
}