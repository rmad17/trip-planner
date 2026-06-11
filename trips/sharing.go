package trips

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// shareCodeAlphabet omits visually ambiguous characters (0/O, 1/I/l).
const shareCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
const shareCodeLength = 10

// generateShareCode returns a cryptographically random share code.
func generateShareCode() string {
	alphabetLen := big.NewInt(int64(len(shareCodeAlphabet)))
	code := make([]byte, shareCodeLength)
	for i := range code {
		n, _ := rand.Int(rand.Reader, alphabetLen)
		code[i] = shareCodeAlphabet[n.Int64()]
	}
	return string(code)
}

// PublicActivity is the PII-free projection of an Activity for public sharing.
// It omits booking_ref, contact_info, and actual_cost.
type PublicActivity struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Description   *string        `json:"description"`
	ActivityType  ActivityType   `json:"activity_type"`
	StartTime     *time.Time     `json:"start_time"`
	EndTime       *time.Time     `json:"end_time"`
	Duration      *int           `json:"duration"`
	Location      *string        `json:"location"`
	EstimatedCost *float64       `json:"estimated_cost"`
	Tags          pq.StringArray `json:"tags" swaggertype:"array,string"`
}

// PublicTripDay is the PII-free projection of a TripDay for public sharing.
type PublicTripDay struct {
	ID          uuid.UUID      `json:"id"`
	Date        core.Date      `json:"date"`
	DayNumber   int            `json:"day_number"`
	Title       *string        `json:"title"`
	DayType     TripDayType    `json:"day_type"`
	Notes       *string        `json:"notes"`
	Activities  []PublicActivity `json:"activities"`
}

// PublicTripHop is the PII-free projection of a TripHop for public sharing.
// It omits selected_hotel, selected_flight (provider offer IDs), stay payment info.
type PublicTripHop struct {
	ID             uuid.UUID      `json:"id"`
	Name           *string        `json:"name"`
	City           *string        `json:"city"`
	Country        *string        `json:"country"`
	StartDate      *time.Time     `json:"start_date"`
	EndDate        *time.Time     `json:"end_date"`
	Transportation *string        `json:"transportation"`
	HopOrder       *int           `json:"hop_order"`
	POIs           pq.StringArray `json:"pois" swaggertype:"array,string"`
	Days           []PublicTripDay `json:"days"`
}

// PublicTripPlan is the sanitized public view of a trip plan.
// Never includes expenses, documents, traveller PII, booking refs, or user IDs.
type PublicTripPlan struct {
	ID          uuid.UUID      `json:"id"`
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	TripType    *string        `json:"trip_type"`
	TravelModes pq.StringArray `json:"travel_modes" swaggertype:"array,string"`
	Tags        pq.StringArray `json:"tags" swaggertype:"array,string"`
	ShareCode   *string        `json:"share_code"`
	Hops        []PublicTripHop `json:"hops"`
}

// PublishTrip godoc
// @Summary Publish a trip plan for public sharing
// @Description Makes the trip publicly accessible via a share code. Generates code on first publish.
// @Tags sharing
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]string "share_code and public_url"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /trip/{id}/publish [post]
func PublishTrip(c *gin.Context) {
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

	isPublic := true
	updates := map[string]interface{}{"is_public": isPublic}

	if trip.ShareCode == nil {
		code := generateShareCode()
		// Retry on the tiny chance of a collision (partial unique index on share_code).
		for {
			var existing TripPlan
			if err := core.DB.Where("share_code = ?", code).First(&existing).Error; err != nil {
				break // no collision
			}
			code = generateShareCode()
		}
		updates["share_code"] = code
		trip.ShareCode = &code
	}
	if err := core.DB.Model(&trip).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_code": trip.ShareCode,
		"public_url": "/api/v1/public/trip/" + *trip.ShareCode,
	})
}

// UnpublishTrip godoc
// @Summary Unpublish a trip plan
// @Description Removes public access. The share code is preserved so re-publishing keeps the same URL.
// @Tags sharing
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]bool "unpublished"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /trip/{id}/unpublish [post]
func UnpublishTrip(c *gin.Context) {
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

	isPublic := false
	if err := core.DB.Model(&trip).Update("is_public", isPublic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unpublished": true})
}

// RotateShareCode godoc
// @Summary Rotate the share code for a trip
// @Description Generates a new share code, invalidating any existing shared links.
// @Tags sharing
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]string "new share_code"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Router /trip/{id}/share/rotate [post]
func RotateShareCode(c *gin.Context) {
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

	code := generateShareCode()
	for {
		var existing TripPlan
		if err := core.DB.Where("share_code = ?", code).First(&existing).Error; err != nil {
			break
		}
		code = generateShareCode()
	}
	if err := core.DB.Model(&trip).Update("share_code", code).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"share_code": code,
		"public_url": "/api/v1/public/trip/" + code,
	})
}

// GetPublicTrip godoc
// @Summary Get a publicly shared trip plan
// @Description Returns a sanitized read-only view of a trip plan by share code. No authentication required.
// @Tags sharing
// @Produce json
// @Param share_code path string true "Share Code"
// @Success 200 {object} PublicTripPlan "Sanitized trip plan"
// @Failure 404 {object} map[string]string "Not found or not public"
// @Router /public/trip/{share_code} [get]
func GetPublicTrip(c *gin.Context) {
	shareCode := c.Param("share_code")

	var trip TripPlan
	// Use identical 404 for "not found" and "not public" — don't leak existence.
	if err := core.DB.
		Preload("TripHops.TripDays.Activities").
		Where("share_code = ? AND is_public = ?", shareCode, true).
		First(&trip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, sanitizeTripForPublic(&trip))
}

// ClonePublicTrip godoc
// @Summary Clone a public trip plan
// @Description Creates a copy of a public trip (structure only) owned by the authenticated user.
// @Tags sharing
// @Produce json
// @Param share_code path string true "Share Code of the trip to clone"
// @Success 201 {object} TripPlan "Cloned trip plan"
// @Failure 404 {object} map[string]string "Not found or not public"
// @Security BearerAuth
// @Router /trip/clone/{share_code} [post]
func ClonePublicTrip(c *gin.Context) {
	shareCode := c.Param("share_code")
	user := mustGetUser(c)

	var source TripPlan
	if err := core.DB.
		Preload("TripHops.TripDays.Activities").
		Where("share_code = ? AND is_public = ?", shareCode, true).
		First(&source).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
		return
	}

	// Clone plan — strip share info, set new owner, reset status.
	cloneName := source.Name
	if cloneName != nil {
		s := "Copy of " + *cloneName
		cloneName = &s
	}
	status := TripStatusPlanning
	clone := TripPlan{
		Name:        cloneName,
		Description: source.Description,
		StartDate:   source.StartDate,
		EndDate:     source.EndDate,
		MinDays:     source.MinDays,
		MaxDays:     source.MaxDays,
		TravelModes: source.TravelModes,
		TripType:    source.TripType,
		Budget:      source.Budget,
		Currency:    source.Currency,
		Status:      &status,
		Timezone:    source.Timezone,
		Tags:        source.Tags,
		UserID:      user.BaseModel.ID,
	}

	if err := core.DB.Create(&clone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, hop := range source.TripHops {
		newHop := TripHop{
			Name:           hop.Name,
			City:           hop.City,
			Country:        hop.Country,
			Region:         hop.Region,
			Latitude:       hop.Latitude,
			Longitude:      hop.Longitude,
			StartDate:      hop.StartDate,
			EndDate:        hop.EndDate,
			Transportation: hop.Transportation,
			HopOrder:       hop.HopOrder,
			POIs:           hop.POIs,
			Notes:          hop.Notes,
			TripPlan:       clone.ID,
		}
		if err := core.DB.Create(&newHop).Error; err != nil {
			continue
		}
		for _, day := range hop.TripDays {
			newDay := TripDay{
				Date:        day.Date,
				DayNumber:   day.DayNumber,
				Title:       day.Title,
				DayType:     day.DayType,
				Notes:       day.Notes,
				TripPlan:    clone.ID,
				FromTripHop: &newHop.ID,
			}
			if err := core.DB.Create(&newDay).Error; err != nil {
				continue
			}
			for _, act := range day.Activities {
				newAct := Activity{
					Name:          act.Name,
					Description:   act.Description,
					ActivityType:  act.ActivityType,
					StartTime:     act.StartTime,
					EndTime:       act.EndTime,
					Duration:      act.Duration,
					Location:      act.Location,
					EstimatedCost: act.EstimatedCost,
					Tags:          act.Tags,
					TripDay:       newDay.ID,
					TripHop:       &newHop.ID,
				}
				core.DB.Create(&newAct)
			}
		}
	}

	c.JSON(http.StatusCreated, clone)
}

// sanitizeTripForPublic builds the PII-free public projection.
func sanitizeTripForPublic(trip *TripPlan) PublicTripPlan {
	hops := make([]PublicTripHop, 0, len(trip.TripHops))
	for _, h := range trip.TripHops {
		days := make([]PublicTripDay, 0, len(h.TripDays))
		for _, d := range h.TripDays {
			acts := make([]PublicActivity, 0, len(d.Activities))
			for _, a := range d.Activities {
				acts = append(acts, PublicActivity{
					ID:            a.ID,
					Name:          a.Name,
					Description:   a.Description,
					ActivityType:  a.ActivityType,
					StartTime:     a.StartTime,
					EndTime:       a.EndTime,
					Duration:      a.Duration,
					Location:      a.Location,
					EstimatedCost: a.EstimatedCost,
					Tags:          a.Tags,
				})
			}
			days = append(days, PublicTripDay{
				ID:         d.ID,
				Date:       d.Date,
				DayNumber:  d.DayNumber,
				Title:      d.Title,
				DayType:    d.DayType,
				Notes:      d.Notes,
				Activities: acts,
			})
		}
		hops = append(hops, PublicTripHop{
			ID:             h.ID,
			Name:           h.Name,
			City:           h.City,
			Country:        h.Country,
			StartDate:      h.StartDate,
			EndDate:        h.EndDate,
			Transportation: h.Transportation,
			HopOrder:       h.HopOrder,
			POIs:           h.POIs,
			Days:           days,
		})
	}
	return PublicTripPlan{
		ID:          trip.ID,
		Name:        trip.Name,
		Description: trip.Description,
		StartDate:   trip.StartDate,
		EndDate:     trip.EndDate,
		TripType:    trip.TripType,
		TravelModes: trip.TravelModes,
		Tags:        trip.Tags,
		ShareCode:   trip.ShareCode,
		Hops:        hops,
	}
}

// mustGetUser extracts the authenticated user from context.
// Returns the user and panics if missing (middleware guarantees it exists on auth routes).
func mustGetUser(c *gin.Context) accounts.User {
	u, _ := c.Get("currentUser")
	return u.(accounts.User)
}
