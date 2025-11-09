package trips

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// STAYS CRUD

// GetStays godoc
// @Summary Get stays for a trip hop
// @Description Retrieve all stays for a specific trip hop
// @Tags stays
// @Produce json
// @Param id path string true "Trip Hop ID"
// @Success 200 {object} map[string]interface{} "List of stays"
// @Failure 404 {object} map[string]string "Trip hop not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-hops/{id}/stays [get]
func GetStays(c *gin.Context) {
	tripHopID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip hop ownership through trip plan
	var tripHop TripHop
	if err := core.DB.Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("trip_hops.id = ? AND trip_plans.user_id = ?", tripHopID, user.BaseModel.ID).
		First(&tripHop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip hop not found"})
		return
	}

	var stays []Stay
	result := core.DB.Where("trip_hop = ?", tripHopID).Order("start_date ASC").Find(&stays)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stays": stays})
}

// GetStay godoc
// @Summary Get a specific stay
// @Description Retrieve a stay by ID
// @Tags stays
// @Produce json
// @Param id path string true "Stay ID"
// @Success 200 {object} Stay "Stay details"
// @Failure 404 {object} map[string]string "Stay not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /stays/{id} [get]
func GetStay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var stay Stay
	// Verify ownership through trip hop and trip plan
	result := core.DB.Joins("JOIN trip_hops ON stays.trip_hop = trip_hops.id").
		Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("stays.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&stay)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stay not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stay": stay})
}

// CreateStay godoc
// @Summary Create a new stay
// @Description Add a new stay to a trip hop
// @Tags stays
// @Accept json
// @Produce json
// @Param id path string true "Trip Hop ID"
// @Param stay body Stay true "Stay data"
// @Success 201 {object} Stay "Created stay"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip hop not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-hops/{id}/stays [post]
func CreateStay(c *gin.Context) {
	tripHopID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip hop ownership through trip plan
	var tripHop TripHop
	if err := core.DB.Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("trip_hops.id = ? AND trip_plans.user_id = ?", tripHopID, user.BaseModel.ID).
		First(&tripHop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip hop not found"})
		return
	}

	var stay Stay
	if err := c.BindJSON(&stay); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set trip hop ID
	tripHopUUID, _ := uuid.Parse(tripHopID)
	stay.TripHop = tripHopUUID

	result := core.DB.Create(&stay)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"stay": stay})
}

// UpdateStay godoc
// @Summary Update a stay
// @Description Update an existing stay
// @Tags stays
// @Accept json
// @Produce json
// @Param id path string true "Stay ID"
// @Param stay body Stay true "Updated stay data"
// @Success 200 {object} Stay "Updated stay"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Stay not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /stays/{id} [put]
func UpdateStay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var stay Stay
	// Verify ownership through trip hop and trip plan
	result := core.DB.Joins("JOIN trip_hops ON stays.trip_hop = trip_hops.id").
		Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("stays.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&stay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stay not found"})
		return
	}

	var updateData Stay
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = core.DB.Model(&stay).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stay": stay})
}

// DeleteStay godoc
// @Summary Delete a stay
// @Description Delete a stay
// @Tags stays
// @Param id path string true "Stay ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Stay not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /stays/{id} [delete]
func DeleteStay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var stay Stay
	// Verify ownership through trip hop and trip plan
	result := core.DB.Joins("JOIN trip_hops ON stays.trip_hop = trip_hops.id").
		Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("stays.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&stay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stay not found"})
		return
	}

	result = core.DB.Delete(&stay)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
