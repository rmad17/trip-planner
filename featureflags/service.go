package featureflags

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrFeatureFlagNotFound = errors.New("feature flag not found")
	ErrInvalidFlagType     = errors.New("invalid flag type")
)

// Service provides feature flag business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new feature flag service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// EvaluationContext contains context for evaluating a feature flag
type EvaluationContext struct {
	UserID      *uuid.UUID
	APIEndpoint *string
	HTTPMethod  *string
}

// EvaluationResult contains the result of a feature flag evaluation
type EvaluationResult struct {
	IsEnabled bool
	Value     *string
	Source    string // "default", "subscription", "user_override", "api_override", "rollout"
}

// IsEnabled checks if a feature flag is enabled for a given context
func (s *Service) IsEnabled(key string, ctx EvaluationContext) (bool, error) {
	result, err := s.Evaluate(key, ctx)
	if err != nil {
		return false, err
	}
	return result.IsEnabled, nil
}

// Evaluate evaluates a feature flag and returns the result
func (s *Service) Evaluate(key string, ctx EvaluationContext) (*EvaluationResult, error) {
	// Get the base flag
	var flag FeatureFlag
	err := s.db.Where("key = ?", key).First(&flag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &EvaluationResult{IsEnabled: false, Source: "not_found"}, nil
		}
		return nil, err
	}

	// Check if flag is globally disabled
	if !flag.IsEnabled || flag.Status != StatusActive {
		s.recordEvaluation(key, ctx, false, flag.DefaultValue, "disabled")
		return &EvaluationResult{IsEnabled: false, Value: flag.DefaultValue, Source: "disabled"}, nil
	}

	// Check for API override first (highest priority)
	if ctx.APIEndpoint != nil {
		apiResult, found := s.checkAPIOverride(key, *ctx.APIEndpoint, ctx.HTTPMethod)
		if found {
			s.recordEvaluation(key, ctx, apiResult.IsEnabled, apiResult.Value, "api_override")
			return apiResult, nil
		}
	}

	// Check for user override
	if ctx.UserID != nil {
		userResult, found := s.checkUserOverride(key, *ctx.UserID)
		if found {
			s.recordEvaluation(key, ctx, userResult.IsEnabled, userResult.Value, "user_override")
			return userResult, nil
		}
	}

	// Check rollout percentage
	if flag.RolloutPercentage > 0 && flag.RolloutPercentage < 100 {
		if ctx.UserID != nil {
			inRollout := s.isInRollout(*ctx.UserID, key, flag.RolloutPercentage)
			if !inRollout {
				s.recordEvaluation(key, ctx, false, flag.DefaultValue, "rollout_excluded")
				return &EvaluationResult{IsEnabled: false, Value: flag.DefaultValue, Source: "rollout_excluded"}, nil
			}
		}
	}

	// Return default value
	isEnabled := flag.IsEnabled
	if flag.DefaultValue != nil {
		switch *flag.DefaultValue {
		case "true", "1":
			isEnabled = true
		case "false", "0":
			isEnabled = false
		}
	}

	s.recordEvaluation(key, ctx, isEnabled, flag.DefaultValue, "default")
	return &EvaluationResult{IsEnabled: isEnabled, Value: flag.DefaultValue, Source: "default"}, nil
}

// checkUserOverride checks if there's a user-specific override
func (s *Service) checkUserOverride(key string, userID uuid.UUID) (*EvaluationResult, bool) {
	var override UserFeatureFlag
	err := s.db.Where("feature_flag_key = ? AND user_id = ?", key, userID).First(&override).Error
	if err != nil {
		return nil, false
	}

	// Check if override has expired
	if override.ExpiresAt != nil && override.ExpiresAt.Before(time.Now()) {
		return nil, false
	}

	return &EvaluationResult{
		IsEnabled: override.IsEnabled,
		Value:     override.CustomValue,
		Source:    "user_override",
	}, true
}

// checkAPIOverride checks if there's an API-specific override
func (s *Service) checkAPIOverride(key string, endpoint string, method *string) (*EvaluationResult, bool) {
	query := s.db.Where("feature_flag_key = ? AND api_endpoint = ?", key, endpoint)
	if method != nil {
		query = query.Where("(http_method = ? OR http_method IS NULL)", *method)
	}

	var overrides []APIFeatureFlag
	err := query.Order("priority DESC").Find(&overrides).Error
	if err != nil || len(overrides) == 0 {
		return nil, false
	}

	override := overrides[0]

	// Check if override has expired
	if override.ExpiresAt != nil && override.ExpiresAt.Before(time.Now()) {
		return nil, false
	}

	return &EvaluationResult{
		IsEnabled: override.IsEnabled,
		Value:     override.CustomValue,
		Source:    "api_override",
	}, true
}

// isInRollout determines if a user is in the rollout percentage
// Uses consistent hashing to ensure same user always gets same result
func (s *Service) isInRollout(userID uuid.UUID, flagKey string, percentage int) bool {
	// Create a consistent hash from userID and flagKey
	hash := sha256.Sum256([]byte(userID.String() + flagKey))
	hashInt := binary.BigEndian.Uint64(hash[:8])

	// Convert to percentage (0-100)
	userPercentage := int(hashInt % 100)

	return userPercentage < percentage
}

// recordEvaluation records a feature flag evaluation for analytics
func (s *Service) recordEvaluation(key string, ctx EvaluationContext, wasEnabled bool, value *string, source string) {
	// Only record if we want to track evaluations (can be controlled by a config)
	// For now, we'll record all evaluations but you might want to sample in production

	evaluation := FeatureFlagEvaluation{
		FeatureFlagKey: key,
		UserID:         ctx.UserID,
		APIEndpoint:    ctx.APIEndpoint,
		EvaluatedValue: value,
		WasEnabled:     wasEnabled,
		Source:         source,
		EvaluatedAt:    time.Now(),
	}

	// Fire and forget - don't block on this
	go s.db.Create(&evaluation)
}

// GetValue gets the value of a feature flag (for non-boolean flags)
func (s *Service) GetValue(key string, ctx EvaluationContext) (*string, error) {
	result, err := s.Evaluate(key, ctx)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetIntValue gets an integer value from a feature flag
func (s *Service) GetIntValue(key string, ctx EvaluationContext, defaultValue int) (int, error) {
	value, err := s.GetValue(key, ctx)
	if err != nil || value == nil {
		return defaultValue, err
	}

	var intValue int
	err = json.Unmarshal([]byte(*value), &intValue)
	if err != nil {
		return defaultValue, err
	}

	return intValue, nil
}

// GetStringValue gets a string value from a feature flag
func (s *Service) GetStringValue(key string, ctx EvaluationContext, defaultValue string) (string, error) {
	value, err := s.GetValue(key, ctx)
	if err != nil || value == nil {
		return defaultValue, err
	}
	return *value, nil
}

// SetUserOverride sets a user-specific override for a feature flag
func (s *Service) SetUserOverride(userID uuid.UUID, key string, isEnabled bool, customValue *string, reason *string, setBy *uuid.UUID) error {
	// Check if flag exists
	var flag FeatureFlag
	err := s.db.Where("key = ?", key).First(&flag).Error
	if err != nil {
		return ErrFeatureFlagNotFound
	}

	// Create or update override
	override := UserFeatureFlag{
		UserID:         userID,
		FeatureFlagKey: key,
		IsEnabled:      isEnabled,
		CustomValue:    customValue,
		Reason:         reason,
		SetBy:          setBy,
	}

	return s.db.Where("user_id = ? AND feature_flag_key = ?", userID, key).
		Assign(override).
		FirstOrCreate(&override).Error
}

// RemoveUserOverride removes a user-specific override
func (s *Service) RemoveUserOverride(userID uuid.UUID, key string) error {
	return s.db.Where("user_id = ? AND feature_flag_key = ?", userID, key).
		Delete(&UserFeatureFlag{}).Error
}

// SetAPIOverride sets an API-specific override for a feature flag
func (s *Service) SetAPIOverride(endpoint string, method *string, key string, isEnabled bool, customValue *string, priority int) error {
	// Check if flag exists
	var flag FeatureFlag
	err := s.db.Where("key = ?", key).First(&flag).Error
	if err != nil {
		return ErrFeatureFlagNotFound
	}

	override := APIFeatureFlag{
		APIEndpoint:    endpoint,
		HTTPMethod:     method,
		FeatureFlagKey: key,
		IsEnabled:      isEnabled,
		CustomValue:    customValue,
		Priority:       priority,
	}

	query := s.db.Where("api_endpoint = ? AND feature_flag_key = ?", endpoint, key)
	if method != nil {
		query = query.Where("http_method = ?", *method)
	} else {
		query = query.Where("http_method IS NULL")
	}

	return query.Assign(override).FirstOrCreate(&override).Error
}

// CreateFlag creates a new feature flag
func (s *Service) CreateFlag(flag *FeatureFlag, createdBy *uuid.UUID) error {
	flag.CreatedBy = createdBy
	flag.LastModifiedBy = createdBy

	err := s.db.Create(flag).Error
	if err != nil {
		return err
	}

	// Record history
	s.recordFlagHistory(flag.Key, "created", nil, core.StringPtr("created"), createdBy, core.StringPtr("Flag created"))

	return nil
}

// UpdateFlag updates an existing feature flag
func (s *Service) UpdateFlag(key string, updates map[string]interface{}, modifiedBy *uuid.UUID) error {
	var flag FeatureFlag
	err := s.db.Where("key = ?", key).First(&flag).Error
	if err != nil {
		return ErrFeatureFlagNotFound
	}

	// Store previous state for history
	previousState, _ := json.Marshal(flag)

	updates["last_modified_by"] = modifiedBy
	err = s.db.Model(&flag).Updates(updates).Error
	if err != nil {
		return err
	}

	// Record history
	newState, _ := json.Marshal(updates)
	s.recordFlagHistory(key, "updated", core.StringPtr(string(previousState)), core.StringPtr(string(newState)), modifiedBy, nil)

	return nil
}

// recordFlagHistory records changes to feature flags
func (s *Service) recordFlagHistory(key string, changeType string, previousValue, newValue *string, changedBy *uuid.UUID, reason *string) {
	history := FeatureFlagHistory{
		FeatureFlagKey: key,
		ChangeType:     changeType,
		PreviousValue:  previousValue,
		NewValue:       newValue,
		ChangedBy:      changedBy,
		ChangeReason:   reason,
	}

	go s.db.Create(&history)
}

// GetAllFlags retrieves all feature flags
func (s *Service) GetAllFlags() ([]FeatureFlag, error) {
	var flags []FeatureFlag
	err := s.db.Order("key ASC").Find(&flags).Error
	return flags, err
}

// GetFlagsByScope retrieves flags by scope
func (s *Service) GetFlagsByScope(scope FeatureFlagScope) ([]FeatureFlag, error) {
	var flags []FeatureFlag
	err := s.db.Where("scope = ?", scope).Order("key ASC").Find(&flags).Error
	return flags, err
}

// GetFlagsByStatus retrieves flags by status
func (s *Service) GetFlagsByStatus(status FeatureFlagStatus) ([]FeatureFlag, error) {
	var flags []FeatureFlag
	err := s.db.Where("status = ?", status).Order("key ASC").Find(&flags).Error
	return flags, err
}
