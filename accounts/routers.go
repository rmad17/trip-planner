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
}

func RouterGroupGoogleOAuth(router *gin.RouterGroup) {
	google_client_id := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	google_provider := google.New(google_client_id, google_client_secret, "http://localhost:8080/auth/google/callback", "email", "profile")
	goth.UseProviders(google_provider)
	router.GET("/auth/google/login", GoogleOAuthLogin)
	router.GET("/:provider/begin", GoogleOAuthBegin)
	router.GET("/auth/:provider/callback", GoogleOAuthCallback)
}
