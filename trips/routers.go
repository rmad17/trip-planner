package trips

import "github.com/gin-gonic/gin"

// RouterGroupCreateTrip sets up the original trip creation routes (backward compatibility)
func RouterGroupCreateTrip(router *gin.RouterGroup) {
	router.POST("/create", CreateTrip)
	router.GET("", GetTripPlans) // GET /trip-plans
}

// RouterGroupTripPlans sets up comprehensive CRUD routes for trip plans
func RouterGroupTripPlans(router *gin.RouterGroup) {
	// AI-powered trip generation
	router.POST("/generate", GenerateTripWithAI)              // POST /trip/generate
	router.POST("/generate/confirm", CreateTripFromAIGeneration) // POST /trip/generate/confirm
	router.GET("/suggest-cities", GetMultiCitySuggestions)    // GET /trip/suggest-cities

	// Trip Plans CRUD
	router.GET("/:id", GetTripPlan)                  // GET /trip-plans/:id
	router.GET("/:id/complete", GetTripPlanComplete) // GET /trip-plans/:id/complete
	router.PUT("/:id", UpdateTripPlan)               // PUT /trip-plans/:id
	router.DELETE("/:id", DeleteTripPlan)            // DELETE /trip-plans/:id

	// Trip Hops nested under Trip Plans
	router.GET("/:id/hops", GetTripHops)    // GET /trip-plans/:id/hops
	router.POST("/:id/hops", CreateTripHop) // POST /trip-plans/:id/hops

	// Trip Days nested under Trip Plans
	router.GET("/:id/days", GetTripDays)    // GET /trip-plans/:id/days
	router.POST("/:id/days", CreateTripDay) // POST /trip-plans/:id/days

	// Travellers nested under Trip Plans
	router.GET("/:id/travellers", GetTravellers)           // GET /trip-plans/:id/travellers
	router.POST("/:id/travellers", CreateTraveller)        // POST /trip-plans/:id/travellers
	router.POST("/:id/travellers/invite", InviteTraveller) // POST /trip-plans/:id/travellers/invite

	// Activities nested under Trip Days
	router.GET("/:id/activities", GetActivities)   // GET /trip-plans/:id/days/:day_id/activities
	router.POST("/:id/activities", CreateActivity) // POST /trip-plans/:id/days/:day_id/activities

	// Itinerary endpoints
	router.GET("/:id/itinerary", GetDailyItinerary)               // GET /trip-plans/:id/itinerary
	router.GET("/:id/itinerary/day/:day_number", GetDayItinerary) // GET /trip-plans/:id/itinerary/day/:day_number
}

// RouterGroupTripHops sets up CRUD routes for individual trip hops
func RouterGroupTripHops(router *gin.RouterGroup) {
	router.PUT("/:id", UpdateTripHop)    // PUT /hops/:id
	router.DELETE("/:id", DeleteTripHop) // DELETE /hops/:id

	// Stays nested under Trip Hops
	router.GET("/:id/stays", GetStays)    // GET /hops/:id/stays
	router.POST("/:id/stays", CreateStay) // POST /hops/:id/stays
}

// RouterGroupTripDays sets up CRUD routes for individual trip days
func RouterGroupTripDays(router *gin.RouterGroup) {
	router.GET("/:id", GetTripDay)       // GET /days/:id
	router.PUT("/:id", UpdateTripDay)    // PUT /days/:id
	router.DELETE("/:id", DeleteTripDay) // DELETE /days/:id
}

// RouterGroupActivities sets up CRUD routes for individual activities
func RouterGroupActivities(router *gin.RouterGroup) {
	router.PUT("/:id", UpdateActivity)    // PUT /activities/:id
	router.DELETE("/:id", DeleteActivity) // DELETE /activities/:id
}

// RouterGroupTravellers sets up CRUD routes for individual travellers
func RouterGroupTravellers(router *gin.RouterGroup) {
	router.GET("/:id", GetTraveller)       // GET /travellers/:id
	router.PUT("/:id", UpdateTraveller)    // PUT /travellers/:id
	router.DELETE("/:id", DeleteTraveller) // DELETE /travellers/:id
}

// RouterGroupStays sets up CRUD routes for individual stays
func RouterGroupStays(router *gin.RouterGroup) {
	router.GET("/:id", GetStay)       // GET /stays/:id
	router.PUT("/:id", UpdateStay)    // PUT /stays/:id
	router.DELETE("/:id", DeleteStay) // DELETE /stays/:id
}
