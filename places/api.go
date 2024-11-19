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

func SearchAutocomplete(c *gin.Context) {
	MapboxAPIKey := os.Getenv(core.SEARCH_API_KEY)
	data := c.Request.URL.Query()
	SearchText := data.Get("search_text")
	pretty.Println("Text: ", SearchText)

	SessionToken := uuid.Must(uuid.NewV4()).String()
	response := make_http_request(MapboxApi{"GET", string(Autosuggest), SearchText, MapboxAPIKey, SessionToken, ""})
	pretty.Println("Response: ", string(response))
	c.JSON(http.StatusOK, gin.H{"data": json.Marshal(string(response))})
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
