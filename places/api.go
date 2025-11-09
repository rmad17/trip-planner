package places

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/kr/pretty"
	"googlemaps.github.io/maps"
)

// SearchAutocomplete godoc
// @Summary Search for place suggestions using autocomplete
// @Description Get place suggestions from Mapbox API based on search text
// @Tags places
// @Produce json
// @Param text query string true "Search text for autocomplete"
// @Success 200 {object} map[string]interface{} "Autocomplete suggestions"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /places/autocomplete/search [get]
func SearchAutocomplete(c *gin.Context) {
	MapboxAPIKey := os.Getenv(core.SEARCH_API_KEY)
	if MapboxAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MAPBOX_TOKEN environment variable is not set"})
		return
	}

	data := c.Request.URL.Query()
	SearchText := data.Get("query")

	if SearchText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text parameter is required"})
		return
	}

	pretty.Println("SearchText: ", SearchText)
	pretty.Println("MapboxAPIKey set: ", MapboxAPIKey != "")

	SessionToken := uuid.Must(uuid.NewV4()).String()
	response_data := make_http_request(MapboxApi{"GET", string(Autosuggest), SearchText, MapboxAPIKey, SessionToken, ""})

	if response_data == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from Mapbox API"})
		return
	}

	var res MapboxAPIResponse
	err := json.Unmarshal(response_data, &res)
	if err != nil {
		pretty.Println("JSON unmarshal error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	pretty.Println("Response: ", res)
	c.JSON(http.StatusOK, gin.H{"data": res})
}

// PlaceRetrieve godoc
// @Summary Retrieve detailed place information using Mapbox ID
// @Description Get detailed place information from Mapbox API using a place ID obtained from autocomplete
// @Tags places
// @Produce json
// @Param id path string true "Mapbox Place ID"
// @Param language query string false "ISO language code (default: en)"
// @Param eta_type query string false "Enable ETA calculation (only 'navigation' allowed)"
// @Param navigation_profile query string false "Navigation profile for ETA (driving, walking, cycling)"
// @Param origin query string false "Origin coordinates for ETA calculation (longitude,latitude)"
// @Success 200 {object} map[string]interface{} "Place details"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /places/retrieve/{id} [get]
func PlaceRetrieve(c *gin.Context) {
	MapboxAPIKey := os.Getenv(core.SEARCH_API_KEY)
	if MapboxAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MAPBOX_TOKEN environment variable is not set"})
		return
	}

	// Get the ID from URL path
	placeID := c.Param("id")
	if placeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "place ID parameter is required"})
		return
	}

	// Get optional query parameters
	data := c.Request.URL.Query()
	language := data.Get("language")
	if language == "" {
		language = "en"
	}

	pretty.Println("PlaceID: ", placeID)
	pretty.Println("Language: ", language)

	SessionToken := uuid.Must(uuid.NewV4()).String()

	// Build the request - for retrieve endpoint, the ID goes in the URL path
	response_data := make_http_request(MapboxApi{"GET", string(Retrieve), placeID, MapboxAPIKey, SessionToken, ""})

	if response_data == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from Mapbox API"})
		return
	}

	var res RetrieveAPIResponse
	err := json.Unmarshal(response_data, &res)
	if err != nil {
		pretty.Println("JSON unmarshal error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	pretty.Println("Response: ", res)
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func PlaceDetails(c *gin.Context) {

	apiKey := os.Getenv("GOOGLE_API_KEY")
	client, _ := maps.NewClient(maps.WithAPIKey(apiKey))
	data := c.Request.URL.Query()
	pretty.Println("Json: ", data)

	request := &maps.PlaceDetailsRequest{
		PlaceID: data.Get("place_id"),
	}
	pretty.Println("Request: ", request)
	resp, err := client.PlaceDetails(context.Background(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		pretty.Println("Error: ", err.Error())
	} else {
		c.JSON(http.StatusOK, gin.H{"predictions": resp})
		pretty.Println("Response: ", resp)
	}
}
