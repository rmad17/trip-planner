package trips

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TRIP DAYS CRUD

// GetTripDays godoc
// @Summary Get trip days for a trip plan
// @Description Retrieve all trip days for a specific trip plan
// @Tags trip-days
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "List of trip days"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/days [get]
func GetTripDays(c *gin.Context) {
	tripPlanID := c.Param("trip_plan_id")
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

	var tripDays []TripDay
	result := core.DB.Preload("Activities").Where("trip_plan = ?", tripPlanID).Order("day_number ASC").Find(&tripDays)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_days": tripDays})
}

// GetTripDay godoc
// @Summary Get a specific trip day
// @Description Retrieve a trip day by ID with all activities
// @Tags trip-days
// @Produce json
// @Param id path string true "Trip Day ID"
// @Success 200 {object} TripDay "Trip day details"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-days/{id} [get]
func GetTripDay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripDay TripDay
	// Verify ownership through trip plan
	result := core.DB.Preload("Activities").
		Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("trip_days.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&tripDay)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_day": tripDay})
}

// CreateTripDay godoc
// @Summary Create a new trip day
// @Description Add a new day to a trip plan
// @Tags trip-days
// @Accept json
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param day body TripDay true "Trip day data"
// @Success 201 {object} TripDay "Created trip day"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/days [post]
func CreateTripDay(c *gin.Context) {
	tripPlanID := c.Param("trip_plan_id")
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

	var tripDay TripDay
	if err := c.BindJSON(&tripDay); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set trip plan ID
	tripPlanUUID, _ := uuid.Parse(tripPlanID)
	tripDay.TripPlan = tripPlanUUID

	// Set day number if not provided
	if tripDay.DayNumber == 0 {
		var maxDayNumber int
		core.DB.Model(&TripDay{}).Where("trip_plan = ?", tripPlanID).Select("COALESCE(MAX(day_number), 0)").Scan(&maxDayNumber)
		tripDay.DayNumber = maxDayNumber + 1
	}

	result := core.DB.Create(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"trip_day": tripDay})
}

// UpdateTripDay godoc
// @Summary Update a trip day
// @Description Update an existing trip day
// @Tags trip-days
// @Accept json
// @Produce json
// @Param id path string true "Trip Day ID"
// @Param day body TripDay true "Updated trip day data"
// @Success 200 {object} TripDay "Updated trip day"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-days/{id} [put]
func UpdateTripDay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripDay TripDay
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("trip_days.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	var updateData TripDay
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = core.DB.Model(&tripDay).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trip_day": tripDay})
}

// DeleteTripDay godoc
// @Summary Delete a trip day
// @Description Delete a trip day and all associated activities
// @Tags trip-days
// @Param id path string true "Trip Day ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-days/{id} [delete]
func DeleteTripDay(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripDay TripDay
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("trip_days.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	// Delete activities first, then trip day
	tx := core.DB.Begin()
	tx.Where("trip_day = ?", id).Delete(&Activity{})
	tx.Delete(&tripDay)

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
// @Summary Get activities for a trip day
// @Description Retrieve all activities for a specific trip day
// @Tags activities
// @Produce json
// @Param trip_day_id path string true "Trip Day ID"
// @Success 200 {object} map[string]interface{} "List of activities"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-days/{trip_day_id}/activities [get]
func GetActivities(c *gin.Context) {
	tripDayID := c.Param("trip_day_id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip day ownership
	var tripDay TripDay
	result := core.DB.Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("trip_days.id = ? AND trip_plans.user_id = ?", tripDayID, user.BaseModel.ID).
		First(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	var activities []Activity
	result = core.DB.Where("trip_day = ?", tripDayID).Order("start_time ASC").Find(&activities)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// CreateActivity godoc
// @Summary Create a new activity
// @Description Add a new activity to a trip day
// @Tags activities
// @Accept json
// @Produce json
// @Param trip_day_id path string true "Trip Day ID"
// @Param activity body Activity true "Activity data"
// @Success 201 {object} Activity "Created activity"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip day not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-days/{trip_day_id}/activities [post]
func CreateActivity(c *gin.Context) {
	tripDayID := c.Param("trip_day_id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip day ownership
	var tripDay TripDay
	result := core.DB.Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("trip_days.id = ? AND trip_plans.user_id = ?", tripDayID, user.BaseModel.ID).
		First(&tripDay)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip day not found"})
		return
	}

	var activity Activity
	if err := c.BindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set trip day ID
	tripDayUUID, _ := uuid.Parse(tripDayID)
	activity.TripDay = tripDayUUID

	result = core.DB.Create(&activity)
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
	// Verify ownership through trip plan
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
	// Verify ownership through trip plan
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