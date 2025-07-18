package trips

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

func CreateTrip(c *gin.Context) {
	var newTrip CreateTripRequest

	if err := c.BindJSON(&newTrip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from middleware
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	user := currentUser.(accounts.User)

	// Create trip plan
	tripPlan := TripPlan{
		Name:       newTrip.Name,
		StartDate:  newTrip.StartDate,
		EndDate:    newTrip.EndDate,
		TravelMode: newTrip.TravelMode,
		Notes:      newTrip.Notes,
		Hotels:     newTrip.Hotels,
		Tags:       newTrip.Tags,
		UserID:     user.BaseModel.ID,
	}

	// Handle MinDays conversion if provided
	if newTrip.MinDays != nil {
		minDays := int8(*newTrip.MinDays) // Convert int16 to int8
		tripPlan.MinDays = &minDays
	}

	// Save to database
	result := core.DB.Create(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip": tripPlan})
}

// GetTripsWithUser - Example of loading related data
func GetTripsWithUser(c *gin.Context) {
	var trips []TripPlan
	var tripsWithUsers []map[string]interface{}

	// Get all trips
	core.DB.Find(&trips)

	// Load user data for each trip
	for _, trip := range trips {
		var user accounts.User
		core.DB.First(&user, trip.UserID)

		tripWithUser := map[string]interface{}{
			"trip": trip,
			"user": user,
		}
		tripsWithUsers = append(tripsWithUsers, tripWithUser)
	}

	c.JSON(http.StatusOK, gin.H{"trips": tripsWithUsers})
}
