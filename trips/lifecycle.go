package trips

import (
	"net/http"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validTransitions defines the allowed state machine for trip status.
// terminal states (completed, cancelled) have empty slices.
var validTransitions = map[TripStatus][]TripStatus{
	TripStatusPlanning:  {TripStatusConfirmed, TripStatusCancelled},
	TripStatusConfirmed: {TripStatusOngoing, TripStatusCancelled},
	TripStatusOngoing:   {TripStatusCompleted, TripStatusCancelled},
	TripStatusCompleted: {},
	TripStatusCancelled: {},
}

// isValidTransition returns true if moving from current → next is permitted.
func isValidTransition(current, next TripStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// emergencyNumbers maps ISO country codes to the primary emergency number.
// Falls back to "112" (international standard) for unknown countries.
var emergencyNumbers = map[string]string{
	"IN": "112", // India
	"US": "911",
	"GB": "999",
	"AU": "000",
	"EU": "112",
	"DE": "112",
	"FR": "112",
	"IT": "112",
	"JP": "110",
	"CN": "110",
	"CA": "911",
	"NZ": "111",
	"SG": "999",
	"AE": "999",
	"TH": "191",
	"NP": "100",
	"LK": "119",
	"BD": "999",
}

// getEmergencyNumber returns the emergency number for the given ISO country code.
func getEmergencyNumber(countryCode string) string {
	if n, ok := emergencyNumbers[countryCode]; ok {
		return n
	}
	return "112"
}

// StatusUpdateRequest is the request body for updating trip status.
type StatusUpdateRequest struct {
	Status TripStatus `json:"status" binding:"required" example:"confirmed"`
}

// ShiftDatesRequest is the request body for shifting all trip dates.
type ShiftDatesRequest struct {
	DeltaDays int `json:"delta_days" binding:"required" example:"3"`
}

// TodayView is the response from GET /trip/:id/today.
type TodayView struct {
	TripID            uuid.UUID          `json:"trip_id"`
	TripName          *string            `json:"trip_name"`
	Date              string             `json:"date"`
	DayNumber         int                `json:"day_number"`
	TripDay           *TripDay           `json:"trip_day"`
	CurrentHop        *TripHop           `json:"current_hop"`
	CurrentStay       *Stay              `json:"current_stay"`
	CheckInToday      bool               `json:"check_in_today"`
	CheckOutToday     bool               `json:"check_out_today"`
	UpcomingTransport *TransportSegment  `json:"upcoming_transport"`
	EmergencyContacts []TripContact      `json:"emergency_contacts"`
	EmergencyNumber   string             `json:"emergency_number"`
}

// UpdateTripStatus godoc
// @Summary Update the status of a trip plan
// @Description Transitions the trip through its lifecycle. Validates allowed transitions.
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param body body StatusUpdateRequest true "New status"
// @Success 200 {object} TripPlan "Updated trip plan"
// @Failure 400 {object} map[string]string "Invalid transition"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /trip/{id}/status [patch]
func UpdateTripStatus(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var trip TripPlan
	if err := core.DB.Where("id = ?", tripID).First(&trip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}
	if trip.UserID != user.BaseModel.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this trip"})
		return
	}

	current := TripStatusPlanning
	if trip.Status != nil {
		current = *trip.Status
	}

	if !isValidTransition(current, req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid status transition",
			"from":    current,
			"to":      req.Status,
			"allowed": validTransitions[current],
		})
		return
	}

	if err := core.DB.Model(&trip).Updates(map[string]interface{}{
		"status":          req.Status,
		"user_set_status": true,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// ShiftTripDates godoc
// @Summary Shift all dates in a trip by a number of days
// @Description Atomically shifts trip start/end, all hop dates, all day dates, and all activity times.
// @Tags lifecycle
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param body body ShiftDatesRequest true "Delta in days (positive or negative)"
// @Success 200 {object} map[string]int "days shifted"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 403 {object} map[string]string "Forbidden"
// @Security BearerAuth
// @Router /trip/{id}/shift-dates [post]
func ShiftTripDates(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	var req ShiftDatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DeltaDays == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "delta_days cannot be zero"})
		return
	}

	var trip TripPlan
	if err := core.DB.Where("id = ?", tripID).First(&trip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}
	if trip.UserID != user.BaseModel.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this trip"})
		return
	}

	delta := time.Duration(req.DeltaDays) * 24 * time.Hour

	tx := core.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	// Shift trip plan dates
	if trip.StartDate != nil {
		shifted := trip.StartDate.Add(delta)
		tx.Model(&trip).Update("start_date", shifted)
	}
	if trip.EndDate != nil {
		shifted := trip.EndDate.Add(delta)
		tx.Model(&trip).Update("end_date", shifted)
	}

	// Shift hop dates
	tx.Exec(`UPDATE trip_hops SET
		start_date = start_date + make_interval(days => ?),
		end_date   = end_date   + make_interval(days => ?)
	WHERE trip_plan = ? AND deleted_at IS NULL`, req.DeltaDays, req.DeltaDays, tripID)

	// Shift trip days
	tx.Exec(`UPDATE trip_days SET
		date = date + make_interval(days => ?)
	WHERE trip_plan = ? AND deleted_at IS NULL`, req.DeltaDays, tripID)

	// Shift activity times
	tx.Exec(`UPDATE activities SET
		start_time = start_time + make_interval(days => ?),
		end_time   = end_time   + make_interval(days => ?)
	WHERE trip_day IN (SELECT id FROM trip_days WHERE trip_plan = ?) AND deleted_at IS NULL`,
		req.DeltaDays, req.DeltaDays, tripID)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"days_shifted": req.DeltaDays})
}

// GetTodayView godoc
// @Summary Get today's itinerary for an ongoing trip
// @Description Returns the current day view including current hop/stay, activities, upcoming transport, and emergency contacts.
// @Tags lifecycle
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} TodayView "Today's view"
// @Failure 404 {object} map[string]string "Not found or trip not active"
// @Security BearerAuth
// @Router /trip/{id}/today [get]
func GetTodayView(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	var trip TripPlan
	if err := core.DB.Where("id = ?", tripID).First(&trip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}
	if trip.UserID != user.BaseModel.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this trip"})
		return
	}

	// Determine "today" in the trip's timezone.
	loc := time.UTC
	if trip.Timezone != nil {
		if l, err := time.LoadLocation(*trip.Timezone); err == nil {
			loc = l
		}
	}
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	view := TodayView{
		TripID:   trip.ID,
		TripName: trip.Name,
		Date:     todayStr,
	}

	// Find the current hop (today falls within hop's date range).
	var currentHop TripHop
	hopErr := core.DB.
		Where("trip_plan = ? AND start_date <= ? AND end_date >= ? AND deleted_at IS NULL",
			tripID, now, now).
		Order("hop_order ASC").
		First(&currentHop).Error

	if hopErr == nil {
		view.CurrentHop = &currentHop

		// Current stay (check if today is within stay dates).
		var stay Stay
		if err := core.DB.
			Where("trip_hop = ? AND start_date <= ? AND end_date >= ? AND deleted_at IS NULL",
				currentHop.ID, now, now).
			First(&stay).Error; err == nil {
			view.CurrentStay = &stay
			// Check-in flag: stay's start_date is today.
			if stay.StartDate != nil {
				stayStart := stay.StartDate.In(loc).Format("2006-01-02")
				view.CheckInToday = stayStart == todayStr
			}
			// Check-out flag: stay's end_date is today.
			if stay.EndDate != nil {
				stayEnd := stay.EndDate.In(loc).Format("2006-01-02")
				view.CheckOutToday = stayEnd == todayStr
			}
		}

		// Emergency number derived from the current hop's country.
		if currentHop.Country != nil {
			view.EmergencyNumber = getEmergencyNumber(isoCode(*currentHop.Country))
		}
	}
	if view.EmergencyNumber == "" {
		view.EmergencyNumber = "112"
	}

	// Today's TripDay + activities.
	var tripDay TripDay
	if err := core.DB.
		Preload("Activities", "deleted_at IS NULL").
		Where("trip_plan = ? AND date::date = ? AND deleted_at IS NULL", tripID, todayStr).
		First(&tripDay).Error; err == nil {
		view.TripDay = &tripDay
		view.DayNumber = tripDay.DayNumber
	}

	// Upcoming transport: next segment departing today or tomorrow.
	tomorrow := now.Add(24 * time.Hour)
	var seg TransportSegment
	if err := core.DB.
		Where("trip_plan = ? AND depart_at >= ? AND depart_at <= ? AND deleted_at IS NULL",
			tripID, now, tomorrow).
		Order("depart_at ASC").
		First(&seg).Error; err == nil {
		view.UpcomingTransport = &seg
	}

	// Emergency contacts.
	core.DB.
		Where("trip_plan = ? AND is_active = ? AND deleted_at IS NULL", tripID, true).
		Find(&view.EmergencyContacts)

	c.JSON(http.StatusOK, view)
}

// isoCode is a best-effort country-name to ISO-3166-1 alpha-2 mapper for
// the most common trip destinations. Falls back to returning the input
// (which then falls through to the "112" default in getEmergencyNumber).
func isoCode(country string) string {
	mapping := map[string]string{
		"India": "IN", "United States": "US", "USA": "US",
		"United Kingdom": "GB", "UK": "GB", "France": "FR",
		"Germany": "DE", "Italy": "IT", "Japan": "JP",
		"China": "CN", "Australia": "AU", "Canada": "CA",
		"New Zealand": "NZ", "Singapore": "SG", "UAE": "AE",
		"United Arab Emirates": "AE", "Thailand": "TH",
		"Nepal": "NP", "Sri Lanka": "LK", "Bangladesh": "BD",
	}
	if code, ok := mapping[country]; ok {
		return code
	}
	return country
}
