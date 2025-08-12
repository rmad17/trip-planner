package places

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const BASE_MAPBOX_API = "https://api.mapbox.com"

type APIEndpoint string

const (
	Autosuggest APIEndpoint = "/search/searchbox/v1/suggest"
	Retrieve    APIEndpoint = "/search/searchbox/v1/retrieve"
)

type MapboxApi struct {
	Method       string
	Endpoint     string
	QueryParam   string
	AccessToken  string
	SessionToken string
	Payload      string
}

type MapboxAPIResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Attribution string       `json:"attribution"`
}

type Country struct {
	Name              string `json:"name"`
	CountryCode       string `json:"country_code"`
	CountryCodeAlpha3 string `json:"country_code_alpha_3"`
}

type Region struct {
	Name           string `json:"name"`
	RegionCode     string `json:"region_code"`
	RegionCodeFull string `json:"region_code_full"`
}

type Neighborhood struct {
	Name string `json:"name"`
}

type Street struct {
	Name string `json:"name"`
}

type Postcode struct {
	Name string `json:"name"`
}

type Place struct {
	Name string `json:"name"`
}
type Context struct {
	Country      Country      `json:"country"`
	Region       Region       `json:"region"`
	Postcode     Postcode     `json:"postcode"`
	Place        Place        `json:"place"`
	Neighborhood Neighborhood `json:"neighborhood"`
	Street       Street       `json:"street"`
}

type ExternalIds struct {
	Safegraph  string `json:"safegraph"`
	Foursquare string `json:"foursquare"`
}

type Coordinates struct {
	Latitude       float64         `json:"latitude"`
	Longitude      float64         `json:"longitude"`
	RoutablePoints []RoutablePoint `json:"routable_points"`
}

type RoutablePoint struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
	Type        string    `json:"type"`
}

type FeatureProperties struct {
	Name           string                 `json:"name"`
	MapboxId       string                 `json:"mapbox_id"`
	FeatureType    string                 `json:"feature_type"`
	Address        string                 `json:"address"`
	FullAddress    string                 `json:"full_address"`
	PlaceFormatted string                 `json:"place_formatted"`
	Context        Context                `json:"context"`
	Coordinates    Coordinates            `json:"coordinates"`
	Language       string                 `json:"language"`
	Maki           string                 `json:"maki"`
	PoiCategory    []string               `json:"poi_category"`
	PoiCategoryIds []string               `json:"poi_category_ids"`
	ExternalIds    ExternalIds            `json:"external_ids"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type Feature struct {
	Type       string            `json:"type"`
	Geometry   Geometry          `json:"geometry"`
	Properties FeatureProperties `json:"properties"`
}

type RetrieveAPIResponse struct {
	Type        string    `json:"type"`
	Features    []Feature `json:"features"`
	Attribution string    `json:"attribution"`
}
type Suggestion struct {
	Name           string                 `json:"name"`
	MapboxId       string                 `json:"mapbox_id"`
	FeatureType    string                 `json:"feature_type"`
	Address        string                 `json:"address"`
	FullAddress    string                 `json:"full_address"`
	PlaceFormatted string                 `json:"place_formatted"`
	Context        Context                `json:"context"`
	Language       string                 `json:"language"`
	Maki           string                 `json:"maki"`
	PoiCategory    []string               `json:"poi_category"`
	PoiCategoryIds []string               `json:"poi_category_ids"`
	ExternalIds    ExternalIds            `json:"external_ids"`
	Metadata       map[string]interface{} `json:"metadata"`
}

func GetAPIUrl(data MapboxApi) string {
	if data.Method == "GET" {
		// Handle retrieve endpoint differently (has ID in path)
		if data.Endpoint == string(Retrieve) {
			return fmt.Sprintf("%s%s/%s?language=en&session_token=%s&access_token=%s", BASE_MAPBOX_API, data.Endpoint, data.QueryParam, data.SessionToken, data.AccessToken)
		}
		// Handle suggest endpoint (has query parameter)
		encodedQuery := url.QueryEscape(data.QueryParam)
		return fmt.Sprintf("%s%s?q=%s&language=en&session_token=%s&access_token=%s", BASE_MAPBOX_API, data.Endpoint, encodedQuery, data.SessionToken, data.AccessToken)
	}
	return fmt.Sprintf("%s%s?session_token=%s&access_token=%s", BASE_MAPBOX_API, data.Endpoint, data.SessionToken, data.AccessToken)

}

func make_http_request(data MapboxApi) []byte {
	url := GetAPIUrl(data)
	log.Printf("Making request to URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	log.Printf("HTTP response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("HTTP request failed with status: %d, body: %s", resp.StatusCode, string(body))
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil
	}

	log.Printf("Response body: %s", string(body))
	return body
}

// func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
// 	err = c.parseRequest(r, opts...)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	req, err := http.NewRequest(r.method, r.fullURL, r.body)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	req = req.WithContext(ctx)
// 	req.Header = r.header
// 	c.debug("request: %#v\n", req)
// 	f := c.do
// 	if f == nil {
// 		f = c.HTTPClient.Do
// 	}
// 	res, err := f(req)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	data, err = io.ReadAll(res.Body)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	defer func() {
// 		cerr := res.Body.Close()
// 		// Only overwrite the returned error if the original error was nil and an
// 		// error occurred while closing the body.
// 		if err == nil && cerr != nil {
// 			err = cerr
// 		}
// 	}()
// 	c.debug("response: %#v\n", res)
// 	c.debug("response body: %s\n", string(data))
// 	c.debug("response status code: %d\n", res.StatusCode)
//
// 	if res.StatusCode >= http.StatusBadRequest {
// 		apiErr := new(common.APIError)
// 		e := json.Unmarshal(data, apiErr)
// 		if e != nil {
// 			c.debug("failed to unmarshal json: %s\n", e)
// 		}
// 		if !apiErr.IsValid() {
// 			apiErr.Response = data
// 		}
// 		return nil, apiErr
// 	}
// 	return data, nil
// }
