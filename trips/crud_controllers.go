package trips

import (
	"net/http"
	"strconv"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TRIP PLANS CRUD

// GetTripPlans godoc
// @Summary Get all trip plans for the authenticated user
// @Description Retrieve all trip plans created by the current user with optional pagination
// @Tags trip-plans
// @Produce json
// @Param limit query int false "Number of records to return (default: 50)"
// @Param offset query int false "Number of records to skip (default: 0)"
// @Success 200 {object} map[string]interface{} "List of trip plans"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans [get]
func GetTripPlans(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var tripPlans []TripPlan
	var count int64

	// Get total count
	core.DB.Model(&TripPlan{}).Where("user_id = ?", user.BaseModel.ID).Count(&count)

	// Get trip plans with related data
	result := core.DB.Preload("TripHops").Preload("TripDays").Preload("Travellers").
		Where("user_id = ?", user.BaseModel.ID).
		Limit(limit).Offset(offset).Find(&tripPlans)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trip_plans": tripPlans,
		"total":      count,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetTripPlan godoc
// @Summary Get a specific trip plan
// @Description Retrieve a trip plan by ID with all related data
// @Tags trip-plans
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} TripPlan "Trip plan details"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id} [get]
func GetTripPlan(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	result := core.DB.Preload("TripHops.Stays").Preload("TripDays.Activities").Preload("Travellers").
		Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&tripPlan)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_plan": tripPlan})
}

// GetTripPlanComplete godoc
// @Summary Get complete trip details with all related data
// @Description Retrieve a trip plan with all hops, days, travellers, stays and activities in a single response
// @Tags trip-plans
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "Complete trip plan details"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id}/complete [get]
func GetTripPlanComplete(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Get trip plan with basic data
	var tripPlan TripPlan
	result := core.DB.Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	// Get trip hops with stays
	var tripHops []TripHop
	core.DB.Preload("Stays").Where("trip_plan = ?", id).Order("hop_order ASC").Find(&tripHops)

	// Get trip days with activities
	var tripDays []TripDay
	core.DB.Preload("Activities").Where("trip_plan = ?", id).Order("day_number ASC").Find(&tripDays)

	// Get travellers
	var travellers []Traveller
	core.DB.Where("trip_plan = ?", id).Find(&travellers)

	// Compile complete response
	response := gin.H{
		"trip_plan":  tripPlan,
		"hops":       tripHops,
		"days":       tripDays,
		"travellers": travellers,
		"summary": gin.H{
			"total_hops":       len(tripHops),
			"total_days":       len(tripDays),
			"total_travellers": len(travellers),
			"total_stays":      countStays(tripHops),
			"total_activities": countActivities(tripDays),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions for counting nested items
func countStays(hops []TripHop) int {
	count := 0
	for _, hop := range hops {
		count += len(hop.Stays)
	}
	return count
}

func countActivities(days []TripDay) int {
	count := 0
	for _, day := range days {
		count += len(day.Activities)
	}
	return count
}

// UpdateTripPlan godoc
// @Summary Update a trip plan
// @Description Update an existing trip plan
// @Tags trip-plans
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param trip body TripPlan true "Updated trip plan data"
// @Success 200 {object} TripPlan "Updated trip plan"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id} [put]
func UpdateTripPlan(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	result := core.DB.Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var updateData TripPlan
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	result = core.DB.Model(&tripPlan).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_plan": tripPlan})
}

// DeleteTripPlan godoc
// @Summary Delete a trip plan
// @Description Delete a trip plan and all related data
// @Tags trip-plans
// @Param id path string true "Trip Plan ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id} [delete]
func DeleteTripPlan(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	result := core.DB.Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&tripPlan)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	// Delete all related data (cascading delete)
	tx := core.DB.Begin()

	// Delete activities first
	tx.Where("trip_hop IN (SELECT id FROM trip_hops WHERE trip_plan = ?)", id).Delete(&Activity{})
	// Delete stays
	tx.Where("trip_hop IN (SELECT id FROM trip_hops WHERE trip_plan = ?)", id).Delete(&Stay{})
	// Delete trip days
	tx.Where("trip_plan = ?", id).Delete(&TripDay{})
	// Delete trip hops
	tx.Where("trip_plan = ?", id).Delete(&TripHop{})
	// Delete travellers
	tx.Where("trip_plan = ?", id).Delete(&Traveller{})
	// Finally delete trip plan
	tx.Delete(&tripPlan)

	if tx.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

// TRIP HOPS CRUD

// GetTripHops godoc
// @Summary Get trip hops for a trip plan
// @Description Retrieve all trip hops for a specific trip plan
// @Tags trip-hops
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "List of trip hops"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/hops [get]
func GetTripHops(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan ownership
	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var tripHops []TripHop
	result := core.DB.Preload("Stays").Where("trip_plan = ?", tripPlanID).Order("hop_order ASC").Find(&tripHops)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_hops": tripHops})
}

// CreateTripHop godoc
// @Summary Create a new trip hop
// @Description Add a new hop to a trip plan
// @Tags trip-hops
// @Accept json
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param hop body TripHop true "Trip hop data"
// @Success 201 {object} TripHop "Created trip hop"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/hops [post]
func CreateTripHop(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan ownership
	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var tripHop TripHop
	if err := c.BindJSON(&tripHop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set trip plan ID
	tripPlanUUID, _ := uuid.Parse(tripPlanID)
	tripHop.TripPlan = tripPlanUUID

	// Set hop order if not provided
	if tripHop.HopOrder == nil {
		var maxOrder int
		core.DB.Model(&TripHop{}).Where("trip_plan = ?", tripPlanID).Select("COALESCE(MAX(hop_order), 0)").Scan(&maxOrder)
		newOrder := maxOrder + 1
		tripHop.HopOrder = &newOrder
	}

	result := core.DB.Create(&tripHop)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip_hop": tripHop})
}

// UpdateTripHop godoc
// @Summary Update a trip hop
// @Description Update an existing trip hop
// @Tags trip-hops
// @Accept json
// @Produce json
// @Param id path string true "Trip Hop ID"
// @Param hop body TripHop true "Updated trip hop data"
// @Success 200 {object} TripHop "Updated trip hop"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip hop not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-hops/{id} [put]
func UpdateTripHop(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripHop TripHop
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("trip_hops.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&tripHop)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip hop not found"})
		return
	}

	var updateData TripHop
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = core.DB.Model(&tripHop).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_hop": tripHop})
}

// DeleteTripHop godoc
// @Summary Delete a trip hop
// @Description Delete a trip hop and all related data
// @Tags trip-hops
// @Param id path string true "Trip Hop ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Trip hop not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-hops/{id} [delete]
func DeleteTripHop(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripHop TripHop
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON trip_hops.trip_plan = trip_plans.id").
		Where("trip_hops.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&tripHop)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip hop not found"})
		return
	}

	// Delete related data
	tx := core.DB.Begin()
	tx.Where("trip_hop = ?", id).Delete(&Activity{})
	tx.Where("trip_hop = ?", id).Delete(&Stay{})
	tx.Delete(&tripHop)

	if tx.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

// ACTIVITIES CRUD

// GetActivities godoc
// @Summary Get activities for a trip plan
// @Description Retrieve all activities for a specific trip plan across all days
// @Tags activities
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "List of activities"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id}/activities [get]
func GetActivities(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan ownership
	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var activities []Activity
	result := core.DB.Joins("JOIN trip_days ON activities.trip_day = trip_days.id").
		Where("trip_days.trip_plan = ?", tripPlanID).
		Order("trip_days.day_number ASC, activities.start_time ASC").
		Find(&activities)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// CreateActivity godoc
// @Summary Create a new activity
// @Description Add a new activity to a trip plan
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param activity body Activity true "Activity data"
// @Success 201 {object} Activity "Created activity"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{id}/activities [post]
func CreateActivity(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Parse trip plan ID to UUID
	tripPlanUUID, err := uuid.Parse(tripPlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID"})
		return
	}

	// Verify trip plan ownership
	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanUUID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var activity Activity
	if err := c.BindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the trip day belongs to the trip plan
	var tripDay TripDay
	if err := core.DB.Where("id = ? AND trip_plan = ?", activity.TripDay, tripPlanUUID).First(&tripDay).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip day for this trip plan"})
		return
	}

	result := core.DB.Create(&activity)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"activity": activity})
}

// UpdateActivity godoc
// @Summary Update an activity
// @Description Update an existing activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param activity body Activity true "Updated activity data"
// @Success 200 {object} Activity "Updated activity"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Activity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /activities/{id} [put]
func UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var activity Activity
	// Verify ownership through trip day and trip plan
	result := core.DB.Joins("JOIN trip_days ON activities.trip_day = trip_days.id").
		Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("activities.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&activity)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
		return
	}

	var updateData Activity
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = core.DB.Model(&activity).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activity": activity})
}

// DeleteActivity godoc
// @Summary Delete an activity
// @Description Delete an activity
// @Tags activities
// @Param id path string true "Activity ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Activity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /activities/{id} [delete]
func DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var activity Activity
	// Verify ownership through trip day and trip plan
	result := core.DB.Joins("JOIN trip_days ON activities.trip_day = trip_days.id").
		Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("activities.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&activity)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
		return
	}

	result = core.DB.Delete(&activity)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
