package expenses

import (
	"net/http"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EXPENSE SPLITS CRUD

// UpdateExpenseSplit godoc
// @Summary Update an expense split
// @Description Update payment status or amount for an expense split
// @Tags expense-splits
// @Accept json
// @Produce json
// @Param id path string true "Expense Split ID"
// @Param split body ExpenseSplitUpdateRequest true "Updated split data"
// @Success 200 {object} ExpenseSplit "Updated expense split"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Expense split not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /expense-splits/{id} [put]
func UpdateExpenseSplit(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var expenseSplit ExpenseSplit
	// Verify access through trip plan
	result := core.DB.Joins("JOIN expenses ON expense_splits.expense = expenses.id").
		Joins("JOIN trip_plans ON expenses.trip_plan = trip_plans.id").
		Joins("JOIN travellers ON expense_splits.traveller = travellers.id").
		Where("expense_splits.id = ? AND (trip_plans.user_id = ? OR travellers.user_id = ?)", 
			id, user.BaseModel.ID, user.BaseModel.ID).
		First(&expenseSplit)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense split not found or access denied"})
		return
	}

	var updateReq ExpenseSplitUpdateRequest
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare update data
	updateData := make(map[string]interface{})
	if updateReq.Amount != nil {
		updateData["amount"] = *updateReq.Amount
	}
	if updateReq.Percentage != nil {
		updateData["percentage"] = updateReq.Percentage
	}
	if updateReq.Shares != nil {
		updateData["shares"] = updateReq.Shares
	}
	if updateReq.IsPaid != nil {
		updateData["is_paid"] = *updateReq.IsPaid
		if *updateReq.IsPaid {
			updateData["paid_at"] = time.Now()
		} else {
			updateData["paid_at"] = nil
		}
	}
	if updateReq.PaidAt != nil {
		updateData["paid_at"] = *updateReq.PaidAt
	}
	if updateReq.Notes != nil {
		updateData["notes"] = updateReq.Notes
	}

	result = core.DB.Model(&expenseSplit).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expense_split": expenseSplit})
}

// MarkSplitPaid godoc
// @Summary Mark an expense split as paid
// @Description Mark a specific expense split as paid by the traveller
// @Tags expense-splits
// @Param id path string true "Expense Split ID"
// @Success 200 {object} ExpenseSplit "Updated expense split"
// @Failure 404 {object} map[string]string "Expense split not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /expense-splits/{id}/mark-paid [post]
func MarkSplitPaid(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var expenseSplit ExpenseSplit
	// Verify that current user is the traveller who owes this split
	result := core.DB.Joins("JOIN travellers ON expense_splits.traveller = travellers.id").
		Where("expense_splits.id = ? AND travellers.user_id = ?", id, user.BaseModel.ID).
		First(&expenseSplit)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense split not found or access denied"})
		return
	}

	// Mark as paid
	now := time.Now()
	result = core.DB.Model(&expenseSplit).Updates(map[string]interface{}{
		"is_paid": true,
		"paid_at": now,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expense_split": expenseSplit})
}

// EXPENSE SETTLEMENTS CRUD

// GetSettlements godoc
// @Summary Get settlements for a trip plan
// @Description Retrieve all settlements for a specific trip plan
// @Tags settlements
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} map[string]interface{} "List of settlements"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip/{id}/settlements [get]
func GetSettlements(c *gin.Context) {
	tripPlanIDStr := c.Param("id")
	
	// Validate UUID format
	tripPlanID, err := uuid.Parse(tripPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID format"})
		return
	}
	
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify access
	var hasAccess bool
	var tripPlan struct{ ID uuid.UUID }
	result := core.DB.Table("trip_plans").Select("id").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	var settlements []ExpenseSettlement
	result = core.DB.Where("trip_plan = ?", tripPlanID).Order("created_at DESC").Find(&settlements)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlements": settlements})
}

// CreateSettlement godoc
// @Summary Create a settlement between travellers
// @Description Record a payment between travellers to settle expenses
// @Tags settlements
// @Accept json
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param settlement body ExpenseSettlementRequest true "Settlement data"
// @Success 201 {object} ExpenseSettlement "Created settlement"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip/{id}/settlements [post]
func CreateSettlement(c *gin.Context) {
	tripPlanIDStr := c.Param("id")
	
	// Validate UUID format
	tripPlanID, err := uuid.Parse(tripPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID format"})
		return
	}
	
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify access (must be trip owner or one of the travellers involved)
	var hasAccess bool
	var tripPlan struct{ ID uuid.UUID }
	result := core.DB.Table("trip_plans").Select("id").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	var settlementReq ExpenseSettlementRequest
	if err := c.BindJSON(&settlementReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that both travellers exist and are part of this trip
	var fromExists, toExists bool
	core.DB.Table("travellers").Where("id = ? AND trip_plan = ? AND is_active = ?", 
		settlementReq.FromTraveller, tripPlanID, true).Select("id").Scan(&fromExists)
	core.DB.Table("travellers").Where("id = ? AND trip_plan = ? AND is_active = ?", 
		settlementReq.ToTraveller, tripPlanID, true).Select("id").Scan(&toExists)

	if !fromExists || !toExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid traveller IDs"})
		return
	}

	if settlementReq.FromTraveller == settlementReq.ToTraveller {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot settle with yourself"})
		return
	}

	// Create settlement
	settlement := ExpenseSettlement{
		TripPlan:      tripPlanID,
		FromTraveller: settlementReq.FromTraveller,
		ToTraveller:   settlementReq.ToTraveller,
		Amount:        settlementReq.Amount,
		Currency:      settlementReq.Currency,
		Status:        "paid",
		SettledAt:     &time.Time{},
		PaymentMethod: settlementReq.PaymentMethod,
		Notes:         settlementReq.Notes,
	}
	now := time.Now()
	settlement.SettledAt = &now

	result = core.DB.Create(&settlement)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"settlement": settlement})
}

// GetExpenseSummary godoc
// @Summary Get expense summary for a trip
// @Description Get comprehensive expense summary including totals, splits, and settlements
// @Tags expenses
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Success 200 {object} ExpenseSummaryResponse "Expense summary"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip/{id}/expense-summary [get]
func GetExpenseSummary(c *gin.Context) {
	tripPlanIDStr := c.Param("id")
	
	// Validate UUID format
	tripPlanID, err := uuid.Parse(tripPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID format"})
		return
	}
	
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify access
	var hasAccess bool
	var tripPlan struct{ ID uuid.UUID; Currency string }
	result := core.DB.Table("trip_plans").Select("id, currency").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
		if hasAccess {
			// Get trip plan currency
			core.DB.Table("trip_plans").Select("id, currency").Where("id = ?", tripPlanID).First(&tripPlan)
		}
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	// Get total expenses and count
	var totalExpenses float64
	var expenseCount int64
	core.DB.Model(&Expense{}).Where("trip_plan = ?", tripPlanID).Count(&expenseCount)
	core.DB.Model(&Expense{}).Where("trip_plan = ?", tripPlanID).Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	// Get category totals
	var categoryResults []struct {
		Category string
		Total    float64
	}
	core.DB.Model(&Expense{}).
		Select("category, SUM(amount) as total").
		Where("trip_plan = ?", tripPlanID).
		Group("category").
		Find(&categoryResults)

	categoryTotals := make(map[string]float64)
	for _, result := range categoryResults {
		categoryTotals[result.Category] = result.Total
	}

	// Get traveller expense summaries
	var travellerResults []struct {
		TravellerID   uuid.UUID
		TravellerName string
		TotalPaid     float64
		TotalOwed     float64
	}
	
	core.DB.Raw(`
		SELECT 
			t.id as traveller_id,
			CONCAT(t.first_name, ' ', t.last_name) as traveller_name,
			COALESCE(paid.total_paid, 0) as total_paid,
			COALESCE(owed.total_owed, 0) as total_owed
		FROM travellers t
		LEFT JOIN (
			SELECT paid_by as traveller_id, SUM(amount) as total_paid
			FROM expenses 
			WHERE trip_plan = ?
			GROUP BY paid_by
		) paid ON t.id = paid.traveller_id
		LEFT JOIN (
			SELECT traveller as traveller_id, SUM(amount) as total_owed
			FROM expense_splits
			WHERE expense IN (SELECT id FROM expenses WHERE trip_plan = ?)
			GROUP BY traveller
		) owed ON t.id = owed.traveller_id
		WHERE t.trip_plan = ? AND t.is_active = true
	`, tripPlanID, tripPlanID, tripPlanID).Find(&travellerResults)

	travellerTotals := make(map[string]TravellerExpenseSummary)
	for _, result := range travellerResults {
		balance := result.TotalPaid - result.TotalOwed
		travellerTotals[result.TravellerID.String()] = TravellerExpenseSummary{
			TravellerID:   result.TravellerID,
			TravellerName: result.TravellerName,
			TotalPaid:     result.TotalPaid,
			TotalOwed:     result.TotalOwed,
			Balance:       balance,
		}
	}

	// Calculate pending settlements (simplified - shows who owes whom)
	var pendingSettlements []SettlementSummary
	for _, traveller := range travellerResults {
		if traveller.TotalOwed > traveller.TotalPaid {
			// This traveller owes money - find who they owe the most to
			amountOwed := traveller.TotalOwed - traveller.TotalPaid
			// For simplicity, we'll show they owe money to the trip organizer or highest payer
			// In a real implementation, this would be more sophisticated
			var topPayer struct {
				TravellerID   uuid.UUID
				TravellerName string
			}
			core.DB.Raw(`
				SELECT t.id as traveller_id, CONCAT(t.first_name, ' ', t.last_name) as traveller_name
				FROM travellers t
				JOIN expenses e ON t.id = e.paid_by
				WHERE e.trip_plan = ? AND t.id != ?
				GROUP BY t.id, t.first_name, t.last_name
				ORDER BY SUM(e.amount) DESC
				LIMIT 1
			`, tripPlanID, traveller.TravellerID).Scan(&topPayer)

			if topPayer.TravellerID != uuid.Nil {
				pendingSettlements = append(pendingSettlements, SettlementSummary{
					FromTraveller:     traveller.TravellerID,
					FromTravellerName: traveller.TravellerName,
					ToTraveller:       topPayer.TravellerID,
					ToTravellerName:   topPayer.TravellerName,
					Amount:            amountOwed,
					Currency:          tripPlan.Currency,
				})
			}
		}
	}

	summary := ExpenseSummaryResponse{
		TripPlan:           tripPlan.ID,
		TotalExpenses:      totalExpenses,
		Currency:           tripPlan.Currency,
		ExpenseCount:       expenseCount,
		CategoryTotals:     categoryTotals,
		TravellerTotals:    travellerTotals,
		PendingSettlements: pendingSettlements,
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}