package subscriptions

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

// SubscriptionTier represents the different subscription levels
type SubscriptionTier string

const (
	TierFree              SubscriptionTier = "free"
	TierSeasonalTraveller SubscriptionTier = "seasonal_traveller"
	TierFrequentTraveller SubscriptionTier = "frequent_traveller"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusExpired   SubscriptionStatus = "expired"
	StatusPaused    SubscriptionStatus = "paused"
	StatusTrialing  SubscriptionStatus = "trialing"
)

// SubscriptionPlan defines the available subscription plans
type SubscriptionPlan struct {
	core.BaseModel
	Tier                 SubscriptionTier  `json:"tier" gorm:"type:varchar(50);not null;unique" example:"free" description:"Tier identifier"`
	Name                 string            `json:"name" gorm:"not null" example:"Free Plan" description:"Display name of the plan"`
	Description          *string           `json:"description" example:"Basic features for occasional travelers" description:"Plan description"`
	PriceMonthly         *float64          `json:"price_monthly" example:"0.00" description:"Monthly price in USD"`
	PriceYearly          *float64          `json:"price_yearly" example:"0.00" description:"Yearly price in USD"`
	MaxTrips             *int              `json:"max_trips" example:"5" description:"Maximum number of active trips (null = unlimited)"`
	MaxTripDays          *int              `json:"max_trip_days" example:"10" description:"Maximum days per trip (null = unlimited)"`
	MaxTravellers        *int              `json:"max_travellers" example:"3" description:"Maximum travellers per trip (null = unlimited)"`
	MaxActivities        *int              `json:"max_activities" example:"20" description:"Maximum activities per day (null = unlimited)"`
	MaxDocuments         *int              `json:"max_documents" example:"10" description:"Maximum documents per trip (null = unlimited)"`
	StorageQuotaMB       *int              `json:"storage_quota_mb" example:"100" description:"Storage quota in MB (null = unlimited)"`
	IsActive             bool              `json:"is_active" gorm:"default:true" description:"Whether this plan is available for new subscriptions"`
	DisplayOrder         int               `json:"display_order" gorm:"default:0" description:"Order in which to display plans"`
	StripeProductID      *string           `json:"stripe_product_id" gorm:"unique" description:"Stripe product ID"`
	StripePriceMonthlyID *string           `json:"stripe_price_monthly_id" description:"Stripe monthly price ID"`
	StripePriceYearlyID  *string           `json:"stripe_price_yearly_id" description:"Stripe yearly price ID"`
	FeatureFlags         []PlanFeatureFlag `json:"feature_flags,omitempty" gorm:"foreignKey:PlanID" description:"Feature flags for this plan"`
}

// UserSubscription represents a user's subscription to a plan
type UserSubscription struct {
	core.BaseModel
	UserID               uuid.UUID          `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user"`
	PlanID               uuid.UUID          `json:"plan_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the subscription plan"`
	Plan                 *SubscriptionPlan  `json:"plan,omitempty" gorm:"foreignKey:PlanID" description:"The subscription plan details"`
	Status               SubscriptionStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'" example:"active" description:"Current status of subscription"`
	BillingCycle         *string            `json:"billing_cycle" gorm:"type:varchar(20)" example:"monthly" description:"Billing cycle (monthly, yearly)"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start" example:"2024-01-01T00:00:00Z" description:"Start of current billing period"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end" example:"2024-02-01T00:00:00Z" description:"End of current billing period"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end" gorm:"default:false" description:"Whether to cancel at period end"`
	CancelledAt          *time.Time         `json:"cancelled_at" description:"When the subscription was cancelled"`
	TrialStart           *time.Time         `json:"trial_start" description:"Start of trial period"`
	TrialEnd             *time.Time         `json:"trial_end" description:"End of trial period"`
	StripeSubscriptionID *string            `json:"stripe_subscription_id" gorm:"unique" description:"Stripe subscription ID"`
	StripeCustomerID     *string            `json:"stripe_customer_id" description:"Stripe customer ID"`
	LastPaymentDate      *time.Time         `json:"last_payment_date" description:"Date of last successful payment"`
	LastPaymentAmount    *float64           `json:"last_payment_amount" description:"Amount of last payment"`
	NextBillingDate      *time.Time         `json:"next_billing_date" description:"Next billing date"`
	AutoRenew            bool               `json:"auto_renew" gorm:"default:true" description:"Whether subscription auto-renews"`
	Notes                *string            `json:"notes" description:"Internal notes about the subscription"`
}

// PlanFeatureFlag maps subscription plans to feature flags
// This allows configuring which features are available for each tier
type PlanFeatureFlag struct {
	core.BaseModel
	PlanID         uuid.UUID `json:"plan_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the subscription plan"`
	FeatureFlagKey string    `json:"feature_flag_key" gorm:"not null" example:"advanced_analytics" description:"Key of the feature flag"`
	IsEnabled      bool      `json:"is_enabled" gorm:"default:true" description:"Whether this feature is enabled for this plan"`
	CustomValue    *string   `json:"custom_value" description:"Optional custom value for the feature (JSON)"`
}

// SubscriptionUsage tracks usage against subscription limits
type SubscriptionUsage struct {
	core.BaseModel
	SubscriptionID uuid.UUID `json:"subscription_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the subscription"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user"`
	ActiveTrips    int       `json:"active_trips" gorm:"default:0" description:"Current number of active trips"`
	TotalDocuments int       `json:"total_documents" gorm:"default:0" description:"Total number of documents"`
	StorageUsedMB  float64   `json:"storage_used_mb" gorm:"default:0" description:"Storage used in MB"`
	LastCalculated time.Time `json:"last_calculated" gorm:"not null" description:"When usage was last calculated"`
}

// SubscriptionHistory tracks changes to subscriptions over time
type SubscriptionHistory struct {
	core.BaseModel
	SubscriptionID uuid.UUID           `json:"subscription_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the subscription"`
	UserID         uuid.UUID           `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user"`
	FromPlanID     *uuid.UUID          `json:"from_plan_id" gorm:"type:uuid" description:"Previous plan ID (null if new subscription)"`
	ToPlanID       uuid.UUID           `json:"to_plan_id" gorm:"type:uuid;not null" description:"New plan ID"`
	FromStatus     *SubscriptionStatus `json:"from_status" gorm:"type:varchar(20)" description:"Previous status"`
	ToStatus       SubscriptionStatus  `json:"to_status" gorm:"type:varchar(20);not null" description:"New status"`
	ChangeReason   *string             `json:"change_reason" example:"user_upgrade" description:"Reason for the change"`
	ChangedBy      *uuid.UUID          `json:"changed_by" gorm:"type:uuid" description:"User who made the change (admin ID)"`
	EffectiveDate  time.Time           `json:"effective_date" gorm:"not null" description:"When the change became effective"`
	Notes          *string             `json:"notes" description:"Additional notes about the change"`
}

// GetValidTiers returns all valid subscription tiers
func GetValidTiers() []SubscriptionTier {
	return []SubscriptionTier{
		TierFree,
		TierSeasonalTraveller,
		TierFrequentTraveller,
	}
}

// IsValidTier checks if a tier is valid
func IsValidTier(tier string) bool {
	for _, validTier := range GetValidTiers() {
		if string(validTier) == tier {
			return true
		}
	}
	return false
}

// GetValidStatuses returns all valid subscription statuses
func GetValidStatuses() []SubscriptionStatus {
	return []SubscriptionStatus{
		StatusActive,
		StatusCancelled,
		StatusExpired,
		StatusPaused,
		StatusTrialing,
	}
}

// IsValidStatus checks if a status is valid
func IsValidStatus(status string) bool {
	for _, validStatus := range GetValidStatuses() {
		if string(validStatus) == status {
			return true
		}
	}
	return false
}

// GetModels returns all models for database migrations
func GetModels() []interface{} {
	return []interface{}{
		&SubscriptionPlan{},
		&UserSubscription{},
		&PlanFeatureFlag{},
		&SubscriptionUsage{},
		&SubscriptionHistory{},
	}
}
