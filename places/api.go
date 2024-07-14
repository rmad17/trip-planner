package places

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"googlemaps.github.io/maps"
)

func SearchAutocomplete(c *gin.Context) {
	// authHeader := c.GetHeader("Authorization")

	// if authHeader == "" {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
	//     c.AbortWithStatus(http.StatusUnauthorized)
	//     return
	// }

	apiKey := os.Getenv("GOOGLE_API_KEY")
	client, _ := maps.NewClient(maps.WithAPIKey(apiKey))
	sessiontoken := maps.NewPlaceAutocompleteSessionToken()
	data := c.Request.URL.Query()
	pretty.Println("Json: ", data)

	request := &maps.PlaceAutocompleteRequest{
		Input:    data.Get("input"),
		Language: c.GetString("language"),
		// Offset:       offset,
		// Radius:       c.GetUint("radius"),
		// StrictBounds: c.GetBool("strictbounds"),
		SessionToken: sessiontoken,
	}
	pretty.Println("Request: ", request)
	resp, err := client.PlaceAutocomplete(context.Background(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		pretty.Println("Error: ", err.Error())
	} else {
		c.JSON(http.StatusOK, gin.H{"predictions": resp.Predictions})
		// pretty.Println("Response: ", resp)
	}
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
