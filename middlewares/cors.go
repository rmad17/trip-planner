package middlewares

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
// Allows the frontend hosted on Digital Ocean Spaces to access the API
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment variable
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		var allowedOrigins []string

		if allowedOriginsEnv != "" {
			allowedOrigins = strings.Split(allowedOriginsEnv, ",")
			// Trim spaces
			for i := range allowedOrigins {
				allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
			}
		} else {
			// Default origins if not set
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"https://trip-planner-fe.blr1.digitaloceanspaces.com",
				"https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com",
			}
		}

		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Allow credentials (cookies, authorization headers)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allowed headers
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// Allowed methods
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// Expose headers that frontend can access
		c.Writer.Header().Set("Access-Control-Expose-Headers",
			"Content-Length, Content-Type, Content-Disposition")

		// Max age for preflight cache (24 hours)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
