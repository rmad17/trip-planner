package trips

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTrip(c *gin.Context) {
	var newTrip CreateTripRequest

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newTrip); err != nil {
		return
	}

	// Add the new album to the slice.
	c.IndentedJSON(http.StatusCreated, newTrip)

}
