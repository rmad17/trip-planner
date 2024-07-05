package main

import (
	"triplanner/controllers"
	"triplanner/core"
	"triplanner/middlewares"

	"github.com/gin-gonic/gin"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()

}

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/auth/signup", controllers.CreateUser)
	router.POST("/auth/login", controllers.Login)
	router.GET("/user/profile", middlewares.CheckAuth, controllers.GetUserProfile)
	router.Run() // listen and serve on 0.0.0.0:8080
}
