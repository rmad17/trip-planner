package trips

import (
	"net/http"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validTransportModes is the set of accepted TransportMode values.
var validTransportModes = map[TransportMode]struct{}{
	TransportModeFlight:          {},
	TransportModeTrain:           {},
	TransportModeBus:             {},
	TransportModeCarRental:       {},
	TransportModeFerry:           {},
	TransportModeTaxi:            {},
	TransportModePersonalVehicle: {},
	TransportModeOther:           {},
}

func isValidTransportMode(m TransportMode) bool {
	_, ok := validTransportModes[m]
	return ok
}

// TransportSegmentRequest is the create/update payload for a transport segment.
type TransportSegmentRequest struct {
	FromHopID   *uuid.UUID    `json:"from_hop_id"`
	ToHopID     *uuid.UUID    `json:"to_hop_id"`
	Mode        TransportMode `json:"mode" binding:"required" example:"train"`
	Operator    *string       `json:"operator" example:"IRCTC"`
	BookingRef  *string       `json:"booking_ref" example:"PNR1234567"`
	DepartAt    *string       `json:"depart_at" example:"2026-09-01T06:00:00+05:30"`
	ArriveAt    *string       `json:"arrive_at" example:"2026-09-01T22:00:00+05:30"`
	DepartFrom  *string       `json:"depart_from" example:"New Delhi Railway Station"`
	ArriveTo    *string       `json:"arrive_to" example:"Mumbai CSMT"`
	DepartTZ    *string       `json:"depart_tz" example:"Asia/Kolkata"`
	ArriveTZ    *string       `json:"arrive_tz" example:"Asia/Kolkata"`
	Cost        *float64      `json:"cost" example:"1500.00"`
	Currency    *string       `json:"currency" example:"INR"`
	Seats       *string       `json:"seats" example:"S4 42, S4 43"`
	DeepLinkURL *string       `json:"deep_link_url"`
	Notes       *string       `json:"notes"`
}

// ListTransportSegments godoc
// @Summary List transport segments for a trip
// @Description Returns all non-deleted transport segments for a trip plan, ordered by departure time.
// @Tags transport
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "segments"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /trip/{id}/transport [get]
func ListTransportSegments(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var segs []TransportSegment
	core.DB.Where("trip_plan = ?", tripID).
		Order("depart_at ASC NULLS LAST").
		Find(&segs)

	c.JSON(http.StatusOK, gin.H{"segments": segs})
}

// CreateTransportSegment godoc
// @Summary Add a transport segment to a trip
// @Description Creates a new flight, train, bus, or other transport leg.
// @Tags transport
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param body body TransportSegmentRequest true "Segment details"
// @Success 201 {object} TransportSegment
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip not found"
// @Security BearerAuth
// @Router /trip/{id}/transport [post]
func CreateTransportSegment(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req TransportSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isValidTransportMode(req.Mode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport mode", "valid": []string{
			string(TransportModeFlight), string(TransportModeTrain), string(TransportModeBus),
			string(TransportModeCarRental), string(TransportModeFerry), string(TransportModeTaxi),
			string(TransportModeOther),
		}})
		return
	}

	parsedTripID, _ := uuid.Parse(tripID)
	seg := TransportSegment{
		TripPlan:    parsedTripID,
		FromHopID:   req.FromHopID,
		ToHopID:     req.ToHopID,
		Mode:        req.Mode,
		Operator:    req.Operator,
		BookingRef:  req.BookingRef,
		DepartFrom:  req.DepartFrom,
		ArriveTo:    req.ArriveTo,
		DepartTZ:    req.DepartTZ,
		ArriveTZ:    req.ArriveTZ,
		Cost:        req.Cost,
		Currency:    req.Currency,
		Seats:       req.Seats,
		DeepLinkURL: req.DeepLinkURL,
		Notes:       req.Notes,
	}
	seg.DepartAt = parseTimePtr(req.DepartAt)
	seg.ArriveAt = parseTimePtr(req.ArriveAt)

	if err := core.DB.Create(&seg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, seg)
}

// GetTransportSegment godoc
// @Summary Get a transport segment by ID
// @Tags transport
// @Produce json
// @Param id path string true "Segment ID"
// @Success 200 {object} TransportSegment
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /transport/{id} [get]
func GetTransportSegment(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var seg TransportSegment
	if err := core.DB.Where("id = ?", id).First(&seg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}
	if err := assertTripOwner(seg.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	c.JSON(http.StatusOK, seg)
}

// UpdateTransportSegment godoc
// @Summary Update a transport segment
// @Tags transport
// @Accept json
// @Produce json
// @Param id path string true "Segment ID"
// @Param body body TransportSegmentRequest true "Updated fields"
// @Success 200 {object} TransportSegment
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /transport/{id} [put]
func UpdateTransportSegment(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var seg TransportSegment
	if err := core.DB.Where("id = ?", id).First(&seg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}
	if err := assertTripOwner(seg.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req TransportSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Mode != "" && !isValidTransportMode(req.Mode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport mode"})
		return
	}

	updates := map[string]interface{}{
		"from_hop_id":   req.FromHopID,
		"to_hop_id":     req.ToHopID,
		"operator":      req.Operator,
		"booking_ref":   req.BookingRef,
		"depart_from":   req.DepartFrom,
		"arrive_to":     req.ArriveTo,
		"depart_tz":     req.DepartTZ,
		"arrive_tz":     req.ArriveTZ,
		"cost":          req.Cost,
		"currency":      req.Currency,
		"seats":         req.Seats,
		"deep_link_url": req.DeepLinkURL,
		"notes":         req.Notes,
		"depart_at":     parseTimePtr(req.DepartAt),
		"arrive_at":     parseTimePtr(req.ArriveAt),
	}
	if req.Mode != "" {
		updates["mode"] = req.Mode
	}

	if err := core.DB.Model(&seg).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, seg)
}

// DeleteTransportSegment godoc
// @Summary Soft-delete a transport segment
// @Tags transport
// @Produce json
// @Param id path string true "Segment ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /transport/{id} [delete]
func DeleteTransportSegment(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var seg TransportSegment
	if err := core.DB.Where("id = ?", id).First(&seg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}
	if err := assertTripOwner(seg.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	core.DB.Delete(&seg)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// assertTripOwner returns nil if the user owns the trip, error otherwise.
func assertTripOwner(tripID string, userID uuid.UUID) error {
	var trip TripPlan
	return core.DB.Select("id, user_id").Where("id = ? AND user_id = ?", tripID, userID).First(&trip).Error
}

// parseTimePtr parses a RFC3339 time string pointer, returning nil on any failure.
func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	// Import needed — using a package-level import below.
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}
