package notifications

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Controller handles HTTP requests for notifications
type Controller struct {
	service *Service
}

// NewController creates a new notification controller
func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// SendNotification sends a single notification
// @Summary Send a notification
// @Description Send a single notification through specified channel
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body SendRequest true "Notification details"
// @Success 200 {object} Notification
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/send [post]
func (c *Controller) SendNotification(ctx *gin.Context) {
	var req SendRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := c.service.Send(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": notification})
}

// SendBatchNotifications sends multiple notifications
// @Summary Send batch notifications
// @Description Send multiple notifications in a batch
// @Tags notifications
// @Accept json
// @Produce json
// @Param batch body BatchSendRequest true "Batch notification details"
// @Success 200 {object} NotificationBatch
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/send/batch [post]
func (c *Controller) SendBatchNotifications(ctx *gin.Context) {
	var req BatchSendRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch, err := c.service.SendBatch(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": batch})
}

// SendFromTemplate sends a notification from a template
// @Summary Send notification from template
// @Description Send a notification using a predefined template
// @Tags notifications
// @Accept json
// @Produce json
// @Param template body TemplateSendRequest true "Template notification details"
// @Success 200 {object} Notification
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/send/template [post]
func (c *Controller) SendFromTemplate(ctx *gin.Context) {
	var req TemplateSendRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := c.service.SendFromTemplate(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": notification})
}

// GetNotification gets a notification by ID
// @Summary Get notification
// @Description Get notification details by ID
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} Notification
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /notifications/{id} [get]
func (c *Controller) GetNotification(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	notification, err := c.service.GetStatus(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": notification})
}

// GetAuditTrail gets the audit trail for a notification
// @Summary Get notification audit trail
// @Description Get complete audit trail for a notification
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} []NotificationAudit
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id}/audit [get]
func (c *Controller) GetAuditTrail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	audits, err := c.service.GetAuditTrail(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": audits})
}

// CancelNotification cancels a pending notification
// @Summary Cancel notification
// @Description Cancel a pending or scheduled notification
// @Tags notifications
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id}/cancel [post]
func (c *Controller) CancelNotification(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := c.service.Cancel(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "notification cancelled successfully"})
}

// RetryNotification retries a failed notification
// @Summary Retry notification
// @Description Retry a failed notification
// @Tags notifications
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id}/retry [post]
func (c *Controller) RetryNotification(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := c.service.Retry(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "notification retry initiated"})
}

// ListNotifications lists notifications with filters
// @Summary List notifications
// @Description List notifications with optional filters
// @Tags notifications
// @Produce json
// @Param channel query string false "Filter by channel"
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Param recipient_id query string false "Filter by recipient ID"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} []Notification
// @Failure 500 {object} map[string]interface{}
// @Router /notifications [get]
func (c *Controller) ListNotifications(ctx *gin.Context) {
	filters := NotificationFilters{
		Limit:  50,
		Offset: 0,
	}

	// Parse query parameters
	if channel := ctx.Query("channel"); channel != "" {
		ch := NotificationChannel(channel)
		filters.Channel = &ch
	}

	if notifType := ctx.Query("type"); notifType != "" {
		nt := NotificationType(notifType)
		filters.Type = &nt
	}

	if status := ctx.Query("status"); status != "" {
		st := NotificationStatus(status)
		filters.Status = &st
	}

	if recipientID := ctx.Query("recipient_id"); recipientID != "" {
		id, err := uuid.Parse(recipientID)
		if err == nil {
			filters.RecipientID = &id
		}
	}

	if limit := ctx.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			filters.Limit = val
		}
	}

	if offset := ctx.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			filters.Offset = val
		}
	}

	// Query database
	var notifications []Notification
	query := c.service.db.WithContext(ctx.Request.Context())

	if filters.Channel != nil {
		query = query.Where("channel = ?", *filters.Channel)
	}
	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.RecipientID != nil {
		query = query.Where("recipient_id = ?", *filters.RecipientID)
	}

	if err := query.Limit(filters.Limit).Offset(filters.Offset).
		Order("created_at DESC").Find(&notifications).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": notifications})
}

// Template Controllers

// CreateTemplate creates a new notification template
// @Summary Create template
// @Description Create a new notification template
// @Tags templates
// @Accept json
// @Produce json
// @Param template body NotificationTemplate true "Template details"
// @Success 201 {object} NotificationTemplate
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/templates [post]
func (c *Controller) CreateTemplate(ctx *gin.Context) {
	var template NotificationTemplate

	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.templateService.Create(ctx.Request.Context(), &template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": template})
}

// GetTemplate gets a template by ID
// @Summary Get template
// @Description Get template details by ID
// @Tags templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} NotificationTemplate
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /notifications/templates/{id} [get]
func (c *Controller) GetTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := c.service.templateService.Get(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": template})
}

// ListTemplates lists all templates
// @Summary List templates
// @Description List all notification templates
// @Tags templates
// @Produce json
// @Success 200 {object} []NotificationTemplate
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/templates [get]
func (c *Controller) ListTemplates(ctx *gin.Context) {
	templates, err := c.service.templateService.List(ctx.Request.Context(), TemplateFilters{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": templates})
}

// UpdateTemplate updates a template
// @Summary Update template
// @Description Update an existing template
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param template body NotificationTemplate true "Template details"
// @Success 200 {object} NotificationTemplate
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/templates/{id} [put]
func (c *Controller) UpdateTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var template NotificationTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.ID = id

	if err := c.service.templateService.Update(ctx.Request.Context(), &template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": template})
}

// DeleteTemplate deletes a template
// @Summary Delete template
// @Description Delete a template
// @Tags templates
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/templates/{id} [delete]
func (c *Controller) DeleteTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	if err := c.service.templateService.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "template deleted successfully"})
}

// Preference Controllers

// GetPreferences gets user notification preferences
// @Summary Get user preferences
// @Description Get notification preferences for a user
// @Tags preferences
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} []NotificationPreference
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/preferences/{user_id} [get]
func (c *Controller) GetPreferences(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	prefs, err := c.service.prefService.List(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": prefs})
}

// SetPreference sets user notification preference
// @Summary Set user preference
// @Description Set notification preference for a user
// @Tags preferences
// @Accept json
// @Produce json
// @Param preference body NotificationPreference true "Preference details"
// @Success 200 {object} NotificationPreference
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/preferences [post]
func (c *Controller) SetPreference(ctx *gin.Context) {
	var pref NotificationPreference

	if err := ctx.ShouldBindJSON(&pref); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.prefService.Set(ctx.Request.Context(), &pref); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": pref})
}
