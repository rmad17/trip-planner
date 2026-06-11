package accounts

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const resetTokenExpiry = time.Hour

// generateRawToken returns a cryptographically random 32-byte hex string (64 chars).
func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hex digest of a raw token.
// We store the hash, never the raw token, so DB leaks don't grant access.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// isTokenExpired returns true when the given expiry time is in the past.
func isTokenExpired(expiry *time.Time) bool {
	if expiry == nil {
		return true
	}
	return time.Now().After(*expiry)
}

// PasswordResetRequestInput is the body for requesting a password reset.
type PasswordResetRequestInput struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// PasswordResetConfirmInput is the body for confirming a password reset.
type PasswordResetConfirmInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResendVerificationInput is the body for resending an email verification.
type ResendVerificationInput struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestPasswordReset godoc
// @Summary Request a password reset email
// @Description Sends a password reset link to the given email. Always returns 200 to prevent email enumeration.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body PasswordResetRequestInput true "Email address"
// @Success 200 {object} map[string]string "message"
// @Router /auth/password-reset/request [post]
func RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always return 200 regardless of whether the email exists.
	defer c.JSON(http.StatusOK, gin.H{"message": "If that email is registered, a reset link has been sent."})

	var user User
	if err := core.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return // email not found — 200 already deferred
	}

	raw, err := generateRawToken()
	if err != nil {
		return
	}
	hashed := hashToken(raw)
	expiry := time.Now().Add(resetTokenExpiry)

	core.DB.Model(&user).Updates(map[string]interface{}{
		"password_reset_token":  hashed,
		"password_reset_expiry": expiry,
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/auth/reset-password?token=%s", frontendURL, raw)
	DefaultEmailProvider.SendPasswordReset(req.Email, link) //nolint:errcheck — logged inside
}

// ResetPassword godoc
// @Summary Reset a user's password
// @Description Validates the reset token and updates the password.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body PasswordResetConfirmInput true "Token and new password"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid or expired token"
// @Router /auth/password-reset/confirm [post]
func ResetPassword(c *gin.Context) {
	var req PasswordResetConfirmInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed := hashToken(req.Token)
	var user User
	if err := core.DB.Where("password_reset_token = ?", hashed).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}
	if isTokenExpired(user.PasswordResetExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has expired"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	core.DB.Model(&user).Updates(map[string]interface{}{
		"password":              string(hash),
		"password_reset_token":  nil,
		"password_reset_expiry": nil,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Marks the user's email as verified using the token sent at signup.
// @Tags auth
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid token"
// @Router /auth/verify-email [get]
func VerifyEmail(c *gin.Context) {
	raw := c.Query("token")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	hashed := hashToken(raw)
	var user User
	if err := core.DB.Where("email_verification_token = ?", hashed).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification token"})
		return
	}

	core.DB.Model(&user).Updates(map[string]interface{}{
		"email_verified":           true,
		"email_verification_token": nil,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ResendVerificationEmail godoc
// @Summary Resend email verification
// @Description Regenerates and resends the email verification link.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body ResendVerificationInput true "Email address"
// @Success 200 {object} map[string]string "message"
// @Router /auth/verify-email/resend [post]
func ResendVerificationEmail(c *gin.Context) {
	// TODO: add per-email rate limiting (Phase 7) — prevent abuse of this endpoint.
	var req ResendVerificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer c.JSON(http.StatusOK, gin.H{"message": "If that email is registered and unverified, a verification link has been sent."})

	var user User
	if err := core.DB.Where("email = ? AND email_verified = ?", req.Email, false).First(&user).Error; err != nil {
		return
	}

	raw, err := generateRawToken()
	if err != nil {
		return
	}
	hashed := hashToken(raw)

	core.DB.Model(&user).Update("email_verification_token", hashed)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/auth/verify-email?token=%s", frontendURL, raw)
	DefaultEmailProvider.SendEmailVerification(req.Email, link) //nolint:errcheck
}

// sendVerificationEmailOnSignup is called from CreateUser after a new email-based registration.
// Generates the verification token and dispatches the email.
func sendVerificationEmailOnSignup(user *User) {
	if user.Email == nil {
		return
	}
	raw, err := generateRawToken()
	if err != nil {
		return
	}
	hashed := hashToken(raw)
	core.DB.Model(user).Update("email_verification_token", hashed)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/auth/verify-email?token=%s", frontendURL, raw)
	DefaultEmailProvider.SendEmailVerification(*user.Email, link) //nolint:errcheck
}
