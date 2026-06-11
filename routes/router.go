package routes

import "github.com/gin-gonic/gin"

// RouterGroupRoutes registers route handlers under the supplied group.
func RouterGroupRoutes(router *gin.RouterGroup) {
	router.GET("", GetRouteHandler)
}

// DefaultProvider returns the configured default provider for use by other packages
// (e.g. trips/gemini_provider.go for tool calling).
func DefaultProvider() (Provider, error) {
	return defaultFactory.GetDefaultProvider()
}
