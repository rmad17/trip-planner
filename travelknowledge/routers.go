package travelknowledge

import "github.com/gin-gonic/gin"

// RouterGroupTravelKnowledge sets up routes for travel knowledge API
func RouterGroupTravelKnowledge(router *gin.RouterGroup) {
	// Get available categories/types
	router.GET("/categories", GetAvailableCategoriesHandler)

	// Get all cities
	router.GET("/cities", ListCitiesHandler)

	// Get documents for a city (with optional type filter)
	router.GET("/city", GetCityDocumentsHandler)

	// Get aggregated city info
	router.GET("/city/:city/info", GetCityInfoHandler)

	// Get documents by category/type across all cities
	router.GET("/category/:type", GetDocumentsByTypeHandler)

	// Get knowledge base statistics
	router.GET("/stats", GetDocumentStatsHandler)
}
