package trips

import (
	"net/http"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TRAVELLERS CRUD

// GetTravellers godoc
// @Summary Get travellers for a trip plan
// @Description Retrieve all travellers for a specific trip plan
// @Tags travellers
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "List of travellers"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/travellers [get]
func GetTravellers(c *gin.Context) {
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

	var travellers []Traveller
	result := core.DB.Where("trip_plan = ? AND is_active = ?", tripPlanID, true).Order("joined_at ASC").Find(&travellers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"travellers": travellers})
}

// GetTraveller godoc
// @Summary Get a specific traveller
// @Description Retrieve a traveller by ID
// @Tags travellers
// @Produce json
// @Param id path string true "Traveller ID"
// @Success 200 {object} Traveller "Traveller details"
// @Failure 404 {object} map[string]string "Traveller not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /travellers/{id} [get]
func GetTraveller(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var traveller Traveller
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON travellers.trip_plan = trip_plans.id").
		Where("travellers.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&traveller)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Traveller not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"traveller": traveller})
}

// CreateTraveller godoc
// @Summary Create a new traveller
// @Description Add a new traveller to a trip plan
// @Tags travellers
// @Accept json
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param traveller body Traveller true "Traveller data"
// @Success 201 {object} Traveller "Created traveller"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/travellers [post]
func CreateTraveller(c *gin.Context) {
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

	var traveller Traveller
	if err := c.BindJSON(&traveller); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set trip plan ID and joined date
	tripPlanUUID, _ := uuid.Parse(tripPlanID)
	traveller.TripPlan = tripPlanUUID
	traveller.JoinedAt = time.Now()

	// Set default values
	if !traveller.IsActive && c.PostForm("is_active") == "" {
		traveller.IsActive = true
	}

	result := core.DB.Create(&traveller)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"traveller": traveller})
}

// UpdateTraveller godoc
// @Summary Update a traveller
// @Description Update an existing traveller
// @Tags travellers
// @Accept json
// @Produce json
// @Param id path string true "Traveller ID"
// @Param traveller body Traveller true "Updated traveller data"
// @Success 200 {object} Traveller "Updated traveller"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Traveller not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /travellers/{id} [put]
func UpdateTraveller(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var traveller Traveller
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON travellers.trip_plan = trip_plans.id").
		Where("travellers.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&traveller)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Traveller not found"})
		return
	}

	var updateData Traveller
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = core.DB.Model(&traveller).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"traveller": traveller})
}

// DeleteTraveller godoc
// @Summary Delete a traveller
// @Description Remove a traveller from a trip plan (soft delete by setting is_active=false)
// @Tags travellers
// @Param id path string true "Traveller ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Traveller not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /travellers/{id} [delete]
func DeleteTraveller(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var traveller Traveller
	// Verify ownership through trip plan
	result := core.DB.Joins("JOIN trip_plans ON travellers.trip_plan = trip_plans.id").
		Where("travellers.id = ? AND trip_plans.user_id = ?", id, user.BaseModel.ID).
		First(&traveller)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Traveller not found"})
		return
	}

	// Soft delete - set is_active to false instead of hard delete
	// This preserves expense records and other relationships
	result = core.DB.Model(&traveller).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// InviteTraveller godoc
// @Summary Invite a traveller via email
// @Description Send an invitation to a traveller to join a trip
// @Tags travellers
// @Accept json
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param invitation body TravellerInvitation true "Invitation data"
// @Success 201 {object} map[string]interface{} "Invitation sent"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/travellers/invite [post]
func InviteTraveller(c *gin.Context) {
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

	var invitation TravellerInvitation
	if err := c.BindJSON(&invitation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if traveller with this email already exists in this trip
	var existingTraveller Traveller
	result := core.DB.Where("trip_plan = ? AND email = ? AND is_active = ?", tripPlanID, invitation.Email, true).First(&existingTraveller)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Traveller with this email already exists in this trip"})
		return
	}

	// Create traveller record with pending status
	newTraveller := Traveller{
		FirstName: invitation.FirstName,
		LastName:  invitation.LastName,
		Email:     &invitation.Email,
		Role:      &invitation.Role,
		Notes:     &invitation.Message,
		TripPlan:  uuid.MustParse(tripPlanID),
		JoinedAt:  time.Now(),
		IsActive:  true,
	}

	result = core.DB.Create(&newTraveller)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// TODO: Send email invitation (implement email service)
	// For now, just return success response

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Invitation sent successfully",
		"traveller": newTraveller,
		"note":      "Email invitation feature is pending implementation",
	})
}

// TravellerInvitation represents an invitation request
type TravellerInvitation struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role" binding:"required"`
	Message   string `json:"message"`
}
