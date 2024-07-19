package main

import (
	"triplanner/accounts"
	"triplanner/core"
	"triplanner/places"
	"triplanner/trips"

	"github.com/gin-gonic/gin"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()

}

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	v1 := router.Group("/api/v1")
	accounts.RouterGroupUserAuth(v1.Group("/auth"))
	accounts.RouterGroupGoogleOAuth(v1.Group("/auth"))

	places.RouterGroupPlacesAPI(v1.Group("/places"))

	trips.RouterGroupCreateTrip(v1.Group("/trips"))

	v1.Use(accounts.CheckAuth)
	accounts.RouterGroupUserProfile(v1.Group("/user"))

	router.Run() // listen and serve on 0.0.0.0:8080
}
