package accounts

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser godoc
// @Summary Register a new user
// @Description Create a new user account with email, password, first name, and last name
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body RegisterInput true "Registration details"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 409 {object} map[string]string "Email already registered"
// @Router /auth/signup [post]
func CreateUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing User
	core.DB.Where("email = ?", input.Email).First(&existing)
	if existing.ID != uuid.Nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	emailPrefix := []rune(input.Email)
	if len(emailPrefix) > 2 {
		emailPrefix = emailPrefix[:2]
	}
	defaultName := string(emailPrefix)
	if input.FirstName == "" {
		input.FirstName = defaultName
	}
	if input.LastName == "" {
		input.LastName = defaultName
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := User{
		Username:  input.Email,
		Email:     &input.Email,
		Password:  string(passwordHash),
		FirstName: &input.FirstName,
		LastName:  &input.LastName,
	}

	if err := core.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
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
	core.DB.Where("email = ?", authInput.Email).First(&userFound)

	if userFound.ID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFound.Password), []byte(authInput.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
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

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the current user's display name and related fields
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Updated user profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /user/profile [put]
func UpdateUserProfile(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(User)

	var req struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FirstName != nil {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = req.LastName
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := core.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the refreshed user
	core.DB.First(&user, "id = ?", user.ID)
	c.JSON(http.StatusOK, gin.H{"user": user})
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
	sessionKey := os.Getenv("SECRET")
	if sessionKey == "" {
		sessionKey = "secret-session-key"
	}
	isProd := os.Getenv("GIN_MODE") == "release"

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store
	q := c.Request.URL.Query()
	q.Set("provider", c.Param("provider"))
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
	q := c.Request.URL.Query()
	q.Set("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Google"})
		return
	}

	var dbUser User
	expiresAt := gothUser.ExpiresAt.Unix()
	provider := "google"

	googleUpdates := map[string]interface{}{
		"google_id":     gothUser.UserID,
		"name":          gothUser.Name,
		"first_name":    gothUser.FirstName,
		"last_name":     gothUser.LastName,
		"avatar_url":    gothUser.AvatarURL,
		"provider":      provider,
		"access_token":  gothUser.AccessToken,
		"refresh_token": gothUser.RefreshToken,
		"expires_at":    expiresAt,
		"email_verified": true,
	}

	byGoogleID := core.DB.Where("google_id = ?", gothUser.UserID).First(&dbUser)
	if byGoogleID.Error == nil {
		// Known Google user — refresh tokens and profile
		if updateErr := core.DB.Model(&dbUser).Updates(googleUpdates).Error; updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	} else {
		// Check if an email-based account already exists with this email
		byEmail := core.DB.Where("email = ?", gothUser.Email).First(&dbUser)
		if byEmail.Error == nil {
			// Link Google account to existing email user
			if updateErr := core.DB.Model(&dbUser).Updates(googleUpdates).Error; updateErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Google account"})
				return
			}
		} else {
			// New user — create account
			dbUser = User{
				Username:      gothUser.Email,
				Email:         &gothUser.Email,
				GoogleID:      &gothUser.UserID,
				Name:          &gothUser.Name,
				FirstName:     &gothUser.FirstName,
				LastName:      &gothUser.LastName,
				AvatarURL:     &gothUser.AvatarURL,
				Provider:      &provider,
				AccessToken:   &gothUser.AccessToken,
				RefreshToken:  &gothUser.RefreshToken,
				ExpiresAt:     &expiresAt,
				EmailVerified: true,
			}
			if createErr := core.DB.Create(&dbUser).Error; createErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
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
