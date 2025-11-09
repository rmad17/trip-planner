package subscriptions

import (
	"errors"
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrLimitExceeded        = errors.New("subscription limit exceeded")
	ErrInvalidPlan          = errors.New("invalid subscription plan")
)

// Service provides subscription-related business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new subscription service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetUserSubscription retrieves the active subscription for a user
func (s *Service) GetUserSubscription(userID uuid.UUID) (*UserSubscription, error) {
	var subscription UserSubscription
	err := s.db.Preload("Plan").
		Where("user_id = ? AND status = ?", userID, StatusActive).
		Order("created_at DESC").
		First(&subscription).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default free subscription
			return s.GetOrCreateFreeSubscription(userID)
		}
		return nil, err
	}

	return &subscription, nil
}

// GetOrCreateFreeSubscription gets or creates a free subscription for a user
func (s *Service) GetOrCreateFreeSubscription(userID uuid.UUID) (*UserSubscription, error) {
	// Get the free plan
	var freePlan SubscriptionPlan
	err := s.db.Where("tier = ?", TierFree).First(&freePlan).Error
	if err != nil {
		return nil, ErrInvalidPlan
	}

	// Create a free subscription
	subscription := UserSubscription{
		UserID: userID,
		PlanID: freePlan.ID,
		Plan:   &freePlan,
		Status: StatusActive,
	}

	err = s.db.Create(&subscription).Error
	if err != nil {
		return nil, err
	}

	// Create usage record
	usage := SubscriptionUsage{
		SubscriptionID: subscription.ID,
		UserID:         userID,
		LastCalculated: time.Now(),
	}
	s.db.Create(&usage)

	return &subscription, nil
}

// CheckLimit checks if a user can perform an action based on their subscription limits
func (s *Service) CheckLimit(userID uuid.UUID, limitType string, currentValue int) error {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	if subscription.Plan == nil {
		return ErrInvalidPlan
	}

	var limit *int

	switch limitType {
	case "trips":
		limit = subscription.Plan.MaxTrips
	case "trip_days":
		limit = subscription.Plan.MaxTripDays
	case "travellers":
		limit = subscription.Plan.MaxTravellers
	case "activities":
		limit = subscription.Plan.MaxActivities
	case "documents":
		limit = subscription.Plan.MaxDocuments
	default:
		return errors.New("invalid limit type")
	}

	// If limit is nil, it means unlimited
	if limit == nil {
		return nil
	}

	if currentValue >= *limit {
		return ErrLimitExceeded
	}

	return nil
}

// CheckStorageLimit checks if a user can upload more files based on their storage quota
func (s *Service) CheckStorageLimit(userID uuid.UUID, additionalMB float64) error {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	if subscription.Plan == nil {
		return ErrInvalidPlan
	}

	// If storage quota is nil, it means unlimited
	if subscription.Plan.StorageQuotaMB == nil {
		return nil
	}

	// Get current usage
	var usage SubscriptionUsage
	err = s.db.Where("subscription_id = ?", subscription.ID).First(&usage).Error
	if err != nil {
		return err
	}

	if usage.StorageUsedMB+additionalMB > float64(*subscription.Plan.StorageQuotaMB) {
		return ErrLimitExceeded
	}

	return nil
}

// UpdateUsage updates the usage statistics for a subscription
func (s *Service) UpdateUsage(userID uuid.UUID, updates map[string]interface{}) error {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	updates["last_calculated"] = time.Now()

	return s.db.Model(&SubscriptionUsage{}).
		Where("subscription_id = ?", subscription.ID).
		Updates(updates).Error
}

// GetUsage retrieves the current usage for a user
func (s *Service) GetUsage(userID uuid.UUID) (*SubscriptionUsage, error) {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return nil, err
	}

	var usage SubscriptionUsage
	err = s.db.Where("subscription_id = ?", subscription.ID).First(&usage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create initial usage record
			usage = SubscriptionUsage{
				SubscriptionID: subscription.ID,
				UserID:         userID,
				LastCalculated: time.Now(),
			}
			s.db.Create(&usage)
			return &usage, nil
		}
		return nil, err
	}

	return &usage, nil
}

// UpgradeSubscription upgrades a user's subscription to a new plan
func (s *Service) UpgradeSubscription(userID uuid.UUID, newPlanID uuid.UUID, reason string) error {
	currentSub, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	var newPlan SubscriptionPlan
	err = s.db.First(&newPlan, newPlanID).Error
	if err != nil {
		return ErrInvalidPlan
	}

	// Create history record
	history := SubscriptionHistory{
		SubscriptionID: currentSub.ID,
		UserID:         userID,
		FromPlanID:     &currentSub.PlanID,
		ToPlanID:       newPlanID,
		FromStatus:     &currentSub.Status,
		ToStatus:       StatusActive,
		ChangeReason:   &reason,
		EffectiveDate:  time.Now(),
	}
	s.db.Create(&history)

	// Update subscription
	return s.db.Model(currentSub).Updates(map[string]interface{}{
		"plan_id": newPlanID,
		"status":  StatusActive,
	}).Error
}

// CancelSubscription cancels a user's subscription
func (s *Service) CancelSubscription(userID uuid.UUID, cancelAtPeriodEnd bool) error {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"cancel_at_period_end": cancelAtPeriodEnd,
	}

	if !cancelAtPeriodEnd {
		updates["status"] = StatusCancelled
		updates["cancelled_at"] = time.Now()

		// Create history record
		history := SubscriptionHistory{
			SubscriptionID: subscription.ID,
			UserID:         userID,
			FromPlanID:     &subscription.PlanID,
			ToPlanID:       subscription.PlanID,
			FromStatus:     &subscription.Status,
			ToStatus:       StatusCancelled,
			EffectiveDate:  time.Now(),
		}
		s.db.Create(&history)
	}

	return s.db.Model(subscription).Updates(updates).Error
}

// GetPlanByTier retrieves a subscription plan by its tier
func (s *Service) GetPlanByTier(tier SubscriptionTier) (*SubscriptionPlan, error) {
	var plan SubscriptionPlan
	err := s.db.Where("tier = ? AND is_active = ?", tier, true).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetAllPlans retrieves all active subscription plans
func (s *Service) GetAllPlans() ([]SubscriptionPlan, error) {
	var plans []SubscriptionPlan
	err := s.db.Where("is_active = ?", true).Order("display_order ASC").Find(&plans).Error
	return plans, err
}

// InitializeDefaultPlans creates the default subscription plans if they don't exist
func InitializeDefaultPlans(db *gorm.DB) error {
	plans := []SubscriptionPlan{
		{
			Tier:           TierFree,
			Name:           "Free Tier",
			Description:    core.StringPtr("Perfect for occasional travelers planning short trips"),
			PriceMonthly:   core.Float64Ptr(0),
			PriceYearly:    core.Float64Ptr(0),
			MaxTrips:       core.IntPtr(5),
			MaxTripDays:    core.IntPtr(10),
			MaxTravellers:  core.IntPtr(3),
			MaxActivities:  core.IntPtr(20),
			MaxDocuments:   core.IntPtr(10),
			StorageQuotaMB: core.IntPtr(100),
			IsActive:       true,
			DisplayOrder:   1,
		},
		{
			Tier:           TierSeasonalTraveller,
			Name:           "Seasonal Traveller",
			Description:    core.StringPtr("Ideal for regular travelers who plan multiple trips per year"),
			PriceMonthly:   core.Float64Ptr(9.99),
			PriceYearly:    core.Float64Ptr(99.99),
			MaxTrips:       core.IntPtr(20),
			MaxTripDays:    core.IntPtr(30),
			MaxTravellers:  core.IntPtr(10),
			MaxActivities:  core.IntPtr(50),
			MaxDocuments:   core.IntPtr(50),
			StorageQuotaMB: core.IntPtr(1024),
			IsActive:       true,
			DisplayOrder:   2,
		},
		{
			Tier:           TierFrequentTraveller,
			Name:           "Frequent Traveller",
			Description:    core.StringPtr("For travel enthusiasts and professionals who travel extensively"),
			PriceMonthly:   core.Float64Ptr(24.99),
			PriceYearly:    core.Float64Ptr(249.99),
			MaxTrips:       nil, // Unlimited
			MaxTripDays:    nil, // Unlimited
			MaxTravellers:  nil, // Unlimited
			MaxActivities:  nil, // Unlimited
			MaxDocuments:   nil, // Unlimited
			StorageQuotaMB: core.IntPtr(10240),
			IsActive:       true,
			DisplayOrder:   3,
		},
	}

	for _, plan := range plans {
		var existing SubscriptionPlan
		err := db.Where("tier = ?", plan.Tier).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new plan
			if err := db.Create(&plan).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
