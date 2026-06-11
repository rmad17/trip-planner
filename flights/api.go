package flights

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

// SearchFlightsHandler godoc
// @Summary Search flights
// @Description Returns flight offers from the configured provider.
// @Tags flights
// @Produce json
// @Param from query string true "Origin IATA code (or city name)"
// @Param to query string true "Destination IATA code (or city name)"
// @Param depart query string true "Departure date YYYY-MM-DD"
// @Param return query string false "Return date YYYY-MM-DD (omit for one-way)"
// @Param pax query int false "Number of passengers (default 1)"
// @Param cabin query string false "Cabin: economy | premium_economy | business | first"
// @Param currency query string false "Currency code (e.g. EUR)"
// @Success 200 {array} FlightOffer
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /search/flights [get]
func SearchFlightsHandler(c *gin.Context) {
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to are required"})
		return
	}
	depart := c.Query("depart")
	if depart == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "depart is required"})
		return
	}

	fromCode, ok1 := LookupIATA(from)
	toCode, ok2 := LookupIATA(to)
	if !ok1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve IATA code for from=" + from})
		return
	}
	if !ok2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not resolve IATA code for to=" + to})
		return
	}

	pax := 1
	if v := c.Query("pax"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pax = n
		}
	}
	cabin := c.DefaultQuery("cabin", "economy")
	currency := strings.ToUpper(c.Query("currency"))
	ret := c.Query("return")

	q := FlightSearchQuery{
		From:     fromCode,
		To:       toCode,
		Depart:   depart,
		Return:   ret,
		Pax:      pax,
		Cabin:    cabin,
		Currency: currency,
	}

	cacheKey := "flights:" + fromCode + ":" + toCode + ":" + depart + ":" + ret + ":" + strconv.Itoa(pax) + ":" + cabin + ":" + currency
	if v, ok := cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, v)
		return
	}

	provider, err := defaultFactory.GetDefaultProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	offers, err := provider.SearchFlights(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if offers == nil {
		offers = []FlightOffer{}
	}

	cache.Set(cacheKey, offers)
	c.JSON(http.StatusOK, offers)
}
