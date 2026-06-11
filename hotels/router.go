package hotels

import "github.com/gin-gonic/gin"

// RouterGroupHotels registers hotel search handlers on the supplied group.
func RouterGroupHotels(router *gin.RouterGroup) {
	router.GET("", SearchHotelsHandler)
}

// DefaultProvider returns the configured default provider for use by other packages.
func DefaultProvider() (Provider, error) {
	return defaultFactory.GetDefaultProvider()
}
