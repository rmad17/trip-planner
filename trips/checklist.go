package trips

import (
	"net/http"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var validChecklistCategories = map[ChecklistCategory]struct{}{
	ChecklistCategoryBooking:   {},
	ChecklistCategoryDocuments: {},
	ChecklistCategoryPacking:   {},
	ChecklistCategoryMoney:     {},
	ChecklistCategoryHealth:    {},
	ChecklistCategoryOther:     {},
}

func isValidChecklistCategory(cat ChecklistCategory) bool {
	_, ok := validChecklistCategories[cat]
	return ok
}

// ChecklistItemRequest is the create/update payload for a checklist item.
type ChecklistItemRequest struct {
	HopID       *uuid.UUID        `json:"hop_id"`
	Title       string            `json:"title" binding:"required" example:"Apply for French visa"`
	Category    ChecklistCategory `json:"category" example:"documents"`
	DueDate     *time.Time        `json:"due_date"`
	AssigneeID  *uuid.UUID        `json:"assignee_id"`
	IsCompleted *bool             `json:"is_completed"`
	Notes       *string           `json:"notes"`
	SortOrder   *int              `json:"sort_order"`
}

// ListChecklist godoc
// @Summary List checklist items for a trip
// @Description Returns all checklist items, optionally filtered by category or hop.
// @Tags checklist
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param category query string false "Filter by category"
// @Param completed query bool false "Filter by completion status"
// @Success 200 {object} map[string]interface{} "items"
// @Security BearerAuth
// @Router /trip/{id}/checklist [get]
func ListChecklist(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	q := core.DB.Where("trip_plan = ?", tripID)
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if completed := c.Query("completed"); completed == "true" {
		q = q.Where("is_completed = ?", true)
	} else if completed == "false" {
		q = q.Where("is_completed = ?", false)
	}

	var items []ChecklistItem
	q.Order("sort_order ASC, created_at ASC").Find(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateChecklistItem godoc
// @Summary Add a checklist item to a trip
// @Tags checklist
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param body body ChecklistItemRequest true "Checklist item"
// @Success 201 {object} ChecklistItem
// @Failure 400 {object} map[string]string "Bad request"
// @Security BearerAuth
// @Router /trip/{id}/checklist [post]
func CreateChecklistItem(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req ChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat := req.Category
	if cat == "" {
		cat = ChecklistCategoryOther
	}
	if !isValidChecklistCategory(cat) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	parsedID, _ := uuid.Parse(tripID)
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	item := ChecklistItem{
		TripPlan:   parsedID,
		HopID:      req.HopID,
		Title:      req.Title,
		Category:   cat,
		DueDate:    req.DueDate,
		AssigneeID: req.AssigneeID,
		Notes:      req.Notes,
		SortOrder:  sortOrder,
	}
	if err := core.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// UpdateChecklistItem godoc
// @Summary Update a checklist item
// @Tags checklist
// @Accept json
// @Produce json
// @Param id path string true "Checklist Item ID"
// @Param body body ChecklistItemRequest true "Updated item"
// @Success 200 {object} ChecklistItem
// @Security BearerAuth
// @Router /checklist/{id} [put]
func UpdateChecklistItem(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var item ChecklistItem
	if err := core.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Checklist item not found"})
		return
	}
	if err := assertTripOwner(item.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req ChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"due_date":    req.DueDate,
		"assignee_id": req.AssigneeID,
		"notes":       req.Notes,
	}
	if req.Category != "" && isValidChecklistCategory(req.Category) {
		updates["category"] = req.Category
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if err := core.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteChecklistItem godoc
// @Summary Delete a checklist item
// @Tags checklist
// @Produce json
// @Param id path string true "Checklist Item ID"
// @Success 200 {object} map[string]bool "deleted"
// @Security BearerAuth
// @Router /checklist/{id} [delete]
func DeleteChecklistItem(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var item ChecklistItem
	if err := core.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Checklist item not found"})
		return
	}
	if err := assertTripOwner(item.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	core.DB.Delete(&item)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ToggleChecklistItem godoc
// @Summary Toggle completion of a checklist item
// @Description Flips is_completed; sets completed_at when marking done, clears when undoing.
// @Tags checklist
// @Produce json
// @Param id path string true "Checklist Item ID"
// @Success 200 {object} ChecklistItem
// @Security BearerAuth
// @Router /checklist/{id}/toggle [patch]
func ToggleChecklistItem(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var item ChecklistItem
	if err := core.DB.Where("id = ?", id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Checklist item not found"})
		return
	}
	if err := assertTripOwner(item.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	now := time.Now()
	newDone := !item.IsCompleted
	updates := map[string]interface{}{"is_completed": newDone}
	if newDone {
		updates["completed_at"] = now
	} else {
		updates["completed_at"] = nil
	}

	core.DB.Model(&item).Updates(updates)
	c.JSON(http.StatusOK, item)
}
