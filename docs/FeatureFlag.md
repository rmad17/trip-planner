# Feature Flag System Documentation

## Overview

The Trip Planner feature flag system provides fine-grained control over feature availability and behavior. It supports multiple scopes (global, user, subscription, API), various flag types, and percentage-based rollouts.

## Architecture

### Design Principles

1. **Independence**: Feature flags work independently of subscriptions
2. **Integration**: Seamlessly integrates with subscriptions for tier-based features
3. **Flexibility**: Supports multiple flag types and scopes
4. **Scalability**: Can handle thousands of flags with efficient evaluation
5. **Auditability**: All changes and evaluations are tracked

### Core Components

```
featureflags/
├── models.go          # Data models
└── service.go         # Evaluation logic and business rules
```

## Feature Flag Types

### 1. Boolean (On/Off)
Simple true/false flags for enabling/disabling features.

```go
Type: TypeBoolean
DefaultValue: "false"
```

### 2. String
String-based configuration values.

```go
Type: TypeString
DefaultValue: "default_theme"
```

### 3. Number
Numeric configuration values.

```go
Type: TypeNumber
DefaultValue: "10"
```

### 4. JSON
Complex configuration objects.

```go
Type: TypeJSON
DefaultValue: `{"max_retries": 3, "timeout": 30}`
```

### 5. Percentage
Gradual rollout to percentage of users.

```go
Type: TypePercentage
RolloutPercentage: 25  // 25% of users
```

## Feature Flag Scopes

### 1. Global
Applies to all users and API endpoints.

```go
Scope: ScopeGlobal
```

### 2. User
Can be overridden per user.

```go
Scope: ScopeUser
```

### 3. Subscription
Tied to subscription tiers (integrates with subscription system).

```go
Scope: ScopeSubscription
```

### 4. API
Applies to specific API endpoints.

```go
Scope: ScopeAPI
```

## Data Models

### FeatureFlag

Core feature flag definition.

```go
type FeatureFlag struct {
    Key               string            // Unique identifier
    Name              string            // Display name
    Description       *string           // What this flag controls
    Type              FeatureFlagType   // boolean, string, number, json, percentage
    Scope             FeatureFlagScope  // global, user, subscription, api
    Status            FeatureFlagStatus // draft, active, deprecated, archived
    DefaultValue      *string           // Default value
    IsEnabled         bool              // Master switch
    RolloutPercentage int               // 0-100 for gradual rollout
    Tags              []string          // For organization
    OwnerTeam         *string           // Responsible team
    Dependencies      []string          // Other flags this depends on
    ExpiresAt         *time.Time        // Auto-disable date
}
```

### UserFeatureFlag

User-specific overrides.

```go
type UserFeatureFlag struct {
    UserID          uuid.UUID
    FeatureFlagKey  string
    IsEnabled       bool
    CustomValue     *string
    Reason          *string
    ExpiresAt       *time.Time
    SetBy           *uuid.UUID  // Admin who set this
}
```

### APIFeatureFlag

API endpoint-specific overrides.

```go
type APIFeatureFlag struct {
    APIEndpoint     string
    HTTPMethod      *string  // null = all methods
    FeatureFlagKey  string
    IsEnabled       bool
    CustomValue     *string
    Priority        int      // For overlapping rules
    ExpiresAt       *time.Time
}
```

## Usage

### 1. Create a Feature Flag

```go
flagService := featureflags.NewService(db)

flag := &featureflags.FeatureFlag{
    Key:          "advanced_analytics",
    Name:         "Advanced Analytics",
    Description:  core.StringPtr("Enable advanced analytics dashboard"),
    Type:         featureflags.TypeBoolean,
    Scope:        featureflags.ScopeSubscription,
    Status:       featureflags.StatusActive,
    DefaultValue: core.StringPtr("false"),
    IsEnabled:    true,
    OwnerTeam:    core.StringPtr("analytics"),
}

err := flagService.CreateFlag(flag, &adminUserID)
```

### 2. Check if Feature is Enabled

```go
// Create evaluation context
ctx := featureflags.EvaluationContext{
    UserID:      &userID,
    APIEndpoint: core.StringPtr("/api/v1/analytics"),
    HTTPMethod:  core.StringPtr("GET"),
}

// Check if enabled
isEnabled, err := flagService.IsEnabled("advanced_analytics", ctx)
if isEnabled {
    // Show advanced analytics
}
```

### 3. Get Feature Value

```go
// For non-boolean flags
value, err := flagService.GetStringValue("theme_name", ctx, "default")
fmt.Printf("Using theme: %s\n", value)

// For numeric flags
maxRetries, err := flagService.GetIntValue("max_retries", ctx, 3)
```

### 4. Set User Override

```go
// Enable feature for specific user (e.g., beta tester)
err := flagService.SetUserOverride(
    userID,
    "advanced_analytics",
    true,                              // enabled
    nil,                               // no custom value
    core.StringPtr("beta_tester"),     // reason
    &adminUserID,                      // who set it
)
```

### 5. Set API Override

```go
// Enable rate limiting for specific endpoint
err := flagService.SetAPIOverride(
    "/api/v1/trip-plans",              // endpoint
    core.StringPtr("POST"),            // method
    "rate_limiting",                   // flag key
    true,                              // enabled
    core.StringPtr(`{"limit": 100}`),  // custom config
    10,                                // priority
)
```

### 6. Gradual Rollout

```go
// Create flag with 25% rollout
flag := &featureflags.FeatureFlag{
    Key:               "new_ui",
    Name:              "New UI Design",
    Type:              featureflags.TypeBoolean,
    Scope:             featureflags.ScopeUser,
    Status:            featureflags.StatusActive,
    IsEnabled:         true,
    RolloutPercentage: 25,  // 25% of users
}

flagService.CreateFlag(flag, &adminUserID)

// Gradually increase rollout
flagService.UpdateFlag("new_ui", map[string]interface{}{
    "rollout_percentage": 50,  // Now 50%
}, &adminUserID)
```

### 7. Using Middleware

```go
import "triplanner/subscriptions"

// Require feature to access endpoint
router.GET("/advanced-analytics",
    authMiddleware,
    subscriptionMiddleware.RequireFeature("advanced_analytics"),
    handlers.AdvancedAnalytics,
)

// Check feature in handler
func MyHandler(c *gin.Context) {
    if subscriptions.IsFeatureEnabled(c, flagService, "export_pdf") {
        // Enable PDF export
    }
}
```

## Integration with Subscription System

Feature flags and subscriptions work together through the `PlanFeatureFlag` model in the subscription system.

### 1. Enable Feature for Subscription Tier

```go
// In subscription system
planFeatureFlag := subscriptions.PlanFeatureFlag{
    PlanID:         seasonalTravellerPlanID,
    FeatureFlagKey: "advanced_analytics",
    IsEnabled:      true,
}
db.Create(&planFeatureFlag)
```

### 2. Evaluation Priority

When evaluating a feature flag, the system checks in this order:

1. **API Override** (highest priority)
2. **User Override**
3. **Subscription Plan Feature** (via PlanFeatureFlag)
4. **Rollout Percentage**
5. **Default Value** (lowest priority)

### 3. Example: Tier-Based Feature

```go
// 1. Create the feature flag
flag := &featureflags.FeatureFlag{
    Key:          "trip_collaboration",
    Name:         "Trip Collaboration",
    Scope:        featureflags.ScopeSubscription,
    Status:       featureflags.StatusActive,
    DefaultValue: core.StringPtr("false"),
    IsEnabled:    true,
}
flagService.CreateFlag(flag, nil)

// 2. Enable for Seasonal and Frequent tiers
seasonalPlan, _ := subscriptionService.GetPlanByTier(subscriptions.TierSeasonalTraveller)
frequentPlan, _ := subscriptionService.GetPlanByTier(subscriptions.TierFrequentTraveller)

db.Create(&subscriptions.PlanFeatureFlag{
    PlanID:         seasonalPlan.ID,
    FeatureFlagKey: "trip_collaboration",
    IsEnabled:      true,
})

db.Create(&subscriptions.PlanFeatureFlag{
    PlanID:         frequentPlan.ID,
    FeatureFlagKey: "trip_collaboration",
    IsEnabled:      true,
})

// 3. Now users with Seasonal or Frequent tier automatically get this feature
```

## Independent Use Cases

Feature flags can be used independently of subscriptions:

### 1. A/B Testing

```go
// Test new algorithm with 50% of users
flag := &featureflags.FeatureFlag{
    Key:               "new_recommendation_algorithm",
    Name:              "New Recommendation Algorithm",
    Type:              featureflags.TypeBoolean,
    Scope:             featureflags.ScopeUser,
    Status:            featureflags.StatusActive,
    IsEnabled:         true,
    RolloutPercentage: 50,
}
```

### 2. Emergency Kill Switch

```go
// Disable problematic feature immediately
flagService.UpdateFlag("problematic_feature", map[string]interface{}{
    "is_enabled": false,
}, &adminUserID)
```

### 3. Beta Features

```go
// Enable for specific beta testers
for _, betaUserID := range betaTesters {
    flagService.SetUserOverride(
        betaUserID,
        "beta_feature",
        true,
        nil,
        core.StringPtr("beta_tester"),
        nil,
    )
}
```

### 4. API Rate Limiting

```go
// Different rate limits per endpoint
flagService.SetAPIOverride(
    "/api/v1/ai/recommendations",
    nil,  // all methods
    "rate_limit",
    true,
    core.StringPtr(`{"requests_per_minute": 10}`),
    0,
)
```

### 5. Configuration Management

```go
// Dynamic configuration without redeployment
flag := &featureflags.FeatureFlag{
    Key:          "ai_model_config",
    Name:         "AI Model Configuration",
    Type:         featureflags.TypeJSON,
    DefaultValue: core.StringPtr(`{
        "model": "gpt-4",
        "temperature": 0.7,
        "max_tokens": 2000
    }`),
    IsEnabled:    true,
}
```

## Evaluation Analytics

The system tracks all evaluations for analytics:

```go
// Query evaluation statistics
type EvaluationStats struct {
    FeatureFlagKey string
    EnabledCount   int
    DisabledCount  int
    UniqueUsers    int
}

var stats []EvaluationStats
db.Raw(`
    SELECT
        feature_flag_key,
        COUNT(CASE WHEN was_enabled THEN 1 END) as enabled_count,
        COUNT(CASE WHEN NOT was_enabled THEN 1 END) as disabled_count,
        COUNT(DISTINCT user_id) as unique_users
    FROM feature_flag_evaluations
    WHERE evaluated_at > NOW() - INTERVAL '7 days'
    GROUP BY feature_flag_key
`).Scan(&stats)
```

## Flag Lifecycle Management

### 1. Draft → Active

```go
// Create in draft mode for testing
flag := &featureflags.FeatureFlag{
    Key:    "new_feature",
    Status: featureflags.StatusDraft,
    IsEnabled: false,
}
flagService.CreateFlag(flag, &adminUserID)

// Test with specific users
flagService.SetUserOverride(testUserID, "new_feature", true, nil, nil, nil)

// Activate when ready
flagService.UpdateFlag("new_feature", map[string]interface{}{
    "status":     featureflags.StatusActive,
    "is_enabled": true,
}, &adminUserID)
```

### 2. Active → Deprecated

```go
// Mark for removal
now := time.Now()
flagService.UpdateFlag("old_feature", map[string]interface{}{
    "status":        featureflags.StatusDeprecated,
    "deprecated_at": &now,
}, &adminUserID)
```

### 3. Deprecated → Archived

```go
// After removing from code
flagService.UpdateFlag("old_feature", map[string]interface{}{
    "status":     featureflags.StatusArchived,
    "is_enabled": false,
}, &adminUserID)
```

## Best Practices

### 1. Naming Conventions

```go
// GOOD
"advanced_analytics"
"trip_collaboration"
"ai_recommendations"
"export_pdf"

// BAD
"feature1"
"new_feature"
"test"
```

### 2. Use Appropriate Scopes

```go
// User-facing features
Scope: ScopeUser

// Subscription-tier features
Scope: ScopeSubscription

// Infrastructure/API features
Scope: ScopeAPI

// System-wide features
Scope: ScopeGlobal
```

### 3. Set Expiration Dates

```go
// For temporary flags
expiresAt := time.Now().AddDate(0, 3, 0)  // 3 months
flag.ExpiresAt = &expiresAt
```

### 4. Document Dependencies

```go
flag := &featureflags.FeatureFlag{
    Key:          "advanced_export",
    Dependencies: []string{"export_pdf", "export_excel"},
}

// In code
if flagService.IsEnabled("export_pdf", ctx) &&
   flagService.IsEnabled("export_excel", ctx) {
    // Only then enable advanced_export
}
```

### 5. Clean Up Old Flags

```go
// Regularly archive unused flags
var oldFlags []featureflags.FeatureFlag
db.Where("status = ? AND deprecated_at < ?",
    featureflags.StatusDeprecated,
    time.Now().AddDate(0, -6, 0),  // deprecated > 6 months ago
).Find(&oldFlags)

for _, flag := range oldFlags {
    flagService.UpdateFlag(flag.Key, map[string]interface{}{
        "status":     featureflags.StatusArchived,
        "is_enabled": false,
    }, &adminUserID)
}
```

## Monitoring and Debugging

### 1. View Flag History

```sql
SELECT
    change_type,
    previous_value,
    new_value,
    changed_by,
    change_reason,
    created_at
FROM feature_flag_histories
WHERE feature_flag_key = 'advanced_analytics'
ORDER BY created_at DESC;
```

### 2. Check User-Specific Overrides

```sql
SELECT
    uff.feature_flag_key,
    uff.is_enabled,
    uff.reason,
    uff.expires_at,
    u.email as set_by_email
FROM user_feature_flags uff
LEFT JOIN users u ON uff.set_by = u.id
WHERE uff.user_id = '<user-uuid>';
```

### 3. Monitor Evaluation Performance

```go
// Track evaluation timing
start := time.Now()
result, err := flagService.Evaluate("feature_key", ctx)
duration := time.Since(start)

if duration > 100*time.Millisecond {
    log.Warn("Slow flag evaluation", "key", "feature_key", "duration", duration)
}
```

## Common Patterns

### 1. Feature + Subscription

```go
// Feature available to paying tiers + beta users
func isFeatureAvailable(userID uuid.UUID) bool {
    // Check subscription tier
    sub, _ := subscriptionService.GetUserSubscription(userID)
    if sub.Plan.Tier != subscriptions.TierFree {
        return true
    }

    // Check beta user override
    ctx := featureflags.EvaluationContext{UserID: &userID}
    isEnabled, _ := flagService.IsEnabled("beta_feature", ctx)
    return isEnabled
}
```

### 2. Graceful Degradation

```go
// Provide fallback when feature is disabled
func getRecommendations(userID uuid.UUID) []Recommendation {
    ctx := featureflags.EvaluationContext{UserID: &userID}

    if flagService.IsEnabled("ai_recommendations", ctx) {
        return getAIRecommendations(userID)
    }

    // Fallback to basic recommendations
    return getBasicRecommendations(userID)
}
```

### 3. Feature Gate with Config

```go
// Feature flag with configuration
value, _ := flagService.GetValue("ai_config", ctx)
if value != nil {
    var config struct {
        Model       string  `json:"model"`
        Temperature float64 `json:"temperature"`
    }
    json.Unmarshal([]byte(*value), &config)
    // Use config
}
```

## Database Migrations

Include feature flag models in migrations:

```go
import "triplanner/featureflags"

models := append(coreModels, featureflags.GetModels()...)
db.AutoMigrate(models...)
```

## Summary

The feature flag system provides:

- ✅ **Independent operation** from subscriptions
- ✅ **Flexible integration** with subscription tiers
- ✅ **Multiple evaluation scopes** (global, user, subscription, API)
- ✅ **Various flag types** (boolean, string, number, JSON, percentage)
- ✅ **Comprehensive auditing** of all changes and evaluations
- ✅ **Gradual rollouts** with percentage-based targeting
- ✅ **Emergency controls** for quick feature toggling

Together with the subscription system, it enables:

- Fine-grained access control
- Tier-based feature enablement
- Beta testing and gradual rollouts
- A/B testing capabilities
- Dynamic configuration
- Emergency feature toggles
