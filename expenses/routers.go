package expenses

import "github.com/gin-gonic/gin"

// RouterGroupExpenses sets up comprehensive CRUD routes for expenses
func RouterGroupExpenses(router *gin.RouterGroup) {
	// Expenses nested under Trip Plans
	router.GET("/:id/expenses", GetExpenses)              // GET /trip/:id/expenses
	router.POST("/:id/expenses", CreateExpense)           // POST /trip/:id/expenses
	router.GET("/:id/expense-summary", GetExpenseSummary) // GET /trip/:id/expense-summary

	// Settlements nested under Trip Plans
	router.GET("/:id/settlements", GetSettlements)    // GET /trip/:id/settlements
	router.POST("/:id/settlements", CreateSettlement) // POST /trip/:id/settlements
}

// RouterGroupExpenseItems sets up CRUD routes for individual expenses
func RouterGroupExpenseItems(router *gin.RouterGroup) {
	router.GET("/:id", GetExpense)       // GET /expenses/:id
	router.PUT("/:id", UpdateExpense)    // PUT /expenses/:id
	router.DELETE("/:id", DeleteExpense) // DELETE /expenses/:id
}

// RouterGroupExpenseSplits sets up routes for expense splits
func RouterGroupExpenseSplits(router *gin.RouterGroup) {
	router.PUT("/:id", UpdateExpenseSplit)       // PUT /expense-splits/:id
	router.POST("/:id/mark-paid", MarkSplitPaid) // POST /expense-splits/:id/mark-paid
}

