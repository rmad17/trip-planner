package places

import (
	"context"
	"net/http"
	"os"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"googlemaps.github.io/maps"
)

func SearchAutocomplete(c *gin.Context) {
	api_to_be_used := os.Getenv(core.SEARCH_API_KEY)

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
