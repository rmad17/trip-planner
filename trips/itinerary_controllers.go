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
	tripID := c.Param("id")
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

	// Always get trip hops for additional context
	var tripHops []TripHop
	hopResult := core.DB.Where("trip_plan = ?", tripID).Order("hop_order ASC").Find(&tripHops)
	if hopResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": hopResult.Error.Error()})
		return
	}

	// Organize the response
	itinerary := make([]map[string]interface{}, 0)

	if len(tripDays) > 0 {
		// Process trip days as usual
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
				"date":             day.Date.Time.Format("2006-01-02"),
				"title":            day.Title,
				"day_type":         day.DayType,
				"notes":            day.Notes,
				"start_location":   day.StartLocation,
				"end_location":     day.EndLocation,
				"estimated_budget": day.EstimatedBudget,
				"actual_budget":    day.ActualBudget,
				"weather":          day.Weather,
				"activities":       activities,
				"activity_count":   len(activities),
			}

			// If no activities exist for this day, include hop information for context
			if len(activities) == 0 {
				// Find associated hop information
				var associatedHop *TripHop
				if day.FromTripHop != nil {
					for _, hop := range tripHops {
						if hop.BaseModel.ID == *day.FromTripHop {
							associatedHop = &hop
							break
						}
					}
				}

				if associatedHop != nil {
					dayItinerary["hop_info"] = map[string]interface{}{
						"hop_id":             associatedHop.BaseModel.ID,
						"hop_name":           associatedHop.Name,
						"hop_description":    associatedHop.Description,
						"city":               associatedHop.City,
						"country":            associatedHop.Country,
						"region":             associatedHop.Region,
						"pois":               associatedHop.POIs,
						"restaurants":        associatedHop.Restaurants,
						"planned_activities": associatedHop.Activities,
					}
				}
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
	} else if len(tripHops) > 0 {
		// If no trip days exist, create basic itinerary from trip hops
		for i, hop := range tripHops {
			var hopStartDate string
			var hopEndDate string

			if hop.StartDate != nil {
				hopStartDate = hop.StartDate.Format("2006-01-02")
			}
			if hop.EndDate != nil {
				hopEndDate = hop.EndDate.Format("2006-01-02")
			}

			hopItinerary := map[string]interface{}{
				"day_number":                    i + 1, // Sequential numbering based on hop order
				"date":                          hopStartDate,
				"end_date":                      hopEndDate,
				"title":                         hop.Name,
				"day_type":                      "explore", // Default type when generating from hops
				"notes":                         hop.Notes,
				"start_location":                hop.City,
				"end_location":                  hop.City,
				"estimated_budget":              hop.EstimatedBudget,
				"actual_budget":                 hop.ActualSpent,
				"weather":                       nil,
				"activities":                    []interface{}{}, // Empty activities array
				"activity_count":                0,
				"hop_id":                        hop.BaseModel.ID,
				"hop_name":                      hop.Name,
				"hop_description":               hop.Description,
				"city":                          hop.City,
				"country":                       hop.Country,
				"region":                        hop.Region,
				"pois":                          hop.POIs,
				"restaurants":                   hop.Restaurants,
				"planned_activities":            hop.Activities,
				"total_estimated_activity_cost": 0.0,
				"total_actual_activity_cost":    0.0,
			}

			itinerary = append(itinerary, hopItinerary)
		}
	}

	// Calculate summary
	summary := map[string]interface{}{
		"total_days": len(itinerary),
		"trip_id":    tripID,
		"trip_name":  tripPlan.Name,
	}

	// Add date range and data source information
	if len(tripDays) > 0 {
		summary["start_date"] = tripDays[0].Date.Time.Format("2006-01-02")
		summary["end_date"] = tripDays[len(tripDays)-1].Date.Time.Format("2006-01-02")
		summary["data_source"] = "trip_days"
		summary["has_detailed_days"] = true

		// Count days with no activities
		daysWithoutActivities := 0
		for _, day := range tripDays {
			if len(day.Activities) == 0 {
				daysWithoutActivities++
			}
		}
		summary["days_without_activities"] = daysWithoutActivities
	} else if len(tripHops) > 0 {
		// Get date range from hops
		if tripHops[0].StartDate != nil {
			summary["start_date"] = tripHops[0].StartDate.Format("2006-01-02")
		}
		if tripHops[len(tripHops)-1].EndDate != nil {
			summary["end_date"] = tripHops[len(tripHops)-1].EndDate.Format("2006-01-02")
		}
		summary["data_source"] = "trip_hops"
		summary["has_detailed_days"] = false
	} else {
		summary["data_source"] = "none"
		summary["has_detailed_days"] = false
	}

	// Always include hop information in summary
	summary["total_hops"] = len(tripHops)

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
	tripID := c.Param("id")
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
		"date":             tripDay.Date.Time.Format("2006-01-02"),
		"title":            tripDay.Title,
		"day_type":         tripDay.DayType,
		"notes":            tripDay.Notes,
		"start_location":   tripDay.StartLocation,
		"end_location":     tripDay.EndLocation,
		"estimated_budget": tripDay.EstimatedBudget,
		"actual_budget":    tripDay.ActualBudget,
		"weather":          tripDay.Weather,
		"activities":       activities,
		"summary": gin.H{
			"activity_count":                len(activities),
			"total_estimated_activity_cost": totalEstimated,
			"total_actual_activity_cost":    totalActual,
			"trip_id":                       tripID,
			"trip_name":                     tripPlan.Name,
		},
	}

	c.JSON(http.StatusOK, response)
}
