package accounts

import "log"

// EmailProvider sends transactional emails.
// Phase 6 wires a real implementation (Brevo/Resend/SES).
// The NoOpEmailProvider logs links to stdout so developers can grab them during local dev.
type EmailProvider interface {
	SendPasswordReset(to, resetLink string) error
	SendEmailVerification(to, verifyLink string) error
}

// NoOpEmailProvider satisfies EmailProvider without sending any real email.
// It logs the link so local developers can complete flows without an email service.
type NoOpEmailProvider struct{}

func (n *NoOpEmailProvider) SendPasswordReset(to, resetLink string) error {
	log.Printf("[DEV] Password reset link for %s: %s", to, resetLink)
	return nil
}

func (n *NoOpEmailProvider) SendEmailVerification(to, verifyLink string) error {
	log.Printf("[DEV] Email verification link for %s: %s", to, verifyLink)
	return nil
}

// DefaultEmailProvider is the active provider.
// Swap this in init() or by calling SetEmailProvider for testing and Phase 6.
var DefaultEmailProvider EmailProvider = &NoOpEmailProvider{}

// SetEmailProvider replaces the active email provider.
func SetEmailProvider(p EmailProvider) {
	DefaultEmailProvider = p
}
