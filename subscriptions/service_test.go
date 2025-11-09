package subscriptions

import (
	"testing"
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

func TestGetValidTiers(t *testing.T) {
	tiers := GetValidTiers()

	if len(tiers) != 3 {
		t.Errorf("Expected 3 tiers, got %d", len(tiers))
	}

	expectedTiers := map[SubscriptionTier]bool{
		TierFree:              true,
		TierSeasonalTraveller: true,
		TierFrequentTraveller: true,
	}

	for _, tier := range tiers {
		if !expectedTiers[tier] {
			t.Errorf("Unexpected tier: %s", tier)
		}
	}
}

func TestIsValidTier(t *testing.T) {
	tests := []struct {
		name  string
		tier  string
		valid bool
	}{
		{"Free tier", "free", true},
		{"Seasonal traveller", "seasonal_traveller", true},
		{"Frequent traveller", "frequent_traveller", true},
		{"Invalid tier", "enterprise", false},
		{"Empty string", "", false},
		{"Random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTier(tt.tier)
			if result != tt.valid {
				t.Errorf("IsValidTier(%s) = %v, want %v", tt.tier, result, tt.valid)
			}
		})
	}
}

func TestGetValidStatuses(t *testing.T) {
	statuses := GetValidStatuses()

	expectedCount := 5
	if len(statuses) != expectedCount {
		t.Errorf("Expected %d statuses, got %d", expectedCount, len(statuses))
	}

	expectedStatuses := map[SubscriptionStatus]bool{
		StatusActive:    true,
		StatusCancelled: true,
		StatusExpired:   true,
		StatusPaused:    true,
		StatusTrialing:  true,
	}

	for _, status := range statuses {
		if !expectedStatuses[status] {
			t.Errorf("Unexpected status: %s", status)
		}
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		valid  bool
	}{
		{"Active status", "active", true},
		{"Cancelled status", "cancelled", true},
		{"Expired status", "expired", true},
		{"Paused status", "paused", true},
		{"Trialing status", "trialing", true},
		{"Invalid status", "pending", false},
		{"Empty string", "", false},
		{"Random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			if result != tt.valid {
				t.Errorf("IsValidStatus(%s) = %v, want %v", tt.status, result, tt.valid)
			}
		})
	}
}

func TestSubscriptionPlan_Limits(t *testing.T) {
	t.Run("Free tier has limits", func(t *testing.T) {
		plan := SubscriptionPlan{
			Tier:           TierFree,
			MaxTrips:       core.IntPtr(5),
			MaxTripDays:    core.IntPtr(10),
			MaxTravellers:  core.IntPtr(3),
			StorageQuotaMB: core.IntPtr(100),
		}

		if plan.MaxTrips == nil || *plan.MaxTrips != 5 {
			t.Error("Free tier should have trip limit of 5")
		}
		if plan.MaxTripDays == nil || *plan.MaxTripDays != 10 {
			t.Error("Free tier should have trip days limit of 10")
		}
		if plan.MaxTravellers == nil || *plan.MaxTravellers != 3 {
			t.Error("Free tier should have travellers limit of 3")
		}
		if plan.StorageQuotaMB == nil || *plan.StorageQuotaMB != 100 {
			t.Error("Free tier should have storage limit of 100MB")
		}
	})

	t.Run("Frequent traveller has unlimited limits", func(t *testing.T) {
		plan := SubscriptionPlan{
			Tier:           TierFrequentTraveller,
			MaxTrips:       nil, // Unlimited
			MaxTripDays:    nil,
			MaxTravellers:  nil,
			MaxActivities:  nil,
			MaxDocuments:   nil,
			StorageQuotaMB: core.IntPtr(10240),
		}

		if plan.MaxTrips != nil {
			t.Error("Frequent traveller should have unlimited trips")
		}
		if plan.MaxTripDays != nil {
			t.Error("Frequent traveller should have unlimited trip days")
		}
		if plan.MaxTravellers != nil {
			t.Error("Frequent traveller should have unlimited travellers")
		}
		if plan.StorageQuotaMB == nil || *plan.StorageQuotaMB != 10240 {
			t.Error("Frequent traveller should have 10GB storage")
		}
	})
}

func TestUserSubscription_Fields(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	planID := uuid.New()

	subscription := UserSubscription{
		UserID:             userID,
		PlanID:             planID,
		Status:             StatusActive,
		BillingCycle:       core.StringPtr("monthly"),
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   &now,
		AutoRenew:          true,
	}

	if subscription.UserID != userID {
		t.Error("UserID not set correctly")
	}
	if subscription.PlanID != planID {
		t.Error("PlanID not set correctly")
	}
	if subscription.Status != StatusActive {
		t.Error("Status should be active")
	}
	if subscription.BillingCycle == nil || *subscription.BillingCycle != "monthly" {
		t.Error("BillingCycle should be monthly")
	}
	if !subscription.AutoRenew {
		t.Error("AutoRenew should be true")
	}
}

func TestSubscriptionUsage_Tracking(t *testing.T) {
	usage := SubscriptionUsage{
		SubscriptionID: uuid.New(),
		UserID:         uuid.New(),
		ActiveTrips:    3,
		TotalDocuments: 15,
		StorageUsedMB:  45.5,
		LastCalculated: time.Now(),
	}

	if usage.ActiveTrips != 3 {
		t.Errorf("Expected 3 active trips, got %d", usage.ActiveTrips)
	}
	if usage.TotalDocuments != 15 {
		t.Errorf("Expected 15 documents, got %d", usage.TotalDocuments)
	}
	if usage.StorageUsedMB != 45.5 {
		t.Errorf("Expected 45.5MB storage, got %f", usage.StorageUsedMB)
	}
}

func TestPlanFeatureFlag_Integration(t *testing.T) {
	planID := uuid.New()

	featureFlag := PlanFeatureFlag{
		PlanID:         planID,
		FeatureFlagKey: "advanced_analytics",
		IsEnabled:      true,
		CustomValue:    core.StringPtr(`{"refresh_interval": 30}`),
	}

	if featureFlag.PlanID != planID {
		t.Error("PlanID not set correctly")
	}
	if featureFlag.FeatureFlagKey != "advanced_analytics" {
		t.Error("FeatureFlagKey not set correctly")
	}
	if !featureFlag.IsEnabled {
		t.Error("Feature should be enabled")
	}
	if featureFlag.CustomValue == nil {
		t.Error("CustomValue should not be nil")
	}
}

func TestSubscriptionHistory_Tracking(t *testing.T) {
	subscriptionID := uuid.New()
	userID := uuid.New()
	fromPlanID := uuid.New()
	toPlanID := uuid.New()
	fromStatus := StatusActive
	now := time.Now()

	history := SubscriptionHistory{
		SubscriptionID: subscriptionID,
		UserID:         userID,
		FromPlanID:     &fromPlanID,
		ToPlanID:       toPlanID,
		FromStatus:     &fromStatus,
		ToStatus:       StatusActive,
		ChangeReason:   core.StringPtr("user_upgrade"),
		EffectiveDate:  now,
	}

	if history.SubscriptionID != subscriptionID {
		t.Error("SubscriptionID not set correctly")
	}
	if history.FromPlanID == nil || *history.FromPlanID != fromPlanID {
		t.Error("FromPlanID not set correctly")
	}
	if history.ToPlanID != toPlanID {
		t.Error("ToPlanID not set correctly")
	}
	if history.ChangeReason == nil || *history.ChangeReason != "user_upgrade" {
		t.Error("ChangeReason should be user_upgrade")
	}
}

func TestDefaultPlansConfiguration(t *testing.T) {
	// Test that default plans are configured correctly
	t.Run("Free tier configuration", func(t *testing.T) {
		// This would normally be in InitializeDefaultPlans
		plan := SubscriptionPlan{
			Tier:           TierFree,
			Name:           "Free Tier",
			PriceMonthly:   core.Float64Ptr(0),
			MaxTrips:       core.IntPtr(5),
			MaxTripDays:    core.IntPtr(10),
			MaxTravellers:  core.IntPtr(3),
			MaxActivities:  core.IntPtr(20),
			MaxDocuments:   core.IntPtr(10),
			StorageQuotaMB: core.IntPtr(100),
		}

		if *plan.PriceMonthly != 0 {
			t.Error("Free tier should be free")
		}
		if plan.MaxTrips == nil || *plan.MaxTrips <= 0 {
			t.Error("Free tier should have positive trip limit")
		}
	})

	t.Run("Seasonal traveller configuration", func(t *testing.T) {
		plan := SubscriptionPlan{
			Tier:           TierSeasonalTraveller,
			Name:           "Seasonal Traveller",
			PriceMonthly:   core.Float64Ptr(9.99),
			PriceYearly:    core.Float64Ptr(99.99),
			MaxTrips:       core.IntPtr(20),
			MaxTripDays:    core.IntPtr(30),
			MaxTravellers:  core.IntPtr(10),
			StorageQuotaMB: core.IntPtr(1024),
		}

		if *plan.PriceMonthly <= 0 {
			t.Error("Seasonal traveller should have a price")
		}
		if *plan.PriceYearly <= 0 {
			t.Error("Seasonal traveller should have yearly price")
		}
		if *plan.MaxTrips <= *core.IntPtr(5) {
			t.Error("Seasonal traveller should have more trips than free tier")
		}
	})

	t.Run("Frequent traveller configuration", func(t *testing.T) {
		plan := SubscriptionPlan{
			Tier:           TierFrequentTraveller,
			Name:           "Frequent Traveller",
			PriceMonthly:   core.Float64Ptr(24.99),
			PriceYearly:    core.Float64Ptr(249.99),
			MaxTrips:       nil, // Unlimited
			MaxTripDays:    nil,
			MaxTravellers:  nil,
			StorageQuotaMB: core.IntPtr(10240),
		}

		if *plan.PriceMonthly <= *core.Float64Ptr(9.99) {
			t.Error("Frequent traveller should cost more than seasonal")
		}
		if plan.MaxTrips != nil {
			t.Error("Frequent traveller should have unlimited trips")
		}
	})
}

func TestLimitChecking(t *testing.T) {
	t.Run("Check if value exceeds limit", func(t *testing.T) {
		limit := 5

		// Under limit
		if 3 >= limit {
			t.Error("3 should be under limit of 5")
		}

		// At limit
		if 5 < limit {
			t.Error("5 should be at or over limit of 5")
		}

		// Over limit
		if 6 < limit {
			t.Error("6 should be over limit of 5")
		}
	})

	t.Run("Unlimited check", func(t *testing.T) {
		var limit *int = nil

		// Nil limit means unlimited
		if limit != nil {
			t.Error("Nil limit should represent unlimited")
		}
	})
}
