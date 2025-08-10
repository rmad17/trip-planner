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
	"triplanner/accounts"
	"triplanner/core"
	_ "triplanner/docs" // This line is necessary for go-swagger to find your docs!
	"triplanner/expenses"
	"triplanner/places"
	"triplanner/trips"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Initialize GoAdmin
	// admin.SimpleSetupGoAdmin(router)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := router.Group("/api/v1")
	accounts.RouterGroupUserAuth(v1.Group("/auth"))
	accounts.RouterGroupGoogleOAuth(v1.Group("/auth"))

	v1.Use(accounts.CheckAuth)
	places.RouterGroupPlacesAPI(v1.Group("/places"))
	accounts.RouterGroupUserProfile(v1.Group("/user"))

	// Trip Management Routes
	trips.RouterGroupCreateTrip(v1.Group("/trips"))      // Original routes (backward compatibility)
	trips.RouterGroupTripPlans(v1.Group("/trip-plans"))  // Comprehensive CRUD for trip plans
	trips.RouterGroupTripHops(v1.Group("/trip-hops"))    // Individual trip hop operations
	trips.RouterGroupTripDays(v1.Group("/trip-days"))    // Individual trip day operations
	trips.RouterGroupActivities(v1.Group("/activities")) // Individual activity operations
	trips.RouterGroupTravellers(v1.Group("/travellers")) // Individual traveller operations
	trips.RouterGroupStays(v1.Group("/stays"))           // Individual stay operations

	// Expense Management Routes
	expenses.RouterGroupExpenses(v1.Group("/trip-plans"))          // Expenses nested under trip plans
	expenses.RouterGroupExpenseItems(v1.Group("/expenses"))        // Individual expense operations
	expenses.RouterGroupExpenseSplits(v1.Group("/expense-splits")) // Expense split operations

	router.Run() // listen and serve on 0.0.0.0:8080
}
