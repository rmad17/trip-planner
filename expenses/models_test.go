package expenses

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupExpensesTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&Expense{}, &ExpenseSplit{}, &ExpenseSettlement{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func StringPtr(s string) *string {
	return &s
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

func TestExpenseCategory_Constants(t *testing.T) {
	tests := []struct {
		name     string
		category ExpenseCategory
		expected string
	}{
		{"Accommodation", ExpenseCategoryAccommodation, "accommodation"},
		{"Transportation", ExpenseCategoryTransportation, "transportation"},
		{"Food", ExpenseCategoryFood, "food"},
		{"Activities", ExpenseCategoryActivities, "activities"},
		{"Shopping/Gifts", ExpenseCategoryShoppingGifts, "shopping_gifts"},
		{"Insurance", ExpenseCategoryInsurance, "insurance"},
		{"Visas/Fees", ExpenseCategoryVisasFees, "visas_fees"},
		{"Medical", ExpenseCategoryMedical, "medical"},
		{"Communication", ExpenseCategoryCommunication, "communication"},
		{"Miscellaneous", ExpenseCategoryMiscellaneous, "miscellaneous"},
		{"Other", ExpenseCategoryOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.category))
		})
	}
}

func TestSplitMethod_Constants(t *testing.T) {
	tests := []struct {
		name     string
		method   SplitMethod
		expected string
	}{
		{"Equal", SplitMethodEqual, "equal"},
		{"Exact", SplitMethodExact, "exact"},
		{"Percentage", SplitMethodPercentage, "percentage"},
		{"Shares", SplitMethodShares, "shares"},
		{"Paid By", SplitMethodPaidBy, "paid_by"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.method))
		})
	}
}

func TestPaymentMethod_Constants(t *testing.T) {
	tests := []struct {
		name     string
		method   PaymentMethod
		expected string
	}{
		{"Cash", PaymentMethodCash, "cash"},
		{"Card", PaymentMethodCard, "card"},
		{"Digital Pay", PaymentMethodDigitalPay, "digital_pay"},
		{"Bank Transfer", PaymentMethodBankTransfer, "bank_transfer"},
		{"Cheque", PaymentMethodCheque, "cheque"},
		{"Other", PaymentMethodOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.method))
		})
	}
}

func TestIsValidExpenseCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected bool
	}{
		{"Valid - accommodation", "accommodation", true},
		{"Valid - food", "food", true},
		{"Valid - other", "other", true},
		{"Invalid - random", "random", false},
		{"Invalid - empty", "", false},
		{"Invalid - uppercase", "FOOD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidExpenseCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidSplitMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"Valid - equal", "equal", true},
		{"Valid - exact", "exact", true},
		{"Valid - percentage", "percentage", true},
		{"Invalid - invalid", "invalid", false},
		{"Invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSplitMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidPaymentMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"Valid - cash", "cash", true},
		{"Valid - card", "card", true},
		{"Valid - digital_pay", "digital_pay", true},
		{"Invalid - bitcoin", "bitcoin", false},
		{"Invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPaymentMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidExpenseCategories(t *testing.T) {
	categories := GetValidExpenseCategories()

	t.Run("Returns all categories", func(t *testing.T) {
		assert.Len(t, categories, 11)
	})

	t.Run("Contains expected categories", func(t *testing.T) {
		assert.Contains(t, categories, ExpenseCategoryFood)
		assert.Contains(t, categories, ExpenseCategoryAccommodation)
		assert.Contains(t, categories, ExpenseCategoryTransportation)
		assert.Contains(t, categories, ExpenseCategoryOther)
	})
}

func TestGetValidSplitMethods(t *testing.T) {
	methods := GetValidSplitMethods()

	t.Run("Returns all split methods", func(t *testing.T) {
		assert.Len(t, methods, 5)
	})

	t.Run("Contains expected methods", func(t *testing.T) {
		assert.Contains(t, methods, SplitMethodEqual)
		assert.Contains(t, methods, SplitMethodExact)
		assert.Contains(t, methods, SplitMethodPercentage)
	})
}

func TestGetValidPaymentMethods(t *testing.T) {
	methods := GetValidPaymentMethods()

	t.Run("Returns all payment methods", func(t *testing.T) {
		assert.Len(t, methods, 6)
	})

	t.Run("Contains expected methods", func(t *testing.T) {
		assert.Contains(t, methods, PaymentMethodCash)
		assert.Contains(t, methods, PaymentMethodCard)
		assert.Contains(t, methods, PaymentMethodDigitalPay)
	})
}

func TestExpense_Model(t *testing.T) {
	db := setupExpensesTestDB(t)

	t.Run("Create expense with required fields", func(t *testing.T) {
		tripID := uuid.New()
		paidByID := uuid.New()
		createdByID := uuid.New()

		expense := Expense{
			Title:         "Dinner",
			Amount:        150.50,
			Currency:      "USD",
			Category:      ExpenseCategoryFood,
			Date:          time.Now(),
			PaymentMethod: PaymentMethodCard,
			SplitMethod:   SplitMethodEqual,
			TripPlan:      &tripID,
			PaidBy:        paidByID,
			CreatedBy:     createdByID,
		}

		result := db.Create(&expense)
		assert.NoError(t, result.Error)
		assert.NotEqual(t, uuid.Nil, expense.ID)
		assert.Equal(t, "Dinner", expense.Title)
		assert.Equal(t, 150.50, expense.Amount)
	})

	t.Run("Create expense with all optional fields", func(t *testing.T) {
		tripID := uuid.New()
		activityID := uuid.New()
		paidByID := uuid.New()
		createdByID := uuid.New()

		expense := Expense{
			Title:         "Romantic Dinner",
			Description:   StringPtr("Dinner at Eiffel Tower restaurant"),
			Amount:        250.75,
			Currency:      "EUR",
			Category:      ExpenseCategoryFood,
			OtherCategory: StringPtr("Fine Dining"),
			Date:          time.Now(),
			Location:      StringPtr("Eiffel Tower, Paris"),
			Vendor:        StringPtr("Le Jules Verne"),
			PaymentMethod: PaymentMethodCard,
			SplitMethod:   SplitMethodEqual,
			ReceiptURL:    StringPtr("https://example.com/receipt.pdf"),
			Notes:         StringPtr("Includes 18% service charge"),
			Tags:          []string{"romantic", "special-occasion"},
			IsRecurring:   false,
			TripPlan:      &tripID,
			Activity:      &activityID,
			PaidBy:        paidByID,
			CreatedBy:     createdByID,
		}

		result := db.Create(&expense)
		assert.NoError(t, result.Error)
		assert.NotNil(t, expense.Description)
		assert.Equal(t, "Dinner at Eiffel Tower restaurant", *expense.Description)
		assert.NotNil(t, expense.Location)
		assert.Equal(t, 2, len(expense.Tags))
	})
}

func TestExpense_CalculateSplitAmounts(t *testing.T) {
	tests := []struct {
		name         string
		expense      Expense
		travellerIDs []uuid.UUID
		expected     map[uuid.UUID]float64
	}{
		{
			name: "Equal split among 3 travellers",
			expense: Expense{
				Amount:      300.00,
				SplitMethod: SplitMethodEqual,
			},
			travellerIDs: []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
			expected:     nil, // Will verify each gets 100
		},
		{
			name: "Equal split among 2 travellers",
			expense: Expense{
				Amount:      100.00,
				SplitMethod: SplitMethodEqual,
			},
			travellerIDs: []uuid.UUID{uuid.New(), uuid.New()},
			expected:     nil, // Will verify each gets 50
		},
		{
			name: "Paid by single person",
			expense: Expense{
				Amount:      200.00,
				SplitMethod: SplitMethodPaidBy,
				PaidBy:      uuid.New(),
			},
			travellerIDs: []uuid.UUID{uuid.New(), uuid.New()},
			expected:     nil, // Will verify only paidBy gets full amount
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splits := tt.expense.CalculateSplitAmounts(tt.travellerIDs)

			switch tt.expense.SplitMethod {
			case SplitMethodEqual:
				expectedAmount := tt.expense.Amount / float64(len(tt.travellerIDs))
				for _, id := range tt.travellerIDs {
					assert.Equal(t, expectedAmount, splits[id])
				}

			case SplitMethodPaidBy:
				assert.Len(t, splits, 1)
				assert.Equal(t, tt.expense.Amount, splits[tt.expense.PaidBy])
			}
		})
	}
}

func TestExpense_GetTotalSplitAmount(t *testing.T) {
	expense := Expense{
		Amount: 300.00,
		ExpenseSplits: []ExpenseSplit{
			{Amount: 100.00},
			{Amount: 100.00},
			{Amount: 100.00},
		},
	}

	total := expense.GetTotalSplitAmount()
	assert.Equal(t, 300.00, total)
}

func TestExpense_IsFullyPaid(t *testing.T) {
	tests := []struct {
		name     string
		expense  Expense
		expected bool
	}{
		{
			name: "All splits paid",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{
					{Amount: 100.00, IsPaid: true},
					{Amount: 100.00, IsPaid: true},
				},
			},
			expected: true,
		},
		{
			name: "Some splits unpaid",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{
					{Amount: 100.00, IsPaid: true},
					{Amount: 100.00, IsPaid: false},
				},
			},
			expected: false,
		},
		{
			name: "No splits",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expense.IsFullyPaid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpense_GetUnpaidAmount(t *testing.T) {
	tests := []struct {
		name     string
		expense  Expense
		expected float64
	}{
		{
			name: "All paid",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{
					{Amount: 100.00, IsPaid: true},
					{Amount: 100.00, IsPaid: true},
				},
			},
			expected: 0.0,
		},
		{
			name: "Partially paid",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{
					{Amount: 100.00, IsPaid: true},
					{Amount: 150.00, IsPaid: false},
					{Amount: 50.00, IsPaid: false},
				},
			},
			expected: 200.0,
		},
		{
			name: "None paid",
			expense: Expense{
				ExpenseSplits: []ExpenseSplit{
					{Amount: 100.00, IsPaid: false},
					{Amount: 200.00, IsPaid: false},
				},
			},
			expected: 300.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expense.GetUnpaidAmount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpenseSplit_Model(t *testing.T) {
	db := setupExpensesTestDB(t)

	t.Run("Create expense split", func(t *testing.T) {
		expenseID := uuid.New()
		travellerID := uuid.New()

		split := ExpenseSplit{
			Expense:   expenseID,
			Traveller: travellerID,
			Amount:    125.50,
			IsPaid:    false,
		}

		result := db.Create(&split)
		assert.NoError(t, result.Error)
		assert.NotEqual(t, uuid.Nil, split.ID)
		assert.Equal(t, 125.50, split.Amount)
		assert.False(t, split.IsPaid)
	})

	t.Run("Create expense split with percentage", func(t *testing.T) {
		expenseID := uuid.New()
		travellerID := uuid.New()
		percentage := 50.0

		split := ExpenseSplit{
			Expense:    expenseID,
			Traveller:  travellerID,
			Amount:     100.00,
			Percentage: &percentage,
			IsPaid:     false,
		}

		result := db.Create(&split)
		assert.NoError(t, result.Error)
		assert.NotNil(t, split.Percentage)
		assert.Equal(t, 50.0, *split.Percentage)
	})

	t.Run("Create expense split with shares", func(t *testing.T) {
		expenseID := uuid.New()
		travellerID := uuid.New()
		shares := 2

		split := ExpenseSplit{
			Expense:   expenseID,
			Traveller: travellerID,
			Amount:    100.00,
			Shares:    &shares,
			IsPaid:    false,
		}

		result := db.Create(&split)
		assert.NoError(t, result.Error)
		assert.NotNil(t, split.Shares)
		assert.Equal(t, 2, *split.Shares)
	})

	t.Run("Mark split as paid", func(t *testing.T) {
		expenseID := uuid.New()
		travellerID := uuid.New()

		split := ExpenseSplit{
			Expense:   expenseID,
			Traveller: travellerID,
			Amount:    50.00,
			IsPaid:    false,
		}
		db.Create(&split)

		// Mark as paid
		paidTime := time.Now()
		split.IsPaid = true
		split.PaidAt = &paidTime
		db.Save(&split)

		var updated ExpenseSplit
		db.First(&updated, split.ID)
		assert.True(t, updated.IsPaid)
		assert.NotNil(t, updated.PaidAt)
	})
}

func TestExpenseSettlement_Model(t *testing.T) {
	db := setupExpensesTestDB(t)

	t.Run("Create settlement", func(t *testing.T) {
		tripID := uuid.New()
		fromID := uuid.New()
		toID := uuid.New()

		settlement := ExpenseSettlement{
			TripPlan:      tripID,
			FromTraveller: fromID,
			ToTraveller:   toID,
			Amount:        150.00,
			Currency:      "USD",
			Status:        "pending",
		}

		result := db.Create(&settlement)
		assert.NoError(t, result.Error)
		assert.NotEqual(t, uuid.Nil, settlement.ID)
		assert.Equal(t, 150.00, settlement.Amount)
		assert.Equal(t, "pending", settlement.Status)
	})

	t.Run("Mark settlement as paid", func(t *testing.T) {
		tripID := uuid.New()
		fromID := uuid.New()
		toID := uuid.New()

		settlement := ExpenseSettlement{
			TripPlan:      tripID,
			FromTraveller: fromID,
			ToTraveller:   toID,
			Amount:        100.00,
			Currency:      "USD",
			Status:        "pending",
		}
		db.Create(&settlement)

		// Mark as settled
		settledTime := time.Now()
		paymentMethod := "bank_transfer"
		notes := "Paid via bank transfer"

		settlement.Status = "paid"
		settlement.SettledAt = &settledTime
		settlement.PaymentMethod = &paymentMethod
		settlement.Notes = &notes
		db.Save(&settlement)

		var updated ExpenseSettlement
		db.First(&updated, settlement.ID)
		assert.Equal(t, "paid", updated.Status)
		assert.NotNil(t, updated.SettledAt)
		assert.NotNil(t, updated.PaymentMethod)
		assert.Equal(t, "bank_transfer", *updated.PaymentMethod)
	})
}

func TestCalculateNetSettlements(t *testing.T) {
	tripID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	tests := []struct {
		name      string
		expenses  []Expense
		assertion func(*testing.T, map[string]float64)
	}{
		{
			name: "Simple two-person settlement",
			expenses: []Expense{
				{
					Amount: 100.00,
					PaidBy: user1,
					ExpenseSplits: []ExpenseSplit{
						{Traveller: user1, Amount: 50.00, IsPaid: false},
						{Traveller: user2, Amount: 50.00, IsPaid: false},
					},
				},
			},
			assertion: func(t *testing.T, settlements map[string]float64) {
				key := user2.String() + "-" + user1.String()
				assert.Equal(t, 50.00, settlements[key])
			},
		},
		{
			name: "Netting out settlements",
			expenses: []Expense{
				{
					Amount: 100.00,
					PaidBy: user1,
					ExpenseSplits: []ExpenseSplit{
						{Traveller: user1, Amount: 50.00},
						{Traveller: user2, Amount: 50.00},
					},
				},
				{
					Amount: 80.00,
					PaidBy: user2,
					ExpenseSplits: []ExpenseSplit{
						{Traveller: user1, Amount: 40.00},
						{Traveller: user2, Amount: 40.00},
					},
				},
			},
			assertion: func(t *testing.T, settlements map[string]float64) {
				// user2 owes user1: 50
				// user1 owes user2: 40
				// Net: user2 owes user1: 10
				key := user2.String() + "-" + user1.String()
				assert.Equal(t, 10.00, settlements[key])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settlements := CalculateNetSettlements(tripID, tt.expenses)
			tt.assertion(t, settlements)
		})
	}
}

func TestGetModels_Expenses(t *testing.T) {
	models := GetModels()

	t.Run("Returns correct number of models", func(t *testing.T) {
		assert.Len(t, models, 3)
	})

	t.Run("Contains all expense models", func(t *testing.T) {
		types := make(map[string]bool)
		for _, model := range models {
			switch model.(type) {
			case *Expense:
				types["Expense"] = true
			case *ExpenseSplit:
				types["ExpenseSplit"] = true
			case *ExpenseSettlement:
				types["ExpenseSettlement"] = true
			}
		}

		assert.True(t, types["Expense"])
		assert.True(t, types["ExpenseSplit"])
		assert.True(t, types["ExpenseSettlement"])
	})
}
