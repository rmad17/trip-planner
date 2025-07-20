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
	trips.RouterGroupCreateTrip(v1.Group("/trips"))
	accounts.RouterGroupUserProfile(v1.Group("/user"))

	router.Run() // listen and serve on 0.0.0.0:8080
}
