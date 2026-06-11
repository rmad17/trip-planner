package trips

import (
	"net/http"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var validContactRoles = map[ContactRole]struct{}{
	ContactRoleHotel:     {},
	ContactRoleGuide:     {},
	ContactRoleDriver:    {},
	ContactRoleEmbassy:   {},
	ContactRoleInsurance: {},
	ContactRoleEmergency: {},
	ContactRoleOther:     {},
}

func isValidContactRole(r ContactRole) bool {
	_, ok := validContactRoles[r]
	return ok
}

// TripContactRequest is the create/update payload for a trip contact.
type TripContactRequest struct {
	HopID    *uuid.UUID  `json:"hop_id"`
	Name     string      `json:"name" binding:"required" example:"Taj Hotel Front Desk"`
	Role     ContactRole `json:"role" binding:"required" example:"hotel"`
	Phone    *string     `json:"phone" example:"+91-22-66651234"`
	Email    *string     `json:"email"`
	Address  *string     `json:"address"`
	Notes    *string     `json:"notes"`
	IsActive *bool       `json:"is_active"`
}

// ListContacts godoc
// @Summary List contacts for a trip
// @Description Returns all active contacts for a trip plan.
// @Tags contacts
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param role query string false "Filter by role (hotel, guide, driver, embassy, insurance, emergency, other)"
// @Success 200 {object} map[string]interface{} "contacts"
// @Security BearerAuth
// @Router /trip/{id}/contacts [get]
func ListContacts(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	q := core.DB.Where("trip_plan = ?", tripID)
	if role := c.Query("role"); role != "" {
		q = q.Where("role = ?", role)
	}

	var contacts []TripContact
	q.Order("role ASC, name ASC").Find(&contacts)
	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}

// CreateContact godoc
// @Summary Add a contact to a trip
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param body body TripContactRequest true "Contact details"
// @Success 201 {object} TripContact
// @Failure 400 {object} map[string]string "Bad request"
// @Security BearerAuth
// @Router /trip/{id}/contacts [post]
func CreateContact(c *gin.Context) {
	tripID := c.Param("id")
	user := mustGetUser(c)

	if err := assertTripOwner(tripID, user.BaseModel.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found"})
		return
	}

	var req TripContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isValidContactRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	parsedID, _ := uuid.Parse(tripID)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	contact := TripContact{
		TripPlan: parsedID,
		HopID:    req.HopID,
		Name:     req.Name,
		Role:     req.Role,
		Phone:    req.Phone,
		Email:    req.Email,
		Address:  req.Address,
		Notes:    req.Notes,
		IsActive: isActive,
	}
	if err := core.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

// UpdateContact godoc
// @Summary Update a trip contact
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param body body TripContactRequest true "Updated contact"
// @Success 200 {object} TripContact
// @Security BearerAuth
// @Router /contacts/{id} [put]
func UpdateContact(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var contact TripContact
	if err := core.DB.Where("id = ?", id).First(&contact).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}
	if err := assertTripOwner(contact.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req TripContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":    req.Name,
		"phone":   req.Phone,
		"email":   req.Email,
		"address": req.Address,
		"notes":   req.Notes,
	}
	if req.Role != "" && isValidContactRole(req.Role) {
		updates["role"] = req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.HopID != nil {
		updates["hop_id"] = req.HopID
	}

	if err := core.DB.Model(&contact).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

// DeleteContact godoc
// @Summary Soft-delete a trip contact
// @Tags contacts
// @Produce json
// @Param id path string true "Contact ID"
// @Success 200 {object} map[string]bool "deleted"
// @Security BearerAuth
// @Router /contacts/{id} [delete]
func DeleteContact(c *gin.Context) {
	id := c.Param("id")
	user := mustGetUser(c)

	var contact TripContact
	if err := core.DB.Where("id = ?", id).First(&contact).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}
	if err := assertTripOwner(contact.TripPlan.String(), user.BaseModel.ID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	core.DB.Delete(&contact)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
