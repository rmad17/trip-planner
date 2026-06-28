package adminapi

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"triplanner/accounts"
	"triplanner/core"
	"triplanner/documents"
	"triplanner/expenses"
	"triplanner/trips"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const adminCookieName = "admin_token"

// WebAdminRequired checks the admin_token cookie and redirects to login if absent or invalid.
func WebAdminRequired(c *gin.Context) {
	cookie, err := c.Cookie(adminCookieName)
	if err != nil || cookie == "" {
		c.Redirect(http.StatusFound, "/admin/login?next="+c.Request.URL.Path)
		c.Abort()
		return
	}

	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil || !token.Valid {
		clearAdminCookie(c)
		c.Redirect(http.StatusFound, "/admin/login")
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Redirect(http.StatusFound, "/admin/login")
		c.Abort()
		return
	}

	var user accounts.User
	core.DB.Where("id = ?", claims["id"]).First(&user)
	if user.ID == uuid.Nil || !user.IsAdmin {
		clearAdminCookie(c)
		c.Redirect(http.StatusFound, "/admin/login")
		c.Abort()
		return
	}

	c.Set("currentUser", user)
	c.Next()
}

func WebLoginPage(c *gin.Context) {
	next := c.DefaultQuery("next", "/admin/")
	c.HTML(http.StatusOK, "admin_login.tmpl", gin.H{"Next": next})
}

func WebLoginPost(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	next := c.PostForm("next")
	if next == "" {
		next = "/admin/"
	}

	renderError := func() {
		c.HTML(http.StatusUnauthorized, "admin_login.tmpl", gin.H{
			"Error": "Please enter the correct email address and password for an admin account. Note that both fields are case-sensitive.",
			"Next":  next,
		})
	}

	var user accounts.User
	core.DB.Where("email = ?", email).First(&user)
	if user.ID == uuid.Nil || !user.IsAdmin {
		renderError()
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		renderError()
		return
	}

	tokenStr, err := generateAdminToken(user.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin_login.tmpl", gin.H{
			"Error": "Authentication error. Please try again.",
			"Next":  next,
		})
		return
	}

	isProd := os.Getenv("GIN_MODE") == "release"
	c.SetCookie(adminCookieName, tokenStr, int((24 * time.Hour).Seconds()), "/", "", isProd, true)
	c.Redirect(http.StatusFound, next)
}

func WebLogout(c *gin.Context) {
	clearAdminCookie(c)
	c.Redirect(http.StatusFound, "/admin/login")
}

type webDashboardStats struct {
	Users     int64
	Admins    int64
	Trips     int64
	Expenses  int64
	Documents int64
}

func WebDashboard(c *gin.Context) {
	var stats webDashboardStats
	core.DB.Model(&accounts.User{}).Count(&stats.Users)
	core.DB.Model(&accounts.User{}).Where("is_admin = true").Count(&stats.Admins)
	core.DB.Model(&trips.TripPlan{}).Count(&stats.Trips)
	core.DB.Model(&expenses.Expense{}).Count(&stats.Expenses)
	core.DB.Model(&documents.Document{}).Count(&stats.Documents)

	user := c.MustGet("currentUser").(accounts.User)
	c.HTML(http.StatusOK, "admin_dashboard.tmpl", gin.H{
		"Stats": stats,
		"User":  user,
	})
}

func generateAdminToken(userID uuid.UUID) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return t.SignedString([]byte(os.Getenv("SECRET")))
}

func clearAdminCookie(c *gin.Context) {
	c.SetCookie(adminCookieName, "", -1, "/", "", false, true)
}
