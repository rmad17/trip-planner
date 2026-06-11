package hotels

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

// SearchHotelsHandler godoc
// @Summary Search hotels
// @Description Returns hotel offers from the configured provider.
// @Tags hotels
// @Produce json
// @Param city query string true "City name or IATA city code"
// @Param check_in query string true "Check-in date YYYY-MM-DD"
// @Param check_out query string true "Check-out date YYYY-MM-DD"
// @Param guests query int false "Number of guests (default 2)"
// @Param currency query string false "Currency code (e.g. EUR)"
// @Param max_price query number false "Max price per night"
// @Success 200 {array} HotelOffer
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /search/hotels [get]
func SearchHotelsHandler(c *gin.Context) {
	city := strings.TrimSpace(c.Query("city"))
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city is required"})
		return
	}
	checkIn := c.Query("check_in")
	checkOut := c.Query("check_out")
	if checkIn == "" || checkOut == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "check_in and check_out are required"})
		return
	}

	guests := 2
	if g := c.Query("guests"); g != "" {
		if v, err := strconv.Atoi(g); err == nil && v > 0 {
			guests = v
		}
	}

	currency := strings.ToUpper(c.Query("currency"))

	maxPrice := 0.0
	if mp := c.Query("max_price"); mp != "" {
		if v, err := strconv.ParseFloat(mp, 64); err == nil {
			maxPrice = v
		}
	}

	q := HotelSearchQuery{
		City:             city,
		CheckIn:          checkIn,
		CheckOut:         checkOut,
		Guests:           guests,
		Currency:         currency,
		MaxPricePerNight: maxPrice,
	}

	cacheKey := "hotels:" + city + ":" + checkIn + ":" + checkOut + ":" + strconv.Itoa(guests) + ":" + currency + ":" + strconv.FormatFloat(maxPrice, 'f', 2, 64)
	if v, ok := cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, v)
		return
	}

	provider, err := defaultFactory.GetDefaultProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	offers, err := provider.SearchHotels(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if offers == nil {
		offers = []HotelOffer{}
	}

	cache.Set(cacheKey, offers)
	c.JSON(http.StatusOK, offers)
}
