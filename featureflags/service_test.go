package featureflags

import (
	"testing"
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

func TestGetValidTypes(t *testing.T) {
	types := GetValidTypes()

	if len(types) != 5 {
		t.Errorf("Expected 5 types, got %d", len(types))
	}

	expectedTypes := map[FeatureFlagType]bool{
		TypeBoolean:    true,
		TypeString:     true,
		TypeNumber:     true,
		TypeJSON:       true,
		TypePercentage: true,
	}

	for _, flagType := range types {
		if !expectedTypes[flagType] {
			t.Errorf("Unexpected type: %s", flagType)
		}
	}
}

func TestIsValidType(t *testing.T) {
	tests := []struct {
		name     string
		flagType string
		valid    bool
	}{
		{"Boolean type", "boolean", true},
		{"String type", "string", true},
		{"Number type", "number", true},
		{"JSON type", "json", true},
		{"Percentage type", "percentage", true},
		{"Invalid type", "array", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidType(tt.flagType)
			if result != tt.valid {
				t.Errorf("IsValidType(%s) = %v, want %v", tt.flagType, result, tt.valid)
			}
		})
	}
}

func TestGetValidScopes(t *testing.T) {
	scopes := GetValidScopes()

	if len(scopes) != 4 {
		t.Errorf("Expected 4 scopes, got %d", len(scopes))
	}

	expectedScopes := map[FeatureFlagScope]bool{
		ScopeGlobal:       true,
		ScopeUser:         true,
		ScopeSubscription: true,
		ScopeAPI:          true,
	}

	for _, scope := range scopes {
		if !expectedScopes[scope] {
			t.Errorf("Unexpected scope: %s", scope)
		}
	}
}

func TestIsValidScope(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		valid bool
	}{
		{"Global scope", "global", true},
		{"User scope", "user", true},
		{"Subscription scope", "subscription", true},
		{"API scope", "api", true},
		{"Invalid scope", "organization", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidScope(tt.scope)
			if result != tt.valid {
				t.Errorf("IsValidScope(%s) = %v, want %v", tt.scope, result, tt.valid)
			}
		})
	}
}

func TestGetValidStatuses(t *testing.T) {
	statuses := GetValidStatuses()

	if len(statuses) != 4 {
		t.Errorf("Expected 4 statuses, got %d", len(statuses))
	}

	expectedStatuses := map[FeatureFlagStatus]bool{
		StatusDraft:      true,
		StatusActive:     true,
		StatusDeprecated: true,
		StatusArchived:   true,
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
		{"Draft status", "draft", true},
		{"Active status", "active", true},
		{"Deprecated status", "deprecated", true},
		{"Archived status", "archived", true},
		{"Invalid status", "pending", false},
		{"Empty string", "", false},
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

func TestFeatureFlag_BasicFields(t *testing.T) {
	flag := FeatureFlag{
		Key:          "test_feature",
		Name:         "Test Feature",
		Description:  core.StringPtr("A test feature"),
		Type:         TypeBoolean,
		Scope:        ScopeUser,
		Status:       StatusActive,
		DefaultValue: core.StringPtr("false"),
		IsEnabled:    true,
		OwnerTeam:    core.StringPtr("engineering"),
	}

	if flag.Key != "test_feature" {
		t.Error("Key not set correctly")
	}
	if flag.Type != TypeBoolean {
		t.Error("Type should be boolean")
	}
	if flag.Scope != ScopeUser {
		t.Error("Scope should be user")
	}
	if flag.Status != StatusActive {
		t.Error("Status should be active")
	}
	if !flag.IsEnabled {
		t.Error("Flag should be enabled")
	}
}

func TestUserFeatureFlag_Override(t *testing.T) {
	userID := uuid.New()
	setBy := uuid.New()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	override := UserFeatureFlag{
		UserID:         userID,
		FeatureFlagKey: "beta_feature",
		IsEnabled:      true,
		CustomValue:    core.StringPtr(`{"version": "2.0"}`),
		Reason:         core.StringPtr("beta_tester"),
		ExpiresAt:      &expiresAt,
		SetBy:          &setBy,
	}

	if override.UserID != userID {
		t.Error("UserID not set correctly")
	}
	if override.FeatureFlagKey != "beta_feature" {
		t.Error("FeatureFlagKey not set correctly")
	}
	if !override.IsEnabled {
		t.Error("Override should be enabled")
	}
	if override.Reason == nil || *override.Reason != "beta_tester" {
		t.Error("Reason should be beta_tester")
	}
	if override.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
}

func TestAPIFeatureFlag_Configuration(t *testing.T) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	apiFlag := APIFeatureFlag{
		APIEndpoint:    "/api/v1/analytics",
		HTTPMethod:     core.StringPtr("GET"),
		FeatureFlagKey: "rate_limiting",
		IsEnabled:      true,
		CustomValue:    core.StringPtr(`{"requests_per_minute": 100}`),
		Priority:       10,
		Description:    core.StringPtr("Rate limit for analytics endpoint"),
		ExpiresAt:      &expiresAt,
	}

	if apiFlag.APIEndpoint != "/api/v1/analytics" {
		t.Error("APIEndpoint not set correctly")
	}
	if apiFlag.HTTPMethod == nil || *apiFlag.HTTPMethod != "GET" {
		t.Error("HTTPMethod should be GET")
	}
	if !apiFlag.IsEnabled {
		t.Error("API flag should be enabled")
	}
	if apiFlag.Priority != 10 {
		t.Error("Priority should be 10")
	}
}

func TestFeatureFlagHistory_Tracking(t *testing.T) {
	changedBy := uuid.New()

	history := FeatureFlagHistory{
		FeatureFlagKey: "test_feature",
		ChangeType:     "enabled",
		PreviousValue:  core.StringPtr(`{"is_enabled": false}`),
		NewValue:       core.StringPtr(`{"is_enabled": true}`),
		ChangedBy:      &changedBy,
		ChangeReason:   core.StringPtr("User request"),
		AffectedUsers:  core.IntPtr(1000),
		IPAddress:      core.StringPtr("192.168.1.1"),
	}

	if history.FeatureFlagKey != "test_feature" {
		t.Error("FeatureFlagKey not set correctly")
	}
	if history.ChangeType != "enabled" {
		t.Error("ChangeType should be enabled")
	}
	if history.AffectedUsers == nil || *history.AffectedUsers != 1000 {
		t.Error("AffectedUsers should be 1000")
	}
}

func TestFeatureFlagEvaluation_Recording(t *testing.T) {
	userID := uuid.New()
	evaluatedAt := time.Now()

	evaluation := FeatureFlagEvaluation{
		FeatureFlagKey: "new_ui",
		UserID:         &userID,
		APIEndpoint:    core.StringPtr("/api/v1/dashboard"),
		EvaluatedValue: core.StringPtr("true"),
		WasEnabled:     true,
		Source:         "user_override",
		EvaluatedAt:    evaluatedAt,
	}

	if evaluation.FeatureFlagKey != "new_ui" {
		t.Error("FeatureFlagKey not set correctly")
	}
	if !evaluation.WasEnabled {
		t.Error("Flag should be enabled in evaluation")
	}
	if evaluation.Source != "user_override" {
		t.Error("Source should be user_override")
	}
}

func TestRolloutPercentage(t *testing.T) {
	tests := []struct {
		name       string
		percentage int
		valid      bool
	}{
		{"0 percent", 0, true},
		{"25 percent", 25, true},
		{"50 percent", 50, true},
		{"100 percent", 100, true},
		{"Negative", -10, false},
		{"Over 100", 150, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := FeatureFlag{
				RolloutPercentage: tt.percentage,
			}

			isValid := flag.RolloutPercentage >= 0 && flag.RolloutPercentage <= 100
			if isValid != tt.valid {
				t.Errorf("Rollout percentage %d validity = %v, want %v", tt.percentage, isValid, tt.valid)
			}
		})
	}
}

func TestFeatureFlagDependencies(t *testing.T) {
	flag := FeatureFlag{
		Key:          "advanced_export",
		Dependencies: []string{"export_pdf", "export_excel"},
	}

	if len(flag.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(flag.Dependencies))
	}

	expectedDeps := map[string]bool{
		"export_pdf":   true,
		"export_excel": true,
	}

	for _, dep := range flag.Dependencies {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}
}

func TestFeatureFlagExpiration(t *testing.T) {
	t.Run("Not expired", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		flag := FeatureFlag{
			ExpiresAt: &future,
		}

		if flag.ExpiresAt.Before(time.Now()) {
			t.Error("Flag should not be expired")
		}
	})

	t.Run("Expired", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		flag := FeatureFlag{
			ExpiresAt: &past,
		}

		if !flag.ExpiresAt.Before(time.Now()) {
			t.Error("Flag should be expired")
		}
	})

	t.Run("No expiration", func(t *testing.T) {
		flag := FeatureFlag{
			ExpiresAt: nil,
		}

		if flag.ExpiresAt != nil {
			t.Error("Flag should not have expiration")
		}
	})
}

func TestFeatureFlagTags(t *testing.T) {
	flag := FeatureFlag{
		Tags: []string{"beta", "experimental", "ui"},
	}

	if len(flag.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(flag.Tags))
	}

	expectedTags := map[string]bool{
		"beta":         true,
		"experimental": true,
		"ui":           true,
	}

	for _, tag := range flag.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestEvaluationContext(t *testing.T) {
	t.Run("User context only", func(t *testing.T) {
		userID := uuid.New()
		ctx := EvaluationContext{
			UserID: &userID,
		}

		if ctx.UserID == nil {
			t.Error("UserID should be set")
		}
		if ctx.APIEndpoint != nil {
			t.Error("APIEndpoint should be nil")
		}
	})

	t.Run("API context only", func(t *testing.T) {
		endpoint := "/api/v1/trips"
		method := "POST"
		ctx := EvaluationContext{
			APIEndpoint: &endpoint,
			HTTPMethod:  &method,
		}

		if ctx.APIEndpoint == nil || *ctx.APIEndpoint != endpoint {
			t.Error("APIEndpoint should be set")
		}
		if ctx.HTTPMethod == nil || *ctx.HTTPMethod != method {
			t.Error("HTTPMethod should be set")
		}
		if ctx.UserID != nil {
			t.Error("UserID should be nil")
		}
	})

	t.Run("Full context", func(t *testing.T) {
		userID := uuid.New()
		endpoint := "/api/v1/trips"
		method := "POST"
		ctx := EvaluationContext{
			UserID:      &userID,
			APIEndpoint: &endpoint,
			HTTPMethod:  &method,
		}

		if ctx.UserID == nil {
			t.Error("UserID should be set")
		}
		if ctx.APIEndpoint == nil {
			t.Error("APIEndpoint should be set")
		}
		if ctx.HTTPMethod == nil {
			t.Error("HTTPMethod should be set")
		}
	})
}

func TestEvaluationResult(t *testing.T) {
	t.Run("Enabled result", func(t *testing.T) {
		result := EvaluationResult{
			IsEnabled: true,
			Value:     core.StringPtr("true"),
			Source:    "default",
		}

		if !result.IsEnabled {
			t.Error("Result should be enabled")
		}
		if result.Value == nil || *result.Value != "true" {
			t.Error("Value should be true")
		}
		if result.Source != "default" {
			t.Error("Source should be default")
		}
	})

	t.Run("Disabled result", func(t *testing.T) {
		result := EvaluationResult{
			IsEnabled: false,
			Value:     nil,
			Source:    "rollout_excluded",
		}

		if result.IsEnabled {
			t.Error("Result should be disabled")
		}
		if result.Value != nil {
			t.Error("Value should be nil")
		}
		if result.Source != "rollout_excluded" {
			t.Error("Source should be rollout_excluded")
		}
	})
}

func TestOverrideExpiration(t *testing.T) {
	t.Run("User override expiration check", func(t *testing.T) {
		expiredTime := time.Now().Add(-1 * time.Hour)
		validTime := time.Now().Add(1 * time.Hour)

		expiredOverride := UserFeatureFlag{
			ExpiresAt: &expiredTime,
		}

		validOverride := UserFeatureFlag{
			ExpiresAt: &validTime,
		}

		if !expiredOverride.ExpiresAt.Before(time.Now()) {
			t.Error("Override should be expired")
		}

		if validOverride.ExpiresAt.Before(time.Now()) {
			t.Error("Override should not be expired")
		}
	})

	t.Run("API override expiration check", func(t *testing.T) {
		expiredTime := time.Now().Add(-1 * time.Hour)
		validTime := time.Now().Add(1 * time.Hour)

		expiredOverride := APIFeatureFlag{
			ExpiresAt: &expiredTime,
		}

		validOverride := APIFeatureFlag{
			ExpiresAt: &validTime,
		}

		if !expiredOverride.ExpiresAt.Before(time.Now()) {
			t.Error("Override should be expired")
		}

		if validOverride.ExpiresAt.Before(time.Now()) {
			t.Error("Override should not be expired")
		}
	})
}
