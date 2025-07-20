package trips

import "github.com/gin-gonic/gin"

func RouterGroupCreateTrip(router *gin.RouterGroup) {
	router.POST("/create", CreateTrip)
	router.GET("", GetTripsWithUser) // GET /trips to list all trips
}
