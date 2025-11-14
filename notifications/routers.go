package notifications

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all notification routes
func RegisterRoutes(router *gin.RouterGroup, controller *Controller) {
	// Notification routes
	notifGroup := router.Group("/notifications")
	{
		// Send notifications
		notifGroup.POST("/send", controller.SendNotification)
		notifGroup.POST("/send/batch", controller.SendBatchNotifications)
		notifGroup.POST("/send/template", controller.SendFromTemplate)

		// Notification management
		notifGroup.GET("", controller.ListNotifications)
		notifGroup.GET("/:id", controller.GetNotification)
		notifGroup.GET("/:id/audit", controller.GetAuditTrail)
		notifGroup.POST("/:id/cancel", controller.CancelNotification)
		notifGroup.POST("/:id/retry", controller.RetryNotification)

		// Template routes
		templateGroup := notifGroup.Group("/templates")
		{
			templateGroup.POST("", controller.CreateTemplate)
			templateGroup.GET("", controller.ListTemplates)
			templateGroup.GET("/:id", controller.GetTemplate)
			templateGroup.PUT("/:id", controller.UpdateTemplate)
			templateGroup.DELETE("/:id", controller.DeleteTemplate)
		}

		// Preference routes
		preferenceGroup := notifGroup.Group("/preferences")
		{
			preferenceGroup.GET("/:user_id", controller.GetPreferences)
			preferenceGroup.POST("", controller.SetPreference)
		}
	}
}

// RegisterPublicRoutes registers public notification routes (for webhooks, callbacks)
func RegisterPublicRoutes(router *gin.RouterGroup, controller *Controller) {
	// Public routes for webhook callbacks from providers
	webhookGroup := router.Group("/webhooks/notifications")
	{
		// Provider-specific webhook endpoints
		webhookGroup.POST("/sendgrid", controller.HandleSendGridWebhook)
		webhookGroup.POST("/twilio", controller.HandleTwilioWebhook)
		webhookGroup.POST("/firebase", controller.HandleFirebaseWebhook)
	}
}

// HandleSendGridWebhook handles SendGrid webhook events
func (c *Controller) HandleSendGridWebhook(ctx *gin.Context) {
	// Parse SendGrid webhook payload
	// Update notification status based on events (delivered, bounced, etc.)
	ctx.JSON(200, gin.H{"message": "webhook received"})
}

// HandleTwilioWebhook handles Twilio webhook events
func (c *Controller) HandleTwilioWebhook(ctx *gin.Context) {
	// Parse Twilio webhook payload
	// Update notification status based on events
	ctx.JSON(200, gin.H{"message": "webhook received"})
}

// HandleFirebaseWebhook handles Firebase webhook events
func (c *Controller) HandleFirebaseWebhook(ctx *gin.Context) {
	// Parse Firebase webhook payload
	// Update notification status based on events
	ctx.JSON(200, gin.H{"message": "webhook received"})
}
