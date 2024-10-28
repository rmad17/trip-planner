package places


func search(){}
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
