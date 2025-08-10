package trips

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

// CreateTrip godoc
// @Summary Create a new trip
// @Description Create a new trip plan with automatic creation of default hop and stay
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

	// Create default hop for the trip
	defaultHop := TripHop{
		Name:     newTrip.Name, // Use the same name as the trip
		TripPlan: tripPlan.BaseModel.ID,
	}

	result = core.DB.Create(&defaultHop)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Create default stay for the hop
	defaultStay := Stay{
		TripHop: defaultHop.BaseModel.ID,
	}

	result = core.DB.Create(&defaultStay)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip": tripPlan})
}

// GetTripsWithUser godoc
// @Summary Get all trips with user information
// @Description Retrieve all trips along with associated user data
// @Tags trips
// @Produce json
// @Success 200 {object} map[string]interface{} "List of trips with user information"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trips [get]
// func GetTripsWithUser(c *gin.Context) {
// 	var trips []TripPlan
// 	var tripsWithUsers []map[string]interface{}
//
// 	// Get all trips
// 	core.DB.Find(&trips)
//
// 	// Load user data for each trip
// 	for _, trip := range trips {
// 		var user accounts.User
// 		core.DB.First(&user, trip.UserID)
//
// 		tripWithUser := map[string]interface{}{
// 			"trip": trip,
// 			"user": user,
// 		}
// 		tripsWithUsers = append(tripsWithUsers, tripWithUser)
// 	}
//
// 	c.JSON(http.StatusOK, gin.H{"trips": tripsWithUsers})
// }
