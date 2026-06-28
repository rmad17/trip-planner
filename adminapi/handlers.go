package adminapi

import (
	"net/http"
	"strconv"
	"triplanner/accounts"
	"triplanner/core"
	"triplanner/documents"
	"triplanner/expenses"
	"triplanner/trips"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// pagination helpers

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	return page, limit
}

func offset(page, limit int) int {
	return (page - 1) * limit
}

// --- Dashboard Stats ---

func GetStats(c *gin.Context) {
	var (
		userCount     int64
		adminCount    int64
		tripCount     int64
		expenseCount  int64
		documentCount int64
	)

	core.DB.Model(&accounts.User{}).Count(&userCount)
	core.DB.Model(&accounts.User{}).Where("is_admin = true").Count(&adminCount)
	core.DB.Model(&trips.TripPlan{}).Count(&tripCount)
	core.DB.Model(&expenses.Expense{}).Count(&expenseCount)
	core.DB.Model(&documents.Document{}).Count(&documentCount)

	c.JSON(http.StatusOK, gin.H{
		"users":     userCount,
		"admins":    adminCount,
		"trips":     tripCount,
		"expenses":  expenseCount,
		"documents": documentCount,
	})
}

// --- Users ---

func ListUsers(c *gin.Context) {
	page, limit := pageParams(c)
	search := c.Query("search")
	isAdmin := c.Query("is_admin")

	var users []accounts.User
	var total int64

	q := core.DB.Model(&accounts.User{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("username ILIKE ? OR email ILIKE ?", like, like)
	}
	if isAdmin == "true" {
		q = q.Where("is_admin = true")
	} else if isAdmin == "false" {
		q = q.Where("is_admin = false")
	}

	q.Count(&total)
	q.Omit("password", "access_token", "refresh_token", "email_verification_token", "password_reset_token").
		Order("created_at DESC").
		Limit(limit).Offset(offset(page, limit)).
		Find(&users)

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "results": users})
}

func GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var user accounts.User
	if result := core.DB.Omit("password", "access_token", "refresh_token").First(&user, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		IsAdmin       *bool   `json:"is_admin"`
		EmailVerified *bool   `json:"email_verified"`
		Username      *string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.EmailVerified != nil {
		updates["email_verified"] = *req.EmailVerified
	}
	if req.Username != nil && *req.Username != "" {
		updates["username"] = *req.Username
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := core.DB.Model(&accounts.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var user accounts.User
	core.DB.Omit("password", "access_token", "refresh_token").First(&user, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Prevent deleting yourself
	me := c.MustGet("currentUser").(accounts.User)
	if me.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}

	if err := core.DB.Delete(&accounts.User{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// --- Trips ---

func ListTrips(c *gin.Context) {
	page, limit := pageParams(c)
	search := c.Query("search")
	status := c.Query("status")
	userID := c.Query("user_id")

	var tripList []trips.TripPlan
	var total int64

	q := core.DB.Model(&trips.TripPlan{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}

	q.Count(&total)
	q.Order("created_at DESC").Limit(limit).Offset(offset(page, limit)).Find(&tripList)

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "results": tripList})
}

func GetTrip(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	var trip trips.TripPlan
	if result := core.DB.Preload("TripHops").Preload("TripDays").First(&trip, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trip": trip})
}

func DeleteTrip(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	if err := core.DB.Delete(&trips.TripPlan{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "trip deleted"})
}

// --- Expenses ---

func ListExpenses(c *gin.Context) {
	page, limit := pageParams(c)
	tripID := c.Query("trip_id")
	category := c.Query("category")

	var expenseList []expenses.Expense
	var total int64

	q := core.DB.Model(&expenses.Expense{})
	if tripID != "" {
		q = q.Where("trip_plan_id = ?", tripID)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	q.Count(&total)
	q.Order("created_at DESC").Limit(limit).Offset(offset(page, limit)).Find(&expenseList)

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "results": expenseList})
}

func DeleteExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	if err := core.DB.Delete(&expenses.Expense{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "expense deleted"})
}

// --- Documents ---

func ListDocuments(c *gin.Context) {
	page, limit := pageParams(c)
	userID := c.Query("user_id")
	category := c.Query("category")

	var docList []documents.Document
	var total int64

	q := core.DB.Model(&documents.Document{})
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	q.Count(&total)
	q.Order("created_at DESC").Limit(limit).Offset(offset(page, limit)).Find(&docList)

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "results": docList})
}

func DeleteDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	if err := core.DB.Delete(&documents.Document{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "document deleted"})
}
