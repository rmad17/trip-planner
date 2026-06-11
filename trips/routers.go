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
	router.POST("/generate", GenerateTripWithAI)                 // POST /trip/generate
	router.POST("/generate/confirm", CreateTripFromAIGeneration) // POST /trip/generate/confirm
	router.POST("/refine", RefineTripWithFeedback)               // POST /trip/refine
	router.GET("/suggest-cities", GetMultiCitySuggestions)       // GET /trip/suggest-cities

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

// RouterGroupSharing sets up authenticated trip sharing routes
func RouterGroupSharing(router *gin.RouterGroup) {
	router.POST("/:id/publish", PublishTrip)          // POST /trip/:id/publish
	router.POST("/:id/unpublish", UnpublishTrip)      // POST /trip/:id/unpublish
	router.POST("/:id/share/rotate", RotateShareCode) // POST /trip/:id/share/rotate
	router.POST("/clone/:share_code", ClonePublicTrip) // POST /trip/clone/:share_code
}

// RouterGroupPublicTrips sets up unauthenticated public trip routes
func RouterGroupPublicTrips(router *gin.RouterGroup) {
	router.GET("/:share_code", GetPublicTrip) // GET /public/trip/:share_code
}

// RouterGroupTripLifecycle sets up status and organizer endpoints
func RouterGroupTripLifecycle(router *gin.RouterGroup) {
	router.PATCH("/:id/status", UpdateTripStatus)     // PATCH /trip/:id/status
	router.POST("/:id/shift-dates", ShiftTripDates)   // POST /trip/:id/shift-dates
	router.GET("/:id/today", GetTodayView)            // GET /trip/:id/today
}

// RouterGroupTransportSegments sets up transport segments nested under trips
func RouterGroupTransportSegments(router *gin.RouterGroup) {
	router.GET("/:id/transport", ListTransportSegments)    // GET /trip/:id/transport
	router.POST("/:id/transport", CreateTransportSegment)  // POST /trip/:id/transport
}

// RouterGroupTransportSegmentItems sets up individual transport segment CRUD
func RouterGroupTransportSegmentItems(router *gin.RouterGroup) {
	router.GET("/:id", GetTransportSegment)       // GET /transport/:id
	router.PUT("/:id", UpdateTransportSegment)    // PUT /transport/:id
	router.DELETE("/:id", DeleteTransportSegment) // DELETE /transport/:id
}

// RouterGroupContacts sets up trip contacts nested under trips
func RouterGroupContacts(router *gin.RouterGroup) {
	router.GET("/:id/contacts", ListContacts)   // GET /trip/:id/contacts
	router.POST("/:id/contacts", CreateContact) // POST /trip/:id/contacts
}

// RouterGroupContactItems sets up individual contact CRUD
func RouterGroupContactItems(router *gin.RouterGroup) {
	router.PUT("/:id", UpdateContact)    // PUT /contacts/:id
	router.DELETE("/:id", DeleteContact) // DELETE /contacts/:id
}

// RouterGroupChecklist sets up checklist items nested under trips
func RouterGroupChecklist(router *gin.RouterGroup) {
	router.GET("/:id/checklist", ListChecklist)    // GET /trip/:id/checklist
	router.POST("/:id/checklist", CreateChecklistItem) // POST /trip/:id/checklist
}

// RouterGroupChecklistItems sets up individual checklist item CRUD
func RouterGroupChecklistItems(router *gin.RouterGroup) {
	router.PUT("/:id", UpdateChecklistItem)         // PUT /checklist/:id
	router.DELETE("/:id", DeleteChecklistItem)      // DELETE /checklist/:id
	router.PATCH("/:id/toggle", ToggleChecklistItem) // PATCH /checklist/:id/toggle
}
