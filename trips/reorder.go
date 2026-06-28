package trips

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
)

type reorderRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// ReorderTripHops reorders trip hops by assigning hop_order based on the submitted ID sequence.
// PUT /trip-plans/:id/hops/reorder  body: {"ids": ["uuid1", "uuid2", ...]}
func ReorderTripHops(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req reorderRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := core.DB.Begin()
	for i, id := range req.IDs {
		order := i + 1
		if err := tx.Model(&TripHop{}).
			Where("id = ? AND trip_plan = ?", id, tripPlanID).
			Update("hop_order", order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	var hops []TripHop
	core.DB.Where("trip_plan = ?", tripPlanID).Order("hop_order ASC").Find(&hops)
	c.JSON(http.StatusOK, gin.H{"trip_hops": hops})
}

// ReorderTripDays reorders trip days by assigning day_number based on the submitted ID sequence.
// PUT /trip-plans/:id/days/reorder  body: {"ids": ["uuid1", "uuid2", ...]}
func ReorderTripDays(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req reorderRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := core.DB.Begin()
	for i, id := range req.IDs {
		dayNumber := i + 1
		if err := tx.Model(&TripDay{}).
			Where("id = ? AND trip_plan = ?", id, tripPlanID).
			Update("day_number", dayNumber).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	var days []TripDay
	core.DB.Where("trip_plan = ?", tripPlanID).Order("day_number ASC").Find(&days)
	c.JSON(http.StatusOK, gin.H{"trip_days": days})
}

// ReorderActivities reorders activities within a trip day via sort_order.
// PUT /trip-plans/:id/activities/reorder  body: {"ids": ["uuid1", "uuid2", ...]}
func ReorderActivities(c *gin.Context) {
	tripPlanID := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var tripPlan TripPlan
	if err := core.DB.Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req reorderRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := core.DB.Begin()
	for i, id := range req.IDs {
		if err := tx.Model(&Activity{}).
			Where("id = ?", id).
			Update("sort_order", i).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "activities reordered"})
}

// UpdateActivityStatus updates only the status field of an activity.
// PATCH /activities/:id/status  body: {"status": "done"}
func UpdateActivityStatus(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var activity Activity
	if err := core.DB.
		Joins("JOIN trip_days ON activities.trip_day = trip_days.id").
		Joins("JOIN trip_plans ON trip_days.trip_plan = trip_plans.id").
		Where("activities.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&activity).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid := map[string]bool{
		"planned": true, "confirmed": true, "in_progress": true,
		"done": true, "completed": true, "skipped": true, "cancelled": true,
	}
	if !valid[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid status; must be one of: planned, confirmed, in_progress, done, completed, skipped, cancelled",
		})
		return
	}

	if err := core.DB.Model(&activity).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activity": activity})
}
