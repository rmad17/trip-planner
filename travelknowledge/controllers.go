package travelknowledge

import (
	"net/http"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

var service *Service

func init() {
	db := core.GetDB()
	if db != nil {
		service = NewService(db)
	}
}

// GetCityDocumentsHandler godoc
// @Summary Get travel documents for a city
// @Description Retrieves all travel documents for a specific city with tags as categories
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Param city query string true "City name"
// @Param type query string false "Document type filter (historical, logistics, food, activities, nearby_places)"
// @Success 200 {object} map[string]interface{} "City documents grouped by type"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "City not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /travel-knowledge/city [get]
func GetCityDocumentsHandler(c *gin.Context) {
	cityName := c.Query("city")
	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city parameter is required"})
		return
	}

	docType := c.Query("type")

	var documents []TravelDocument
	var err error

	if docType != "" {
		// Filter by type
		documents, err = service.GetCityDocumentsByType(c.Request.Context(), cityName, DocumentType(docType))
	} else {
		// Get all documents for city
		documents, err = service.GetCityDocuments(c.Request.Context(), cityName)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(documents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no documents found for this city"})
		return
	}

	// Group documents by type (tags/categories)
	documentsByType := make(map[DocumentType][]TravelDocument)
	for _, doc := range documents {
		documentsByType[doc.DocumentType] = append(documentsByType[doc.DocumentType], doc)
	}

	// Get first document's country info
	var country string
	if len(documents) > 0 {
		country = documents[0].CityCountry
	}

	c.JSON(http.StatusOK, gin.H{
		"city_name":    cityName,
		"city_country": country,
		"total_count":  len(documents),
		"categories":   documentsByType,
	})
}

// GetDocumentsByTypeHandler godoc
// @Summary Get travel documents by type/category
// @Description Retrieves all travel documents of a specific type across all cities
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Param type path string true "Document type (historical, logistics, food, activities, nearby_places)"
// @Success 200 {object} map[string]interface{} "Documents of the specified type"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /travel-knowledge/category/{type} [get]
func GetDocumentsByTypeHandler(c *gin.Context) {
	docType := c.Param("type")
	if docType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type parameter is required"})
		return
	}

	documents, err := service.GetDocumentsByType(c.Request.Context(), DocumentType(docType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": docType,
		"count":    len(documents),
		"documents": documents,
	})
}

// ListCitiesHandler godoc
// @Summary List all cities in the knowledge base
// @Description Returns all cities with document counts
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of cities with document counts"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /travel-knowledge/cities [get]
func ListCitiesHandler(c *gin.Context) {
	cities, err := service.ListCities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(cities),
		"cities": cities,
	})
}

// GetCityInfoHandler godoc
// @Summary Get aggregated city information
// @Description Returns comprehensive travel information for a city grouped by categories
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Param city path string true "City name"
// @Success 200 {object} CityTravelInfo "Aggregated city information"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "City not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /travel-knowledge/city/{city}/info [get]
func GetCityInfoHandler(c *gin.Context) {
	cityName := c.Param("city")
	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city parameter is required"})
		return
	}

	info, err := service.GetCityInfo(c.Request.Context(), cityName)
	if err != nil {
		if err.Error() == "no documents found for city: "+cityName {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetDocumentStatsHandler godoc
// @Summary Get knowledge base statistics
// @Description Returns statistics about the travel knowledge base
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Knowledge base statistics"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /travel-knowledge/stats [get]
func GetDocumentStatsHandler(c *gin.Context) {
	stats, err := service.GetDocumentStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAvailableCategoriesHandler godoc
// @Summary Get available document categories/types
// @Description Returns all available document categories that can be used as filters
// @Tags travel-knowledge
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Available categories"
// @Router /travel-knowledge/categories [get]
func GetAvailableCategoriesHandler(c *gin.Context) {
	categories := []map[string]string{
		{"type": string(DocumentTypeHistorical), "description": "Historical & cultural information"},
		{"type": string(DocumentTypeLogistics), "description": "Travel planning & logistics"},
		{"type": string(DocumentTypeFood), "description": "Cuisine & dining recommendations"},
		{"type": string(DocumentTypeActivities), "description": "Attractions & things to do"},
		{"type": string(DocumentTypeNearbyPlaces), "description": "Nearby destinations & day trips"},
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}
