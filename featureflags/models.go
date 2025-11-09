package featureflags

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

// FeatureFlagType represents the type of feature flag
type FeatureFlagType string

const (
	TypeBoolean    FeatureFlagType = "boolean"    // Simple on/off flag
	TypeString     FeatureFlagType = "string"     // String value
	TypeNumber     FeatureFlagType = "number"     // Numeric value
	TypeJSON       FeatureFlagType = "json"       // JSON configuration
	TypePercentage FeatureFlagType = "percentage" // Percentage rollout (0-100)
)

// FeatureFlagScope defines where the flag can be applied
type FeatureFlagScope string

const (
	ScopeGlobal       FeatureFlagScope = "global"       // Applied to all users
	ScopeUser         FeatureFlagScope = "user"         // Applied per user
	ScopeSubscription FeatureFlagScope = "subscription" // Applied based on subscription tier
	ScopeAPI          FeatureFlagScope = "api"          // Applied per API endpoint
)

// FeatureFlagStatus represents the status of a feature flag
type FeatureFlagStatus string

const (
	StatusDraft      FeatureFlagStatus = "draft"      // Being developed
	StatusActive     FeatureFlagStatus = "active"     // Active and in use
	StatusDeprecated FeatureFlagStatus = "deprecated" // Marked for removal
	StatusArchived   FeatureFlagStatus = "archived"   // No longer in use
)

// FeatureFlag defines a feature flag in the system
type FeatureFlag struct {
	core.BaseModel
	Key               string            `json:"key" gorm:"unique;not null" example:"advanced_analytics" description:"Unique key for the feature flag"`
	Name              string            `json:"name" gorm:"not null" example:"Advanced Analytics" description:"Display name of the feature"`
	Description       *string           `json:"description" example:"Enable advanced analytics dashboard" description:"Description of what this flag controls"`
	Type              FeatureFlagType   `json:"type" gorm:"type:varchar(20);not null;default:'boolean'" example:"boolean" description:"Type of flag value"`
	Scope             FeatureFlagScope  `json:"scope" gorm:"type:varchar(20);not null;default:'global'" example:"subscription" description:"Scope of the flag"`
	Status            FeatureFlagStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft'" example:"active" description:"Current status of the flag"`
	DefaultValue      *string           `json:"default_value" example:"false" description:"Default value when flag is not explicitly set"`
	IsEnabled         bool              `json:"is_enabled" gorm:"default:false" description:"Master switch for this flag"`
	RolloutPercentage int               `json:"rollout_percentage" gorm:"default:0" description:"Percentage of users to enable (0-100)"`
	Tags              []string          `json:"tags" gorm:"type:text[]" description:"Tags for organizing flags"`
	OwnerTeam         *string           `json:"owner_team" example:"backend" description:"Team responsible for this flag"`
	OwnerEmail        *string           `json:"owner_email" example:"team@example.com" description:"Contact email for flag owner"`
	CreatedBy         *uuid.UUID        `json:"created_by" gorm:"type:uuid" description:"User who created the flag"`
	LastModifiedBy    *uuid.UUID        `json:"last_modified_by" gorm:"type:uuid" description:"User who last modified the flag"`
	DeprecatedAt      *time.Time        `json:"deprecated_at" description:"When the flag was deprecated"`
	ExpiresAt         *time.Time        `json:"expires_at" description:"When the flag should be automatically disabled"`
	Dependencies      []string          `json:"dependencies" gorm:"type:text[]" description:"Keys of flags this flag depends on"`
	Notes             *string           `json:"notes" description:"Internal notes about the flag"`
	UserOverrides     []UserFeatureFlag `json:"user_overrides,omitempty" gorm:"foreignKey:FeatureFlagKey;references:Key" description:"User-specific overrides"`
	APIOverrides      []APIFeatureFlag  `json:"api_overrides,omitempty" gorm:"foreignKey:FeatureFlagKey;references:Key" description:"API-specific overrides"`
}

// UserFeatureFlag allows per-user overrides of feature flags
type UserFeatureFlag struct {
	core.BaseModel
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user"`
	FeatureFlagKey string     `json:"feature_flag_key" gorm:"not null" example:"advanced_analytics" description:"Key of the feature flag"`
	IsEnabled      bool       `json:"is_enabled" gorm:"default:true" description:"Whether this feature is enabled for this user"`
	CustomValue    *string    `json:"custom_value" description:"Custom value for this user (overrides default)"`
	Reason         *string    `json:"reason" example:"beta_tester" description:"Reason for the override"`
	ExpiresAt      *time.Time `json:"expires_at" description:"When this override expires"`
	SetBy          *uuid.UUID `json:"set_by" gorm:"type:uuid" description:"Who set this override (admin ID)"`
	Notes          *string    `json:"notes" description:"Notes about this override"`
}

// APIFeatureFlag allows per-API endpoint control of features
type APIFeatureFlag struct {
	core.BaseModel
	APIEndpoint    string     `json:"api_endpoint" gorm:"not null" example:"/api/v1/trip-plans" description:"API endpoint pattern"`
	HTTPMethod     *string    `json:"http_method" gorm:"type:varchar(10)" example:"GET" description:"HTTP method (null = all methods)"`
	FeatureFlagKey string     `json:"feature_flag_key" gorm:"not null" example:"rate_limiting" description:"Key of the feature flag"`
	IsEnabled      bool       `json:"is_enabled" gorm:"default:true" description:"Whether this feature is enabled for this endpoint"`
	CustomValue    *string    `json:"custom_value" description:"Custom value for this endpoint (e.g., rate limit)"`
	Priority       int        `json:"priority" gorm:"default:0" description:"Priority for overlapping rules (higher = higher priority)"`
	Description    *string    `json:"description" description:"Description of why this override exists"`
	ExpiresAt      *time.Time `json:"expires_at" description:"When this override expires"`
}

// FeatureFlagHistory tracks changes to feature flags over time
type FeatureFlagHistory struct {
	core.BaseModel
	FeatureFlagKey string     `json:"feature_flag_key" gorm:"not null" example:"advanced_analytics" description:"Key of the feature flag"`
	ChangeType     string     `json:"change_type" gorm:"not null" example:"enabled" description:"Type of change (created, enabled, disabled, updated, deleted)"`
	PreviousValue  *string    `json:"previous_value" description:"Previous value (JSON)"`
	NewValue       *string    `json:"new_value" description:"New value (JSON)"`
	ChangedBy      *uuid.UUID `json:"changed_by" gorm:"type:uuid" description:"User who made the change"`
	ChangeReason   *string    `json:"change_reason" description:"Reason for the change"`
	AffectedUsers  *int       `json:"affected_users" description:"Estimated number of affected users"`
	IPAddress      *string    `json:"ip_address" description:"IP address of change origin"`
	UserAgent      *string    `json:"user_agent" description:"User agent of change origin"`
}

// FeatureFlagEvaluation tracks when and how flags are evaluated
// Useful for analytics and debugging
type FeatureFlagEvaluation struct {
	core.BaseModel
	FeatureFlagKey string     `json:"feature_flag_key" gorm:"not null;index" example:"advanced_analytics" description:"Key of the feature flag"`
	UserID         *uuid.UUID `json:"user_id" gorm:"type:uuid;index" description:"User ID if user-scoped"`
	APIEndpoint    *string    `json:"api_endpoint" gorm:"index" description:"API endpoint if API-scoped"`
	EvaluatedValue *string    `json:"evaluated_value" description:"The value that was evaluated"`
	WasEnabled     bool       `json:"was_enabled" description:"Whether the flag was enabled"`
	Source         string     `json:"source" gorm:"not null" example:"subscription" description:"Source of the value (default, subscription, user_override, api_override)"`
	EvaluatedAt    time.Time  `json:"evaluated_at" gorm:"not null;index" description:"When the evaluation occurred"`
}

// GetValidTypes returns all valid feature flag types
func GetValidTypes() []FeatureFlagType {
	return []FeatureFlagType{
		TypeBoolean,
		TypeString,
		TypeNumber,
		TypeJSON,
		TypePercentage,
	}
}

// IsValidType checks if a type is valid
func IsValidType(flagType string) bool {
	for _, validType := range GetValidTypes() {
		if string(validType) == flagType {
			return true
		}
	}
	return false
}

// GetValidScopes returns all valid feature flag scopes
func GetValidScopes() []FeatureFlagScope {
	return []FeatureFlagScope{
		ScopeGlobal,
		ScopeUser,
		ScopeSubscription,
		ScopeAPI,
	}
}

// IsValidScope checks if a scope is valid
func IsValidScope(scope string) bool {
	for _, validScope := range GetValidScopes() {
		if string(validScope) == scope {
			return true
		}
	}
	return false
}

// GetValidStatuses returns all valid feature flag statuses
func GetValidStatuses() []FeatureFlagStatus {
	return []FeatureFlagStatus{
		StatusDraft,
		StatusActive,
		StatusDeprecated,
		StatusArchived,
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
		&FeatureFlag{},
		&UserFeatureFlag{},
		&APIFeatureFlag{},
		&FeatureFlagHistory{},
		&FeatureFlagEvaluation{},
	}
}
