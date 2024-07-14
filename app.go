package main

import (
	"os"
	"triplanner/controllers"
	"triplanner/core"
	"triplanner/middlewares"
	"triplanner/places"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()

}

func main() {
	google_client_id := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	google_provider := google.New(google_client_id, google_client_secret, "http://localhost:8080/auth/google/callback", "email", "profile")
	goth.UseProviders(google_provider)

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/maps/autocomplete", places.SearchAutocomplete)
	router.POST("/auth/signup", controllers.CreateUser)
	router.POST("/auth/login", controllers.Login)
	router.GET("/user/profile", middlewares.CheckAuth, controllers.GetUserProfile)
	router.GET("/auth/:provider/begin", controllers.GoogleOAuthBegin)

	router.GET("/places/autocomplete/search", places.SearchAutocomplete)

	router.LoadHTMLGlob("templates/*")
	router.GET("/auth/google/login", controllers.GoogleOAuthLogin)
	router.GET("/auth/:provider/callback", controllers.GoogleOAuthCallback)
	router.Run() // listen and serve on 0.0.0.0:8080
}
