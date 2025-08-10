package expenses

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

// ExpenseCategory represents the category of an expense
type ExpenseCategory string

const (
	ExpenseCategoryAccommodation ExpenseCategory = "accommodation" // Hotels, lodging
	ExpenseCategoryTransportation ExpenseCategory = "transportation" // Flights, trains, taxis, etc.
	ExpenseCategoryFood          ExpenseCategory = "food"          // Meals, restaurants, groceries
	ExpenseCategoryActivities    ExpenseCategory = "activities"    // Tours, tickets, entertainment
	ExpenseCategoryShoppingGifts ExpenseCategory = "shopping_gifts" // Shopping and souvenirs
	ExpenseCategoryInsurance     ExpenseCategory = "insurance"     // Travel insurance
	ExpenseCategoryVisasFees     ExpenseCategory = "visas_fees"    // Visa processing fees
	ExpenseCategoryMedical       ExpenseCategory = "medical"       // Medical expenses
	ExpenseCategoryCommunication ExpenseCategory = "communication" // Internet, phone, roaming
	ExpenseCategoryMiscellaneous ExpenseCategory = "miscellaneous" // Tips, laundry, etc.
	ExpenseCategoryOther         ExpenseCategory = "other"         // Other expenses with custom notes
)

// SplitMethod represents how an expense is split among travellers
type SplitMethod string

const (
	SplitMethodEqual      SplitMethod = "equal"      // Split equally among all participants
	SplitMethodExact      SplitMethod = "exact"      // Exact amounts specified for each person
	SplitMethodPercentage SplitMethod = "percentage" // Split by percentage
	SplitMethodShares     SplitMethod = "shares"     // Split by shares/units
	SplitMethodPaidBy     SplitMethod = "paid_by"    // Paid entirely by specific person
)

// PaymentMethod represents how the expense was paid
type PaymentMethod string

const (
	PaymentMethodCash        PaymentMethod = "cash"
	PaymentMethodCard        PaymentMethod = "card"
	PaymentMethodDigitalPay  PaymentMethod = "digital_pay" // UPI, PayPal, etc.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCheque      PaymentMethod = "cheque"
	PaymentMethodOther       PaymentMethod = "other"
)

// Expense represents a single expense in a trip
type Expense struct {
	core.BaseModel
	Title           string          `json:"title" gorm:"not null" example:"Dinner at Le Jules Verne" description:"Title/name of the expense"`
	Description     *string         `json:"description" example:"Romantic dinner at Eiffel Tower restaurant" description:"Detailed description"`
	Amount          float64         `json:"amount" gorm:"not null" example:"250.75" description:"Total amount of the expense"`
	Currency        string          `json:"currency" gorm:"type:varchar(10);not null" example:"EUR" description:"Currency code (should match trip currency)"`
	Category        ExpenseCategory `json:"category" gorm:"type:varchar(30);not null" example:"food" description:"Category of the expense"`
	OtherCategory   *string         `json:"other_category" example:"Local transport" description:"Custom category when category is 'other'"`
	Date            time.Time       `json:"date" gorm:"not null" example:"2024-06-02T19:00:00Z" description:"Date and time of the expense"`
	Location        *string         `json:"location" example:"Eiffel Tower, Paris" description:"Location where expense occurred"`
	Vendor          *string         `json:"vendor" example:"Le Jules Verne Restaurant" description:"Merchant/vendor name"`
	PaymentMethod   PaymentMethod   `json:"payment_method" gorm:"type:varchar(20);not null" example:"card" description:"How the payment was made"`
	SplitMethod     SplitMethod     `json:"split_method" gorm:"type:varchar(20);not null;default:'equal'" example:"equal" description:"How the expense is split"`
	ReceiptURL      *string         `json:"receipt_url" example:"https://storage.example.com/receipts/abc123.pdf" description:"URL to receipt/invoice"`
	Notes           *string         `json:"notes" example:"Included 18% service charge" description:"Additional notes"`
	Tags            []string        `json:"tags" gorm:"type:text[]" example:"romantic,special-occasion" description:"Tags for categorization"`
	IsRecurring     bool            `json:"is_recurring" gorm:"default:false" description:"Whether this is a recurring expense"`
	
	// Entity Relationships - expense can be linked to trip, hop, or day
	TripPlan        *uuid.UUID      `json:"trip_plan" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of associated trip plan"`
	TripHop         *uuid.UUID      `json:"trip_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of associated trip hop"`
	TripDay         *uuid.UUID      `json:"trip_day" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of associated trip day"`
	Activity        *uuid.UUID      `json:"activity" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of associated activity"`
	
	// Who paid and splits
	PaidBy          uuid.UUID       `json:"paid_by" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of traveller who paid"`
	CreatedBy       uuid.UUID       `json:"created_by" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of traveller who created this expense record"`
	
	// Related records
	ExpenseSplits   []ExpenseSplit  `json:"expense_splits,omitempty" gorm:"foreignKey:Expense" description:"How this expense is split among travellers"`
}

// ExpenseSplit represents how an expense is divided among travellers
type ExpenseSplit struct {
	core.BaseModel
	Expense       uuid.UUID `json:"expense" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the parent expense"`
	Traveller     uuid.UUID `json:"traveller" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the traveller"`
	Amount        float64   `json:"amount" gorm:"not null" example:"125.38" description:"Amount owed by this traveller"`
	Percentage    *float64  `json:"percentage" example:"50.0" description:"Percentage of total expense (when split by percentage)"`
	Shares        *int      `json:"shares" example:"2" description:"Number of shares (when split by shares)"`
	IsPaid        bool      `json:"is_paid" gorm:"default:false" description:"Whether this traveller has paid their share"`
	PaidAt        *time.Time `json:"paid_at" description:"When this traveller paid their share"`
	Notes         *string   `json:"notes" example:"Paid via UPI" description:"Notes about this split/payment"`
}

// ExpenseSettlement represents settlements between travellers
type ExpenseSettlement struct {
	core.BaseModel
	TripPlan      uuid.UUID  `json:"trip_plan" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the trip plan"`
	FromTraveller uuid.UUID  `json:"from_traveller" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"Traveller who owes money"`
	ToTraveller   uuid.UUID  `json:"to_traveller" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"Traveller who is owed money"`
	Amount        float64    `json:"amount" gorm:"not null" example:"150.50" description:"Amount to be settled"`
	Currency      string     `json:"currency" gorm:"type:varchar(10);not null" example:"USD" description:"Currency of settlement"`
	Status        string     `json:"status" gorm:"type:varchar(20);not null;default:'pending'" example:"pending" description:"Status (pending, paid, cancelled)"`
	SettledAt     *time.Time `json:"settled_at" description:"When the settlement was completed"`
	PaymentMethod *string    `json:"payment_method" example:"bank_transfer" description:"How the settlement was paid"`
	Notes         *string    `json:"notes" example:"Paid via bank transfer on 2024-06-15" description:"Settlement notes"`
}

// GetModels returns all models for Atlas/GORM
func GetModels() []interface{} {
	return []interface{}{
		&Expense{},
		&ExpenseSplit{},
		&ExpenseSettlement{},
	}
}

// GetValidExpenseCategories returns all valid expense categories
func GetValidExpenseCategories() []ExpenseCategory {
	return []ExpenseCategory{
		ExpenseCategoryAccommodation,
		ExpenseCategoryTransportation,
		ExpenseCategoryFood,
		ExpenseCategoryActivities,
		ExpenseCategoryShoppingGifts,
		ExpenseCategoryInsurance,
		ExpenseCategoryVisasFees,
		ExpenseCategoryMedical,
		ExpenseCategoryCommunication,
		ExpenseCategoryMiscellaneous,
		ExpenseCategoryOther,
	}
}

// IsValidExpenseCategory checks if an expense category is valid
func IsValidExpenseCategory(category string) bool {
	for _, validCategory := range GetValidExpenseCategories() {
		if string(validCategory) == category {
			return true
		}
	}
	return false
}

// GetValidSplitMethods returns all valid split methods
func GetValidSplitMethods() []SplitMethod {
	return []SplitMethod{
		SplitMethodEqual,
		SplitMethodExact,
		SplitMethodPercentage,
		SplitMethodShares,
		SplitMethodPaidBy,
	}
}

// IsValidSplitMethod checks if a split method is valid
func IsValidSplitMethod(method string) bool {
	for _, validMethod := range GetValidSplitMethods() {
		if string(validMethod) == method {
			return true
		}
	}
	return false
}

// GetValidPaymentMethods returns all valid payment methods
func GetValidPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCash,
		PaymentMethodCard,
		PaymentMethodDigitalPay,
		PaymentMethodBankTransfer,
		PaymentMethodCheque,
		PaymentMethodOther,
	}
}

// IsValidPaymentMethod checks if a payment method is valid
func IsValidPaymentMethod(method string) bool {
	for _, validMethod := range GetValidPaymentMethods() {
		if string(validMethod) == method {
			return true
		}
	}
	return false
}

// Helper methods for Expense

// CalculateSplitAmounts calculates how much each traveller owes based on split method
func (e *Expense) CalculateSplitAmounts(travellerIDs []uuid.UUID) map[uuid.UUID]float64 {
	splits := make(map[uuid.UUID]float64)
	
	switch e.SplitMethod {
	case SplitMethodEqual:
		amountPerPerson := e.Amount / float64(len(travellerIDs))
		for _, id := range travellerIDs {
			splits[id] = amountPerPerson
		}
	case SplitMethodPaidBy:
		// Only the person who paid owes the amount (no split)
		splits[e.PaidBy] = e.Amount
	// For other methods (exact, percentage, shares), amounts should be set in ExpenseSplit records
	}
	
	return splits
}

// GetTotalSplitAmount returns the sum of all expense splits
func (e *Expense) GetTotalSplitAmount() float64 {
	total := 0.0
	for _, split := range e.ExpenseSplits {
		total += split.Amount
	}
	return total
}

// IsFullyPaid checks if all splits are paid
func (e *Expense) IsFullyPaid() bool {
	for _, split := range e.ExpenseSplits {
		if !split.IsPaid {
			return false
		}
	}
	return len(e.ExpenseSplits) > 0
}

// GetUnpaidAmount returns the total unpaid amount
func (e *Expense) GetUnpaidAmount() float64 {
	unpaid := 0.0
	for _, split := range e.ExpenseSplits {
		if !split.IsPaid {
			unpaid += split.Amount
		}
	}
	return unpaid
}

// Helper methods for ExpenseSettlement

// CalculateNetSettlements calculates net amounts between travellers for a trip
func CalculateNetSettlements(tripID uuid.UUID, expenses []Expense) map[string]float64 {
	// Map of "fromID-toID" -> net amount
	settlements := make(map[string]float64)
	
	for _, expense := range expenses {
		for _, split := range expense.ExpenseSplits {
			if split.Traveller != expense.PaidBy && split.Amount > 0 {
				key := split.Traveller.String() + "-" + expense.PaidBy.String()
				reverseKey := expense.PaidBy.String() + "-" + split.Traveller.String()
				
				if existing, exists := settlements[reverseKey]; exists {
					// Net out against existing reverse settlement
					if existing > split.Amount {
						settlements[reverseKey] = existing - split.Amount
					} else if split.Amount > existing {
						delete(settlements, reverseKey)
						settlements[key] = split.Amount - existing
					} else {
						delete(settlements, reverseKey)
					}
				} else {
					settlements[key] += split.Amount
				}
			}
		}
	}
	
	return settlements
}