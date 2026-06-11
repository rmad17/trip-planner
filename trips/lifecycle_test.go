package trips

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidStatusTransition tests all allowed and blocked state transitions.
func TestValidStatusTransition(t *testing.T) {
	tests := []struct {
		from    TripStatus
		to      TripStatus
		allowed bool
	}{
		// Valid forward transitions
		{TripStatusPlanning, TripStatusConfirmed, true},
		{TripStatusPlanning, TripStatusCancelled, true},
		{TripStatusConfirmed, TripStatusOngoing, true},
		{TripStatusConfirmed, TripStatusCancelled, true},
		{TripStatusOngoing, TripStatusCompleted, true},
		{TripStatusOngoing, TripStatusCancelled, true},
		// Terminal states cannot transition
		{TripStatusCompleted, TripStatusPlanning, false},
		{TripStatusCompleted, TripStatusOngoing, false},
		{TripStatusCompleted, TripStatusConfirmed, false},
		{TripStatusCancelled, TripStatusPlanning, false},
		{TripStatusCancelled, TripStatusConfirmed, false},
		// No backwards transitions
		{TripStatusConfirmed, TripStatusPlanning, false},
		{TripStatusOngoing, TripStatusPlanning, false},
		{TripStatusOngoing, TripStatusConfirmed, false},
		// Same-state is not a valid transition (not in allowed list)
		{TripStatusPlanning, TripStatusPlanning, false},
		{TripStatusOngoing, TripStatusOngoing, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := isValidTransition(tt.from, tt.to)
			assert.Equal(t, tt.allowed, got,
				"isValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.allowed)
		})
	}
}

// TestGetEmergencyNumber tests country code lookup and fallback.
func TestGetEmergencyNumber(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"IN", "112"},
		{"US", "911"},
		{"GB", "999"},
		{"AU", "000"},
		{"JP", "110"},
		{"SG", "999"},
		{"NP", "100"},
		{"UNKNOWN", "112"}, // fallback
		{"", "112"},        // empty fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := getEmergencyNumber(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestISOCode tests the country-name to ISO code mapper.
func TestISOCode(t *testing.T) {
	tests := []struct {
		country  string
		expected string
	}{
		{"India", "IN"},
		{"United States", "US"},
		{"USA", "US"},
		{"France", "FR"},
		{"Nepal", "NP"},
		{"Sri Lanka", "LK"},
		{"Andorra", "Andorra"}, // not in map, returns input
	}
	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			assert.Equal(t, tt.expected, isoCode(tt.country))
		})
	}
}

// TestValidTransitions_TableCompleteness verifies all defined TripStatus values
// appear as keys in validTransitions (catches typos in the constant definitions).
func TestValidTransitions_TableCompleteness(t *testing.T) {
	statuses := []TripStatus{
		TripStatusPlanning,
		TripStatusConfirmed,
		TripStatusOngoing,
		TripStatusCompleted,
		TripStatusCancelled,
	}
	for _, s := range statuses {
		_, ok := validTransitions[s]
		assert.True(t, ok, "status %q missing from validTransitions map", s)
	}
}
