# Subscription System Documentation

## Overview

The Trip Planner subscription system provides a flexible, scalable tier-based access control mechanism. It supports three subscription tiers and can easily be extended to support additional tiers in the future.

## Architecture

### Design Principles

1. **Scalability**: Adding new tiers requires only adding a new constant and creating a plan record
2. **Independence**: Subscription system works independently of feature flags
3. **Integration**: Seamlessly integrates with feature flags for fine-grained control
4. **Flexibility**: Supports various limit types and custom billing cycles

### Core Components

```
subscriptions/
├── models.go          # Data models
├── service.go         # Business logic
└── middleware.go      # HTTP middleware for access control
```

## Subscription Tiers

### 1. Free Tier
- **Price**: $0/month
- **Limits**:
  - Max 5 active trips
  - Max 10 days per trip
  - Max 3 travellers per trip
  - Max 20 activities per day
  - Max 10 documents per trip
  - 100 MB storage quota
- **Use Case**: Occasional travelers planning short trips

### 2. Seasonal Traveller
- **Price**: $9.99/month or $99.99/year
- **Limits**:
  - Max 20 active trips
  - Max 30 days per trip
  - Max 10 travellers per trip
  - Max 50 activities per day
  - Max 50 documents per trip
  - 1 GB storage quota
- **Use Case**: Regular travelers who plan multiple trips per year

### 3. Frequent Traveller
- **Price**: $24.99/month or $249.99/year
- **Limits**:
  - Unlimited active trips
  - Unlimited days per trip
  - Unlimited travellers per trip
  - Unlimited activities per day
  - Unlimited documents per trip
  - 10 GB storage quota
- **Use Case**: Travel enthusiasts and professionals

## Data Models

### SubscriptionPlan

Defines the available subscription plans with their limits and pricing.

```go
type SubscriptionPlan struct {
    Tier              SubscriptionTier  // free, seasonal_traveller, frequent_traveller
    Name              string             // Display name
    Description       *string            // Plan description
    PriceMonthly      *float64          // Monthly price
    PriceYearly       *float64          // Yearly price
    MaxTrips          *int              // null = unlimited
    MaxTripDays       *int              // null = unlimited
    MaxTravellers     *int              // null = unlimited
    MaxActivities     *int              // null = unlimited
    MaxDocuments      *int              // null = unlimited
    StorageQuotaMB    *int              // null = unlimited
    IsActive          bool              // Whether plan is available
    DisplayOrder      int               // Display order
}
```

### UserSubscription

Represents a user's active subscription.

```go
type UserSubscription struct {
    UserID                uuid.UUID
    PlanID                uuid.UUID
    Status                SubscriptionStatus  // active, cancelled, expired, paused, trialing
    BillingCycle          *string             // monthly, yearly
    CurrentPeriodStart    *time.Time
    CurrentPeriodEnd      *time.Time
    CancelAtPeriodEnd     bool
    StripeSubscriptionID  *string
    AutoRenew             bool
}
```

### SubscriptionUsage

Tracks current usage against limits.

```go
type SubscriptionUsage struct {
    SubscriptionID  uuid.UUID
    UserID          uuid.UUID
    ActiveTrips     int
    TotalDocuments  int
    StorageUsedMB   float64
    LastCalculated  time.Time
}
```

## Usage

### 1. Initialize Default Plans

Call during application startup:

```go
import "triplanner/subscriptions"

// In main.go or initialization code
err := subscriptions.InitializeDefaultPlans(db)
if err != nil {
    log.Fatal("Failed to initialize subscription plans:", err)
}
```

### 2. Get User Subscription

```go
subscriptionService := subscriptions.NewService(db)
subscription, err := subscriptionService.GetUserSubscription(userID)
if err != nil {
    // Handle error
}

fmt.Printf("User is on %s tier\n", subscription.Plan.Tier)
```

### 3. Check Limits

```go
// Check if user can create another trip
err := subscriptionService.CheckLimit(userID, "trips", currentTripCount)
if err == subscriptions.ErrLimitExceeded {
    // User has reached their limit
    // Prompt to upgrade
}

// Check storage limit
err = subscriptionService.CheckStorageLimit(userID, fileSize)
if err == subscriptions.ErrLimitExceeded {
    // Not enough storage
}
```

### 4. Update Usage

```go
// After creating a trip
err := subscriptionService.UpdateUsage(userID, map[string]interface{}{
    "active_trips": gorm.Expr("active_trips + 1"),
})

// After deleting a trip
err := subscriptionService.UpdateUsage(userID, map[string]interface{}{
    "active_trips": gorm.Expr("active_trips - 1"),
})

// After uploading a file
err := subscriptionService.UpdateUsage(userID, map[string]interface{}{
    "storage_used_mb": gorm.Expr("storage_used_mb + ?", fileSizeMB),
})
```

### 5. Using Middleware

```go
// Require minimum tier
router.POST("/advanced-feature",
    middleware.RequireSubscriptionTier(subscriptions.TierSeasonalTraveller),
    handlers.AdvancedFeature,
)

// Check limit before allowing action
router.POST("/trip-plans",
    middleware.CheckLimit("trips", func(c *gin.Context) int {
        // Return current trip count for user
        var count int64
        db.Model(&trips.TripPlan{}).Where("user_id = ?", userID).Count(&count)
        return int(count)
    }),
    handlers.CreateTripPlan,
)
```

### 6. Upgrade Subscription

```go
err := subscriptionService.UpgradeSubscription(
    userID,
    newPlanID,
    "user_initiated_upgrade",
)
```

### 7. Cancel Subscription

```go
// Cancel immediately
err := subscriptionService.CancelSubscription(userID, false)

// Cancel at period end
err := subscriptionService.CancelSubscription(userID, true)
```

## Adding New Tiers

To add a new subscription tier:

1. **Add Tier Constant** in `models.go`:

```go
const (
    TierFree              SubscriptionTier = "free"
    TierSeasonalTraveller SubscriptionTier = "seasonal_traveller"
    TierFrequentTraveller SubscriptionTier = "frequent_traveller"
    TierEnterprise        SubscriptionTier = "enterprise"  // NEW
)
```

2. **Update Validation Functions** in `models.go`:

```go
func GetValidTiers() []SubscriptionTier {
    return []SubscriptionTier{
        TierFree,
        TierSeasonalTraveller,
        TierFrequentTraveller,
        TierEnterprise,  // NEW
    }
}
```

3. **Add Plan Definition** in `service.go`:

```go
{
    Tier:          TierEnterprise,
    Name:          "Enterprise",
    Description:   core.StringPtr("For organizations with advanced needs"),
    PriceMonthly:  core.Float64Ptr(99.99),
    PriceYearly:   core.Float64Ptr(999.99),
    MaxTrips:      nil, // Unlimited
    MaxTripDays:   nil,
    MaxTravellers: nil,
    MaxActivities: nil,
    MaxDocuments:  nil,
    StorageQuotaMB: core.IntPtr(102400), // 100GB
    IsActive:      true,
    DisplayOrder:  4,
}
```

4. **Update Tier Hierarchy** in `middleware.go`:

```go
tierLevels := map[SubscriptionTier]int{
    TierFree:              1,
    TierSeasonalTraveller: 2,
    TierFrequentTraveller: 3,
    TierEnterprise:        4,  // NEW
}
```

## Integration with Feature Flags

Subscriptions and feature flags work together through the `PlanFeatureFlag` model:

```go
// Enable a feature for a specific tier
planFeatureFlag := PlanFeatureFlag{
    PlanID:         seasonalPlanID,
    FeatureFlagKey: "advanced_analytics",
    IsEnabled:      true,
}
db.Create(&planFeatureFlag)
```

See [FeatureFlag.md](./FeatureFlag.md) for more details on integration.

## Best Practices

### 1. Always Check Limits Before Actions

```go
// GOOD
err := subscriptionService.CheckLimit(userID, "trips", currentCount)
if err == subscriptions.ErrLimitExceeded {
    return gin.H{"error": "Limit exceeded", "upgrade_url": "/upgrade"}
}
// Proceed with action

// BAD - Creating resource without checking
db.Create(&trip)
```

### 2. Update Usage Consistently

```go
// Use transactions to ensure consistency
tx := db.Begin()
if err := tx.Create(&trip).Error; err != nil {
    tx.Rollback()
    return err
}
if err := subscriptionService.UpdateUsage(userID, map[string]interface{}{
    "active_trips": gorm.Expr("active_trips + 1"),
}); err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

### 3. Handle Upgrades Gracefully

```go
// Don't immediately apply new limits - let existing resources stay
// Only enforce new limits for new resources
```

### 4. Use Middleware for Consistent Enforcement

```go
// Apply middleware at router level for consistency
apiV1 := router.Group("/api/v1")
apiV1.Use(middleware.InjectSubscriptionContext())
```

## Monitoring and Analytics

### Track Usage Trends

```sql
SELECT
    plan.tier,
    AVG(usage.active_trips) as avg_trips,
    AVG(usage.storage_used_mb) as avg_storage
FROM subscription_usages usage
JOIN user_subscriptions sub ON usage.subscription_id = sub.id
JOIN subscription_plans plan ON sub.plan_id = plan.id
GROUP BY plan.tier;
```

### Monitor Limit Hits

Track when users hit limits to identify upgrade opportunities:

```go
// In your limit check logic, log when limits are hit
if err == subscriptions.ErrLimitExceeded {
    analytics.Track("subscription_limit_hit", map[string]interface{}{
        "user_id":    userID,
        "limit_type": limitType,
        "tier":       subscription.Plan.Tier,
    })
}
```

## Billing Integration

The subscription system is designed to integrate with Stripe:

- `StripeProductID`: Maps to Stripe Product
- `StripePriceMonthlyID`: Maps to Stripe Price for monthly billing
- `StripePriceYearlyID`: Maps to Stripe Price for yearly billing
- `StripeSubscriptionID`: Maps to Stripe Subscription
- `StripeCustomerID`: Maps to Stripe Customer

See Stripe documentation for webhook handling and payment processing.

## Database Migrations

The subscription models will be automatically migrated if using the application's migration system. Ensure you include subscription models:

```go
import "triplanner/subscriptions"

// In migration code
models := append(coreModels, subscriptions.GetModels()...)
db.AutoMigrate(models...)
```
