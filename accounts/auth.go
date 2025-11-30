package accounts

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser godoc
// @Summary Register a new user
// @Description Create a new user account with username and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body AuthInput true "User credentials"
// @Success 200 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /auth/signup [post]
func CreateUser(c *gin.Context) {

	var authInput AuthInput

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFound User
	core.DB.Where("username=?", authInput.Username).Find(&userFound)

	// if userFound.if ) {
	//     c.JSON(http.StatusBadRequest, gin.H{"error": "username already used"})
	//     return
	// }

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(authInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := User{
		Username: authInput.Username,
		Password: string(passwordHash),
	}

	core.DB.Create(&user)

	c.JSON(http.StatusOK, gin.H{"data": user})

}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body AuthInput true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Invalid credentials"
// @Router /auth/login [post]
func Login(c *gin.Context) {

	var authInput AuthInput

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFound User
	core.DB.Where("username=?", authInput.Username).Find(&userFound)

	// if userFound.ID == 0 {
	//     c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
	//     return
	// }

	if err := bcrypt.CompareHashAndPassword([]byte(userFound.Password), []byte(authInput.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userFound.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to generate token"})
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Retrieve current authenticated user's profile
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /user/profile [get]
func GetUserProfile(c *gin.Context) {

	user, _ := c.Get("currentUser")

	c.JSON(200, gin.H{
		"user": user,
	})
}

// GoogleOAuthLogin godoc
// @Summary Google OAuth login page
// @Description Render Google OAuth login page
// @Tags authentication
// @Produce html
// @Success 200 {string} string "HTML page"
// @Router /auth/google [get]
func GoogleOAuthLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Google Login",
	})
}

// GoogleOAuthBegin godoc
// @Summary Begin Google OAuth flow
// @Description Start Google OAuth authentication process
// @Tags authentication
// @Produce json
// @Success 302 {string} string "Redirect to Google"
// @Failure 400 {object} map[string]string "Authentication error"
// @Router /auth/google/begin [get]
func GoogleOAuthBegin(c *gin.Context) {
	key := "Secret-session-key" // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30        // 30 days
	isProd := false             // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store
	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// GoogleOAuthCallback godoc
// @Summary Google OAuth callback
// @Description Handle Google OAuth callback and create/login user
// @Tags authentication
// @Produce json
// @Success 200 {object} map[string]interface{} "JWT token and user data"
// @Failure 500 {object} map[string]string "Authentication error"
// @Router /auth/google/callback [get]
func GoogleOAuthCallback(c *gin.Context) {
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Google"})
		return
	}

	// Find or create user in database
	var dbUser User
	result := core.DB.Where("google_id = ?", gothUser.UserID).First(&dbUser)

	expiresAt := gothUser.ExpiresAt.Unix()
	provider := "google"

	if result.Error != nil {
		// User doesn't exist, create new user
		dbUser = User{
			Username:     gothUser.Email, // Use email as username
			Email:        &gothUser.Email,
			GoogleID:     &gothUser.UserID,
			Name:         &gothUser.Name,
			FirstName:    &gothUser.FirstName,
			LastName:     &gothUser.LastName,
			AvatarURL:    &gothUser.AvatarURL,
			Provider:     &provider,
			AccessToken:  &gothUser.AccessToken,
			RefreshToken: &gothUser.RefreshToken,
			ExpiresAt:    &expiresAt,
		}

		if createErr := core.DB.Create(&dbUser).Error; createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// User exists, update their info
		updates := map[string]interface{}{
			"name":          gothUser.Name,
			"first_name":    gothUser.FirstName,
			"last_name":     gothUser.LastName,
			"avatar_url":    gothUser.AvatarURL,
			"access_token":  gothUser.AccessToken,
			"refresh_token": gothUser.RefreshToken,
			"expires_at":    expiresAt,
		}

		if updateErr := core.DB.Model(&dbUser).Updates(updates).Error; updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Generate JWT token
	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  dbUser.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Get frontend URL from environment variable
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Default for development
	}

	// Redirect back to frontend with token
	// Frontend should handle this route and store the token
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, token)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)

	// Alternative: Return JSON if requested via API (for mobile apps, etc.)
	// You can detect this by checking Accept header or a query parameter
	// if c.GetHeader("Accept") == "application/json" {
	//     c.JSON(http.StatusOK, gin.H{
	//         "token": token,
	//         "user": gin.H{
	//             "id":         dbUser.ID,
	//             "username":   dbUser.Username,
	//             "email":      dbUser.Email,
	//             "name":       dbUser.Name,
	//             "avatar_url": dbUser.AvatarURL,
	//         },
	//     })
	// }
}
