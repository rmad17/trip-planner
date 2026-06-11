package flights

import "github.com/gin-gonic/gin"

// RouterGroupFlights registers flight search handlers on the supplied group.
func RouterGroupFlights(router *gin.RouterGroup) {
	router.GET("", SearchFlightsHandler)
}

// DefaultProvider returns the configured default provider for use by other packages.
func DefaultProvider() (Provider, error) {
	return defaultFactory.GetDefaultProvider()
}
