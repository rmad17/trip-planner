package adminapi

import "github.com/gin-gonic/gin"

// RouterGroupAdminWeb mounts the browser-facing admin UI at the given group (e.g. /admin).
// Login and logout are public; the dashboard requires a valid admin_token cookie.
func RouterGroupAdminWeb(r *gin.RouterGroup) {
	r.GET("/login", WebLoginPage)
	r.POST("/login", WebLoginPost)

	protected := r.Group("")
	protected.Use(WebAdminRequired)
	protected.GET("/", WebDashboard)
	protected.GET("/logout", WebLogout)
}

// RouterGroupAdmin mounts all admin endpoints.
// The router group must already be protected by accounts.CheckAuth and accounts.AdminRequired.
func RouterGroupAdmin(r *gin.RouterGroup) {
	r.GET("/stats", GetStats)

	users := r.Group("/users")
	{
		users.GET("", ListUsers)
		users.GET("/:id", GetUser)
		users.PUT("/:id", UpdateUser)
		users.DELETE("/:id", DeleteUser)
	}

	tripGroup := r.Group("/trips")
	{
		tripGroup.GET("", ListTrips)
		tripGroup.GET("/:id", GetTrip)
		tripGroup.DELETE("/:id", DeleteTrip)
	}

	expenseGroup := r.Group("/expenses")
	{
		expenseGroup.GET("", ListExpenses)
		expenseGroup.DELETE("/:id", DeleteExpense)
	}

	docGroup := r.Group("/documents")
	{
		docGroup.GET("", ListDocuments)
		docGroup.DELETE("/:id", DeleteDocument)
	}
}
