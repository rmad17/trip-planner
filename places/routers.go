package places

import "github.com/gin-gonic/gin"

func RouterGroupPlacesAPI(router *gin.RouterGroup) {
	router.GET("/autocomplete/search", SearchAutocomplete)
	router.GET("/details", PlaceDetails)
}
