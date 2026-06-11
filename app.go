// @title Trip Planner API
// @description API for managing trip plans, hops, and stays
// @version 1.0
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"triplanner/accounts"
	"triplanner/core"
	_ "triplanner/docs" // This line is necessary for go-swagger to find your docs!
	"triplanner/documents"
	"triplanner/expenses"
	"triplanner/flights"
	"triplanner/hotels"
	"triplanner/middlewares"
	"triplanner/places"
	"triplanner/routes"
	"triplanner/storage"
	"triplanner/travelknowledge"
	"triplanner/trips"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()

	// Initialize storage provider
	if err := storage.InitializeStorage(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
}

func main() {
	router := gin.Default()

	// CORS middleware - must be applied before routes
	router.Use(middlewares.CORSMiddleware())

	router.LoadHTMLGlob("templates/*")

	// Initialize GoAdmin
	// admin.SimpleSetupGoAdmin(router)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint for Caddy and monitoring
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Trip Planner API is running",
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := router.Group("/api/v1")
	accounts.RouterGroupUserAuth(v1.Group("/auth"))
	accounts.RouterGroupGoogleOAuth(v1.Group("/auth"))
	// Password reset & email verification — no auth required
	accounts.RouterGroupPasswordReset(v1.Group("/auth"))

	// Public trip share — no auth required
	publicV1 := router.Group("/api/v1/public")
	trips.RouterGroupPublicTrips(publicV1.Group("/trip"))

	v1.Use(accounts.CheckAuth)
	places.RouterGroupPlacesAPI(v1.Group("/places"))
	accounts.RouterGroupUserProfile(v1.Group("/user"))

	// Trip Management Routes
	trips.RouterGroupCreateTrip(v1.Group("/trips"))      // Original routes (backward compatibility)
	trips.RouterGroupTripPlans(v1.Group("/trip"))        // Comprehensive CRUD for trip plans
	trips.RouterGroupTripHops(v1.Group("/hops"))         // Individual trip hop operations
	trips.RouterGroupTripDays(v1.Group("/days"))         // Individual trip day operations
	trips.RouterGroupActivities(v1.Group("/activities")) // Individual activity operations
	trips.RouterGroupTravellers(v1.Group("/travellers")) // Individual traveller operations
	trips.RouterGroupStays(v1.Group("/stays"))           // Individual stay operations

	// Expense Management Routes
	expenses.RouterGroupExpenses(v1.Group("/trip"))                // Expenses nested under trip plans
	expenses.RouterGroupExpenseItems(v1.Group("/expenses"))        // Individual expense operations
	expenses.RouterGroupExpenseSplits(v1.Group("/expense-splits")) // Expense split operations

	// Document Management Routes
	documents.RouterGroupDocuments(v1.Group("/trip"))          // Documents nested under trip plans
	documents.RouterGroupDocumentItems(v1.Group("/documents")) // Individual document operations

	// Phase 2 — Trip sharing
	trips.RouterGroupSharing(v1.Group("/trip"))

	// Phase 3 — Trip lifecycle, transport, contacts, checklist
	trips.RouterGroupTripLifecycle(v1.Group("/trip"))
	trips.RouterGroupTransportSegments(v1.Group("/trip"))
	trips.RouterGroupTransportSegmentItems(v1.Group("/transport"))
	trips.RouterGroupContacts(v1.Group("/trip"))
	trips.RouterGroupContactItems(v1.Group("/contacts"))
	trips.RouterGroupChecklist(v1.Group("/trip"))
	trips.RouterGroupChecklistItems(v1.Group("/checklist"))

	// AI provider search endpoints (real hotel/flight offers, map routes)
	hotels.RouterGroupHotels(v1.Group("/search/hotels"))
	flights.RouterGroupFlights(v1.Group("/search/flights"))
	routes.RouterGroupRoutes(v1.Group("/routes"))

	// Travel Knowledge Routes (public - no auth required for reading travel docs)
	travelKnowledgePublic := router.Group("/api/v1/travel-knowledge")
	travelknowledge.RouterGroupTravelKnowledge(travelKnowledgePublic)

	if err := router.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
