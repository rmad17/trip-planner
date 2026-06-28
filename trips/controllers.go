package trips

import (
	"fmt"
	"net/http"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateTrip godoc
// @Summary Create a new trip
// @Description Create a new trip plan with automatic creation of default hop, stay, and daily plans
// @Tags trips
// @Accept json
// @Produce json
// @Param trip body CreateTripRequest true "Trip creation request"
// @Success 201 {object} map[string]interface{} "Trip created successfully"
// @Failure 400 {object} map[string]string "Bad request - validation errors"
// @Failure 401 {object} map[string]string "Unauthorized - user not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trips/create [post]
func CreateTrip(c *gin.Context) {
	var newTrip CreateTripRequest

	if err := c.BindJSON(&newTrip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	tripPlan := TripPlan{
		Name:        newTrip.Name,
		StartDate:   newTrip.StartDate,
		EndDate:     newTrip.EndDate,
		TravelModes: newTrip.ParsedTravelModes(),
		Notes:       newTrip.Notes,
		Hotels:      newTrip.Hotels,
		Tags:        newTrip.Tags,
		UserID:      user.ID,
	}

	if newTrip.MinDays != nil {
		minDays := int8(*newTrip.MinDays)
		tripPlan.MinDays = &minDays
	}

	if err := core.DB.Create(&tripPlan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Default hop
	defaultHop := TripHop{
		Name:     newTrip.Name,
		TripPlan: tripPlan.ID,
	}
	if err := core.DB.Create(&defaultHop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Default stay for the hop
	if err := core.DB.Create(&Stay{TripHop: defaultHop.ID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-create one TripDay per travel date so the frontend has days to work with immediately.
	// The frontend also attempts this after creation; the backend ensures it happens even if
	// the frontend call fails, and skips dates that already exist.
	if newTrip.StartDate != nil && newTrip.EndDate != nil {
		createDefaultTripDays(tripPlan.ID, *newTrip.StartDate, *newTrip.EndDate)
	}

	c.JSON(http.StatusCreated, gin.H{"trip": tripPlan})
}

// createDefaultTripDays inserts one TripDay per calendar date in [startDate, endDate].
// Existing days for the trip are left untouched (insert is skipped on conflict).
func createDefaultTripDays(tripPlanID uuid.UUID, startDate, endDate time.Time) {
	start := startDate.UTC().Truncate(24 * time.Hour)
	end := endDate.UTC().Truncate(24 * time.Hour)

	dayNumber := 1
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dt := d
		title := fmt.Sprintf("Day %d", dayNumber)
		day := TripDay{
			SoftDeleteModel: core.SoftDeleteModel{BaseModel: core.BaseModel{ID: uuid.New()}},
			Date:            &core.Date{Time: dt},
			DayNumber:       dayNumber,
			Title:           &title,
			DayType:         TripDayTypeExplore,
			TripPlan:        tripPlanID,
		}
		// Use OnConflict to skip if a day with the same trip_plan+day_number already exists
		core.DB.Where(TripDay{TripPlan: tripPlanID, DayNumber: dayNumber}).
			FirstOrCreate(&day)
		dayNumber++
	}
}
