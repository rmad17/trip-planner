package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

var (
	defaultFactory = NewProviderFactory()
	cache          = core.NewTTLCache(60*time.Second, 256)
)

// GetRouteHandler godoc
// @Summary Compute a multi-stop route
// @Description Returns distance, duration, and a GeoJSON polyline for a route.
// @Tags routes
// @Produce json
// @Param stops query string true "Semicolon-separated lng,lat pairs (e.g. 2.35,48.85;4.83,45.76)"
// @Param mode query string false "Travel mode: drive (default) | walk | transit | cycle"
// @Success 200 {object} Route
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /routes [get]
func GetRouteHandler(c *gin.Context) {
	stopsParam := strings.TrimSpace(c.Query("stops"))
	if stopsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stops is required"})
		return
	}
	mode := c.DefaultQuery("mode", "drive")

	stops, err := parseStops(stopsParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cacheKey := "routes:" + mode + ":" + stopsParam
	if v, ok := cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, v)
		return
	}

	provider, err := defaultFactory.GetDefaultProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	route, err := provider.GetRoute(c.Request.Context(), RouteQuery{Stops: stops, Mode: mode})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Set(cacheKey, route)
	c.JSON(http.StatusOK, route)
}

// parseStops parses "lng,lat;lng,lat;..." into []LngLat.
func parseStops(raw string) ([]LngLat, error) {
	pairs := strings.Split(raw, ";")
	if len(pairs) < 2 {
		return nil, errInvalidStops
	}
	out := make([]LngLat, 0, len(pairs))
	for _, p := range pairs {
		parts := strings.Split(strings.TrimSpace(p), ",")
		if len(parts) != 2 {
			return nil, errInvalidStops
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, errInvalidStops
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, errInvalidStops
		}
		out = append(out, LngLat{Lng: lng, Lat: lat})
	}
	return out, nil
}

var errInvalidStops = invalidStopsError{}

type invalidStopsError struct{}

func (invalidStopsError) Error() string {
	return "stops must be semicolon-separated lng,lat pairs (need at least 2)"
}
