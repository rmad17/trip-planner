package accounts

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func RouterGroupUserAuth(router *gin.RouterGroup) {
	router.POST("/signup", CreateUser)
	router.POST("/login", Login)
}

func RouterGroupUserProfile(router *gin.RouterGroup) {
	router.GET("/profile", GetUserProfile)
	router.PUT("/profile", UpdateUserProfile)
}

// RouterGroupPasswordReset sets up password reset and email verification routes (no auth required)
func RouterGroupPasswordReset(router *gin.RouterGroup) {
	router.POST("/password-reset/request", RequestPasswordReset)   // POST /auth/password-reset/request
	router.POST("/password-reset/confirm", ResetPassword)          // POST /auth/password-reset/confirm
	router.GET("/verify-email", VerifyEmail)                       // GET /auth/verify-email?token=...
	router.POST("/verify-email/resend", ResendVerificationEmail)   // POST /auth/verify-email/resend
}

func RouterGroupGoogleOAuth(router *gin.RouterGroup) {
	google_client_id := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	callback_url := os.Getenv("GOOGLE_OAUTH_CALLBACK_URL")
	if callback_url == "" {
		callback_url = "http://localhost:8080/api/v1/auth/google/callback"
	}
	google_provider := google.New(google_client_id, google_client_secret, callback_url, "email", "profile")
	goth.UseProviders(google_provider)
	router.GET("/google/login", GoogleOAuthLogin)
	router.GET("/:provider/begin", GoogleOAuthBegin)
	router.GET("/:provider/callback", GoogleOAuthCallback)
}
