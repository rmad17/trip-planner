package subscriptions

import (
	"net/http"
	"triplanner/accounts"
	"triplanner/featureflags"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Middleware provides subscription-based access control
type Middleware struct {
	subscriptionService *Service
	featureFlagService  *featureflags.Service
}

// NewMiddleware creates a new subscription middleware
func NewMiddleware(subscriptionService *Service, featureFlagService *featureflags.Service) *Middleware {
	return &Middleware{
		subscriptionService: subscriptionService,
		featureFlagService:  featureFlagService,
	}
}

// RequireSubscriptionTier ensures user has at least the specified tier
func (m *Middleware) RequireSubscriptionTier(minTier SubscriptionTier) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		user := currentUser.(accounts.User)

		subscription, err := m.subscriptionService.GetUserSubscription(user.BaseModel.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
			c.Abort()
			return
		}

		if !hasRequiredTier(subscription.Plan.Tier, minTier) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Subscription tier insufficient",
				"required_tier": minTier,
				"current_tier":  subscription.Plan.Tier,
			})
			c.Abort()
			return
		}

		// Store subscription in context for later use
		c.Set("subscription", subscription)
		c.Next()
	}
}

// CheckLimit middleware checks if user can perform action based on limit type
func (m *Middleware) CheckLimit(limitType string, getValue func(*gin.Context) int) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		user := currentUser.(accounts.User)

		currentValue := getValue(c)
		err := m.subscriptionService.CheckLimit(user.BaseModel.ID, limitType, currentValue)
		if err != nil {
			if err == ErrLimitExceeded {
				subscription, _ := m.subscriptionService.GetUserSubscription(user.BaseModel.ID)
				c.JSON(http.StatusForbidden, gin.H{
					"error":        "Subscription limit exceeded",
					"limit_type":   limitType,
					"current_tier": subscription.Plan.Tier,
					"upgrade_url":  "/api/v1/subscriptions/upgrade",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireFeature middleware checks if a feature is enabled for the user
func (m *Middleware) RequireFeature(featureKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}
		user := currentUser.(accounts.User)

		// Create evaluation context
		endpoint := c.Request.URL.Path
		method := c.Request.Method
		ctx := featureflags.EvaluationContext{
			UserID:      &user.BaseModel.ID,
			APIEndpoint: &endpoint,
			HTTPMethod:  &method,
		}

		// Check if feature is enabled
		isEnabled, err := m.featureFlagService.IsEnabled(featureKey, ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature flag"})
			c.Abort()
			return
		}

		if !isEnabled {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Feature not available",
				"feature": featureKey,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InjectSubscriptionContext adds subscription info to context for all requests
func (m *Middleware) InjectSubscriptionContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if exists {
			user := currentUser.(accounts.User)
			subscription, err := m.subscriptionService.GetUserSubscription(user.BaseModel.ID)
			if err == nil {
				c.Set("subscription", subscription)
			}
		}
		c.Next()
	}
}

// hasRequiredTier checks if current tier meets the minimum required tier
func hasRequiredTier(currentTier, requiredTier SubscriptionTier) bool {
	tierLevels := map[SubscriptionTier]int{
		TierFree:              1,
		TierSeasonalTraveller: 2,
		TierFrequentTraveller: 3,
	}

	return tierLevels[currentTier] >= tierLevels[requiredTier]
}

// Helper function to get subscription from context
func GetSubscriptionFromContext(c *gin.Context) (*UserSubscription, bool) {
	subscription, exists := c.Get("subscription")
	if !exists {
		return nil, false
	}
	sub, ok := subscription.(*UserSubscription)
	return sub, ok
}

// Helper function to check if feature is enabled for current user
func IsFeatureEnabled(c *gin.Context, featureFlagService *featureflags.Service, featureKey string) bool {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		return false
	}
	user := currentUser.(accounts.User)

	endpoint := c.Request.URL.Path
	method := c.Request.Method
	ctx := featureflags.EvaluationContext{
		UserID:      &user.BaseModel.ID,
		APIEndpoint: &endpoint,
		HTTPMethod:  &method,
	}

	isEnabled, err := featureFlagService.IsEnabled(featureKey, ctx)
	if err != nil {
		return false
	}

	return isEnabled
}

// GetFeatureValue gets a feature flag value for the current user
func GetFeatureValue(c *gin.Context, featureFlagService *featureflags.Service, featureKey string) (*string, error) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		return nil, nil
	}
	user := currentUser.(accounts.User)

	endpoint := c.Request.URL.Path
	method := c.Request.Method
	ctx := featureflags.EvaluationContext{
		UserID:      &user.BaseModel.ID,
		APIEndpoint: &endpoint,
		HTTPMethod:  &method,
	}

	return featureFlagService.GetValue(featureKey, ctx)
}
