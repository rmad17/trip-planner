package expenses

import (
	"net/http"
	"strconv"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EXPENSES CRUD

// GetExpenses godoc
// @Summary Get expenses for a trip plan
// @Description Retrieve all expenses for a specific trip plan with optional filtering
// @Tags expenses
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param category query string false "Filter by category"
// @Param traveller query string false "Filter by traveller who paid"
// @Param limit query int false "Number of records to return (default: 50)"
// @Param offset query int false "Number of records to skip (default: 0)"
// @Success 200 {object} map[string]interface{} "List of expenses"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/expenses [get]
func GetExpenses(c *gin.Context) {
	tripPlanID := c.Param("trip_plan_id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan access (user owns trip or is a traveller)
	var hasAccess bool
	// Check if user owns the trip
	var tripPlan struct{ ID uuid.UUID }
	result := core.DB.Table("trip_plans").Select("id").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		// Check if user is a traveller in this trip
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	// Get query parameters
	category := c.Query("category")
	travellerFilter := c.Query("traveller")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Build query
	query := core.DB.Preload("ExpenseSplits").Where("trip_plan = ?", tripPlanID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if travellerFilter != "" {
		query = query.Where("paid_by = ?", travellerFilter)
	}

	var expenses []Expense
	var count int64

	// Get total count
	core.DB.Model(&Expense{}).Where("trip_plan = ?", tripPlanID).Count(&count)

	// Get expenses
	result = query.Order("date DESC").Limit(limit).Offset(offset).Find(&expenses)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expenses": expenses,
		"total":    count,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetExpense godoc
// @Summary Get a specific expense
// @Description Retrieve an expense by ID with all splits
// @Tags expenses
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} Expense "Expense details"
// @Failure 404 {object} map[string]string "Expense not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /expenses/{id} [get]
func GetExpense(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var expense Expense
	// Verify access through trip plan ownership or traveller status
	result := core.DB.Preload("ExpenseSplits").
		Joins("JOIN trip_plans ON expenses.trip_plan = trip_plans.id").
		Where("expenses.id = ? AND (trip_plans.user_id = ? OR EXISTS (SELECT 1 FROM travellers WHERE trip_plan = trip_plans.id AND user_id = ? AND is_active = true))",
			id, user.BaseModel.ID, user.BaseModel.ID).
		First(&expense)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expense": expense})
}

// CreateExpense godoc
// @Summary Create a new expense
// @Description Add a new expense to a trip plan
// @Tags expenses
// @Accept json
// @Produce json
// @Param trip_plan_id path string true "Trip Plan ID"
// @Param expense body ExpenseCreateRequest true "Expense data"
// @Success 201 {object} Expense "Created expense"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip-plans/{trip_plan_id}/expenses [post]
func CreateExpense(c *gin.Context) {
	tripPlanID := c.Param("trip_plan_id")
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

	var expenseReq ExpenseCreateRequest
	if err := c.BindJSON(&expenseReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the traveller ID for the current user in this trip
	var currentTraveller struct {
		ID uuid.UUID
	}
	result = core.DB.Table("travellers").Select("id").
		Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).
		First(&currentTraveller)

	var createdByTraveller uuid.UUID
	if result.Error == nil {
		createdByTraveller = currentTraveller.ID
	} else {
		// If not a traveller, use the paid_by traveller
		createdByTraveller = expenseReq.PaidBy
	}

	// Create expense
	expense := Expense{
		Title:         expenseReq.Title,
		Description:   expenseReq.Description,
		Amount:        expenseReq.Amount,
		Currency:      expenseReq.Currency,
		Category:      expenseReq.Category,
		OtherCategory: expenseReq.OtherCategory,
		Date:          expenseReq.Date,
		Location:      expenseReq.Location,
		Vendor:        expenseReq.Vendor,
		PaymentMethod: expenseReq.PaymentMethod,
		SplitMethod:   expenseReq.SplitMethod,
		ReceiptURL:    expenseReq.ReceiptURL,
		Notes:         expenseReq.Notes,
		Tags:          expenseReq.Tags,
		IsRecurring:   expenseReq.IsRecurring,
		// TripPlan:      (*uuid.UUID)(&uuid.MustParse(tripPlanID)),
		TripHop:   expenseReq.TripHop,
		TripDay:   expenseReq.TripDay,
		Activity:  expenseReq.Activity,
		PaidBy:    expenseReq.PaidBy,
		CreatedBy: createdByTraveller,
	}

	// Start transaction
	tx := core.DB.Begin()

	// Create expense
	result = tx.Create(&expense)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Create expense splits
	if len(expenseReq.Splits) > 0 {
		for _, splitReq := range expenseReq.Splits {
			split := ExpenseSplit{
				Expense:    expense.BaseModel.ID,
				Traveller:  splitReq.Traveller,
				Amount:     splitReq.Amount,
				Percentage: splitReq.Percentage,
				Shares:     splitReq.Shares,
				IsPaid:     splitReq.IsPaid,
				Notes:      splitReq.Notes,
			}
			if result := tx.Create(&split); result.Error != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}
	} else {
		// Auto-create equal splits for all active travellers if no splits provided
		if expenseReq.SplitMethod == SplitMethodEqual {
			var travellers []struct {
				ID uuid.UUID
			}
			tx.Table("travellers").Select("id").Where("trip_plan = ? AND is_active = ?", tripPlanID, true).Find(&travellers)

			amountPerPerson := expense.Amount / float64(len(travellers))
			for _, traveller := range travellers {
				split := ExpenseSplit{
					Expense:   expense.BaseModel.ID,
					Traveller: traveller.ID,
					Amount:    amountPerPerson,
					IsPaid:    traveller.ID == expense.PaidBy, // Mark as paid if they paid
				}
				if result := tx.Create(&split); result.Error != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
					return
				}
			}
		}
	}

	tx.Commit()

	// Reload expense with splits
	core.DB.Preload("ExpenseSplits").First(&expense, expense.BaseModel.ID)

	c.JSON(http.StatusCreated, gin.H{"expense": expense})
}

// UpdateExpense godoc
// @Summary Update an expense
// @Description Update an existing expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Param expense body ExpenseUpdateRequest true "Updated expense data"
// @Success 200 {object} Expense "Updated expense"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Expense not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /expenses/{id} [put]
func UpdateExpense(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var expense Expense
	// Verify access - user must be trip owner or the one who created the expense
	result := core.DB.Joins("JOIN trip_plans ON expenses.trip_plan = trip_plans.id").
		Joins("JOIN travellers ON expenses.created_by = travellers.id").
		Where("expenses.id = ? AND (trip_plans.user_id = ? OR travellers.user_id = ?)",
			id, user.BaseModel.ID, user.BaseModel.ID).
		First(&expense)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found or access denied"})
		return
	}

	var updateReq ExpenseUpdateRequest
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update expense fields
	result = core.DB.Model(&expense).Updates(map[string]interface{}{
		"title":          updateReq.Title,
		"description":    updateReq.Description,
		"amount":         updateReq.Amount,
		"currency":       updateReq.Currency,
		"category":       updateReq.Category,
		"other_category": updateReq.OtherCategory,
		"date":           updateReq.Date,
		"location":       updateReq.Location,
		"vendor":         updateReq.Vendor,
		"payment_method": updateReq.PaymentMethod,
		"split_method":   updateReq.SplitMethod,
		"receipt_url":    updateReq.ReceiptURL,
		"notes":          updateReq.Notes,
		"tags":           updateReq.Tags,
		"is_recurring":   updateReq.IsRecurring,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Reload with splits
	core.DB.Preload("ExpenseSplits").First(&expense, expense.BaseModel.ID)

	c.JSON(http.StatusOK, gin.H{"expense": expense})
}

// DeleteExpense godoc
// @Summary Delete an expense
// @Description Delete an expense and all associated splits
// @Tags expenses
// @Param id path string true "Expense ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Expense not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /expenses/{id} [delete]
func DeleteExpense(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var expense Expense
	// Verify access - user must be trip owner or the one who created the expense
	result := core.DB.Joins("JOIN trip_plans ON expenses.trip_plan = trip_plans.id").
		Joins("JOIN travellers ON expenses.created_by = travellers.id").
		Where("expenses.id = ? AND (trip_plans.user_id = ? OR travellers.user_id = ?)",
			id, user.BaseModel.ID, user.BaseModel.ID).
		First(&expense)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found or access denied"})
		return
	}

	// Delete splits first, then expense
	tx := core.DB.Begin()
	tx.Where("expense = ?", id).Delete(&ExpenseSplit{})
	tx.Delete(&expense)

	if tx.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

