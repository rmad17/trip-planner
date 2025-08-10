package expenses

import (
	"time"

	"github.com/google/uuid"
)

// ExpenseCreateRequest represents the request body for creating an expense
type ExpenseCreateRequest struct {
	Title         string          `json:"title" binding:"required"`
	Description   *string         `json:"description"`
	Amount        float64         `json:"amount" binding:"required,gt=0"`
	Currency      string          `json:"currency" binding:"required,len=3"`
	Category      ExpenseCategory `json:"category" binding:"required"`
	OtherCategory *string         `json:"other_category"`
	Date          time.Time       `json:"date" binding:"required"`
	Location      *string         `json:"location"`
	Vendor        *string         `json:"vendor"`
	PaymentMethod PaymentMethod   `json:"payment_method" binding:"required"`
	SplitMethod   SplitMethod     `json:"split_method" binding:"required"`
	ReceiptURL    *string         `json:"receipt_url"`
	Notes         *string         `json:"notes"`
	Tags          []string        `json:"tags"`
	IsRecurring   bool            `json:"is_recurring"`
	
	// Entity Relationships - expense can be linked to trip, hop, day, or activity
	TripHop  *uuid.UUID `json:"trip_hop"`
	TripDay  *uuid.UUID `json:"trip_day"`
	Activity *uuid.UUID `json:"activity"`
	
	// Who paid
	PaidBy uuid.UUID `json:"paid_by" binding:"required"`
	
	// Splits (optional - if not provided, will auto-split based on split_method)
	Splits []ExpenseSplitRequest `json:"splits"`
}

// ExpenseUpdateRequest represents the request body for updating an expense
type ExpenseUpdateRequest struct {
	Title         *string          `json:"title"`
	Description   *string          `json:"description"`
	Amount        *float64         `json:"amount" binding:"omitempty,gt=0"`
	Currency      *string          `json:"currency" binding:"omitempty,len=3"`
	Category      *ExpenseCategory `json:"category"`
	OtherCategory *string          `json:"other_category"`
	Date          *time.Time       `json:"date"`
	Location      *string          `json:"location"`
	Vendor        *string          `json:"vendor"`
	PaymentMethod *PaymentMethod   `json:"payment_method"`
	SplitMethod   *SplitMethod     `json:"split_method"`
	ReceiptURL    *string          `json:"receipt_url"`
	Notes         *string          `json:"notes"`
	Tags          []string         `json:"tags"`
	IsRecurring   *bool            `json:"is_recurring"`
}

// ExpenseSplitRequest represents a split in an expense creation request
type ExpenseSplitRequest struct {
	Traveller  uuid.UUID `json:"traveller" binding:"required"`
	Amount     float64   `json:"amount" binding:"required,gt=0"`
	Percentage *float64  `json:"percentage" binding:"omitempty,gte=0,lte=100"`
	Shares     *int      `json:"shares" binding:"omitempty,gt=0"`
	IsPaid     bool      `json:"is_paid"`
	Notes      *string   `json:"notes"`
}

// ExpenseSplitUpdateRequest represents updating an expense split
type ExpenseSplitUpdateRequest struct {
	Amount     *float64 `json:"amount" binding:"omitempty,gt=0"`
	Percentage *float64 `json:"percentage" binding:"omitempty,gte=0,lte=100"`
	Shares     *int     `json:"shares" binding:"omitempty,gt=0"`
	IsPaid     *bool    `json:"is_paid"`
	PaidAt     *time.Time `json:"paid_at"`
	Notes      *string  `json:"notes"`
}

// ExpenseSettlementRequest represents creating a settlement
type ExpenseSettlementRequest struct {
	FromTraveller uuid.UUID `json:"from_traveller" binding:"required"`
	ToTraveller   uuid.UUID `json:"to_traveller" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
	Currency      string    `json:"currency" binding:"required,len=3"`
	PaymentMethod *string   `json:"payment_method"`
	Notes         *string   `json:"notes"`
}

// ExpenseSummaryResponse represents expense summary for a trip
type ExpenseSummaryResponse struct {
	TripPlan       uuid.UUID            `json:"trip_plan"`
	TotalExpenses  float64              `json:"total_expenses"`
	Currency       string               `json:"currency"`
	ExpenseCount   int64                `json:"expense_count"`
	CategoryTotals map[string]float64   `json:"category_totals"`
	TravellerTotals map[string]TravellerExpenseSummary `json:"traveller_totals"`
	PendingSettlements []SettlementSummary `json:"pending_settlements"`
}

// TravellerExpenseSummary represents expense summary for a traveller
type TravellerExpenseSummary struct {
	TravellerID   uuid.UUID `json:"traveller_id"`
	TravellerName string    `json:"traveller_name"`
	TotalPaid     float64   `json:"total_paid"`
	TotalOwed     float64   `json:"total_owed"`
	Balance       float64   `json:"balance"` // Positive if owed money, negative if owes money
}

// SettlementSummary represents a settlement between two travellers
type SettlementSummary struct {
	FromTraveller     uuid.UUID `json:"from_traveller"`
	FromTravellerName string    `json:"from_traveller_name"`
	ToTraveller       uuid.UUID `json:"to_traveller"`
	ToTravellerName   string    `json:"to_traveller_name"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
}