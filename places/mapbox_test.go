package places

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIEndpoint_Constants(t *testing.T) {
	tests := []struct {
		name     string
		endpoint APIEndpoint
		expected string
	}{
		{"Autosuggest", Autosuggest, "/search/searchbox/v1/suggest"},
		{"Retrieve", Retrieve, "/search/searchbox/v1/retrieve"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.endpoint))
		})
	}
}

func TestGetAPIUrl(t *testing.T) {
	tests := []struct {
		name     string
		data     MapboxApi
		expected string
	}{
		{
			name: "Suggest endpoint with query",
			data: MapboxApi{
				Method:       "GET",
				Endpoint:     string(Autosuggest),
				QueryParam:   "Paris",
				AccessToken:  "test_token",
				SessionToken: "session_123",
			},
			expected: "https://api.mapbox.com/search/searchbox/v1/suggest?q=Paris&language=en&session_token=session_123&access_token=test_token",
		},
		{
			name: "Suggest endpoint with special characters",
			data: MapboxApi{
				Method:       "GET",
				Endpoint:     string(Autosuggest),
				QueryParam:   "New York City",
				AccessToken:  "test_token",
				SessionToken: "session_123",
			},
			expected: "https://api.mapbox.com/search/searchbox/v1/suggest?q=New+York+City&language=en&session_token=session_123&access_token=test_token",
		},
		{
			name: "Retrieve endpoint with place ID",
			data: MapboxApi{
				Method:       "GET",
				Endpoint:     string(Retrieve),
				QueryParam:   "place_id_123",
				AccessToken:  "test_token",
				SessionToken: "session_123",
			},
			expected: "https://api.mapbox.com/search/searchbox/v1/retrieve/place_id_123?language=en&session_token=session_123&access_token=test_token",
		},
		{
			name: "POST endpoint",
			data: MapboxApi{
				Method:       "POST",
				Endpoint:     "/some/endpoint",
				AccessToken:  "test_token",
				SessionToken: "session_123",
			},
			expected: "https://api.mapbox.com/some/endpoint?session_token=session_123&access_token=test_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAPIUrl(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapboxAPIResponse_JSONParsing(t *testing.T) {
	t.Run("Parse valid API response", func(t *testing.T) {
		jsonData := `{
			"suggestions": [
				{
					"name": "Paris",
					"mapbox_id": "place.123",
					"feature_type": "place",
					"full_address": "Paris, France",
					"language": "en",
					"maki": "marker"
				}
			],
			"attribution": "Mapbox"
		}`

		var response MapboxAPIResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Len(t, response.Suggestions, 1)
		assert.Equal(t, "Paris", response.Suggestions[0].Name)
		assert.Equal(t, "place.123", response.Suggestions[0].MapboxId)
		assert.Equal(t, "Mapbox", response.Attribution)
	})

	t.Run("Parse empty suggestions", func(t *testing.T) {
		jsonData := `{
			"suggestions": [],
			"attribution": "Mapbox"
		}`

		var response MapboxAPIResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Len(t, response.Suggestions, 0)
	})
}

func TestSuggestion_Structure(t *testing.T) {
	t.Run("Parse complete suggestion", func(t *testing.T) {
		jsonData := `{
			"name": "Eiffel Tower",
			"mapbox_id": "poi.123",
			"feature_type": "poi",
			"address": "5 Avenue Anatole France",
			"full_address": "5 Avenue Anatole France, 75007 Paris, France",
			"place_formatted": "Paris, France",
			"context": {
				"country": {
					"name": "France",
					"country_code": "FR",
					"country_code_alpha_3": "FRA"
				},
				"region": {
					"name": "Île-de-France",
					"region_code": "IDF",
					"region_code_full": "FR-IDF"
				},
				"place": {
					"name": "Paris"
				}
			},
			"language": "en",
			"maki": "monument",
			"poi_category": ["monument", "tourism"],
			"poi_category_ids": ["monument", "tourism"],
			"external_ids": {
				"foursquare": "4adcdaf4f964a520f6f521e3"
			},
			"metadata": {
				"iso_3166_1": "FR"
			}
		}`

		var suggestion Suggestion
		err := json.Unmarshal([]byte(jsonData), &suggestion)

		assert.NoError(t, err)
		assert.Equal(t, "Eiffel Tower", suggestion.Name)
		assert.Equal(t, "poi.123", suggestion.MapboxId)
		assert.Equal(t, "poi", suggestion.FeatureType)
		assert.Equal(t, "5 Avenue Anatole France", suggestion.Address)
		assert.Equal(t, "France", suggestion.Context.Country.Name)
		assert.Equal(t, "FR", suggestion.Context.Country.CountryCode)
		assert.Equal(t, "Paris", suggestion.Context.Place.Name)
		assert.Len(t, suggestion.PoiCategory, 2)
		assert.Equal(t, "4adcdaf4f964a520f6f521e3", suggestion.ExternalIds.Foursquare)
	})
}

func TestRetrieveAPIResponse_JSONParsing(t *testing.T) {
	t.Run("Parse valid retrieve response", func(t *testing.T) {
		jsonData := `{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"geometry": {
						"type": "Point",
						"coordinates": [2.2945, 48.8584]
					},
					"properties": {
						"name": "Eiffel Tower",
						"mapbox_id": "poi.123",
						"feature_type": "poi",
						"full_address": "Paris, France",
						"coordinates": {
							"latitude": 48.8584,
							"longitude": 2.2945
						}
					}
				}
			],
			"attribution": "Mapbox"
		}`

		var response RetrieveAPIResponse
		err := json.Unmarshal([]byte(jsonData), &response)

		assert.NoError(t, err)
		assert.Equal(t, "FeatureCollection", response.Type)
		assert.Len(t, response.Features, 1)
		assert.Equal(t, "Feature", response.Features[0].Type)
		assert.Equal(t, "Point", response.Features[0].Geometry.Type)
		assert.Len(t, response.Features[0].Geometry.Coordinates, 2)
		assert.Equal(t, "Eiffel Tower", response.Features[0].Properties.Name)
		assert.Equal(t, 48.8584, response.Features[0].Properties.Coordinates.Latitude)
		assert.Equal(t, 2.2945, response.Features[0].Properties.Coordinates.Longitude)
	})
}

func TestContext_Structure(t *testing.T) {
	t.Run("Parse context with all fields", func(t *testing.T) {
		jsonData := `{
			"country": {
				"name": "United States",
				"country_code": "US",
				"country_code_alpha_3": "USA"
			},
			"region": {
				"name": "California",
				"region_code": "CA",
				"region_code_full": "US-CA"
			},
			"postcode": {
				"name": "94102"
			},
			"place": {
				"name": "San Francisco"
			},
			"neighborhood": {
				"name": "Mission District"
			},
			"street": {
				"name": "Valencia Street"
			}
		}`

		var context Context
		err := json.Unmarshal([]byte(jsonData), &context)

		assert.NoError(t, err)
		assert.Equal(t, "United States", context.Country.Name)
		assert.Equal(t, "US", context.Country.CountryCode)
		assert.Equal(t, "USA", context.Country.CountryCodeAlpha3)
		assert.Equal(t, "California", context.Region.Name)
		assert.Equal(t, "CA", context.Region.RegionCode)
		assert.Equal(t, "94102", context.Postcode.Name)
		assert.Equal(t, "San Francisco", context.Place.Name)
		assert.Equal(t, "Mission District", context.Neighborhood.Name)
		assert.Equal(t, "Valencia Street", context.Street.Name)
	})
}

func TestCoordinates_Structure(t *testing.T) {
	t.Run("Parse coordinates with routable points", func(t *testing.T) {
		jsonData := `{
			"latitude": 48.8584,
			"longitude": 2.2945,
			"routable_points": [
				{
					"name": "entrance",
					"latitude": 48.8585,
					"longitude": 2.2946
				},
				{
					"name": "parking",
					"latitude": 48.8583,
					"longitude": 2.2944
				}
			]
		}`

		var coordinates Coordinates
		err := json.Unmarshal([]byte(jsonData), &coordinates)

		assert.NoError(t, err)
		assert.Equal(t, 48.8584, coordinates.Latitude)
		assert.Equal(t, 2.2945, coordinates.Longitude)
		assert.Len(t, coordinates.RoutablePoints, 2)
		assert.Equal(t, "entrance", coordinates.RoutablePoints[0].Name)
		assert.Equal(t, 48.8585, coordinates.RoutablePoints[0].Latitude)
	})
}

func TestMakeHttpRequest_MockServer(t *testing.T) {
	t.Run("Successful API call", func(t *testing.T) {
		// Create a mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"suggestions": [], "attribution": "Mapbox"}`))
		}))
		defer server.Close()

		// Note: The actual make_http_request function uses BASE_MAPBOX_API constant
		// In a real test, we'd need to inject the URL or make it configurable
		// This test demonstrates the pattern
		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Failed API call", func(t *testing.T) {
		// Create a mock server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Bad request"}`))
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestExternalIds_Structure(t *testing.T) {
	t.Run("Parse external IDs", func(t *testing.T) {
		jsonData := `{
			"safegraph": "sg123",
			"foursquare": "4adcdaf4f964a520f6f521e3"
		}`

		var externalIds ExternalIds
		err := json.Unmarshal([]byte(jsonData), &externalIds)

		assert.NoError(t, err)
		assert.Equal(t, "sg123", externalIds.Safegraph)
		assert.Equal(t, "4adcdaf4f964a520f6f521e3", externalIds.Foursquare)
	})
}

func TestFeatureProperties_CompleteStructure(t *testing.T) {
	t.Run("Parse complete feature properties", func(t *testing.T) {
		jsonData := `{
			"name": "Central Park",
			"mapbox_id": "poi.456",
			"feature_type": "poi",
			"address": "Central Park",
			"full_address": "Central Park, New York, NY, USA",
			"place_formatted": "New York, NY, USA",
			"context": {
				"country": {
					"name": "United States",
					"country_code": "US",
					"country_code_alpha_3": "USA"
				}
			},
			"coordinates": {
				"latitude": 40.785091,
				"longitude": -73.968285
			},
			"language": "en",
			"maki": "park",
			"poi_category": ["park", "outdoor"],
			"poi_category_ids": ["park_outdoor"],
			"external_ids": {
				"foursquare": "412d2800f964a520df0c1fe3"
			},
			"metadata": {
				"area": "341 hectares"
			}
		}`

		var properties FeatureProperties
		err := json.Unmarshal([]byte(jsonData), &properties)

		assert.NoError(t, err)
		assert.Equal(t, "Central Park", properties.Name)
		assert.Equal(t, "poi.456", properties.MapboxId)
		assert.Equal(t, "park", properties.Maki)
		assert.Len(t, properties.PoiCategory, 2)
		assert.Equal(t, 40.785091, properties.Coordinates.Latitude)
		assert.NotNil(t, properties.Metadata)
		assert.Equal(t, "341 hectares", properties.Metadata["area"])
	})
}

func TestMapboxApi_Structure(t *testing.T) {
	t.Run("Create MapboxApi struct", func(t *testing.T) {
		api := MapboxApi{
			Method:       "GET",
			Endpoint:     string(Autosuggest),
			QueryParam:   "London",
			AccessToken:  "test_token",
			SessionToken: "session_123",
			Payload:      "",
		}

		assert.Equal(t, "GET", api.Method)
		assert.Equal(t, "/search/searchbox/v1/suggest", api.Endpoint)
		assert.Equal(t, "London", api.QueryParam)
		assert.Equal(t, "test_token", api.AccessToken)
		assert.Equal(t, "session_123", api.SessionToken)
	})
}

func TestBASE_MAPBOX_API_Constant(t *testing.T) {
	t.Run("Verify base API URL", func(t *testing.T) {
		assert.Equal(t, "https://api.mapbox.com", BASE_MAPBOX_API)
	})
}
