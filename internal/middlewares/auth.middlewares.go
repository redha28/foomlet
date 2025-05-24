package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redha28/foomlet/internal/config"
	"github.com/redha28/foomlet/internal/models"
	"github.com/redha28/foomlet/pkg"
)

// AuthMiddleware checks if the request has a valid JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := models.NewResponse(c)
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			response.Unauthorized("Authorization header is required", nil)
			return
		}

		// Check if the header format is correct
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized("Invalid authorization format", "Authorization header must be Bearer {token}")
			return
		}

		tokenString := parts[1]

		// Get configuration
		cfg := config.GetConfig()

		// Initialize JWT util with config values
		jwtUtil := pkg.NewJwtUtil(
			cfg.JWT.AccessSecret,
			cfg.JWT.RefreshSecret,
			cfg.JWT.AccessExpiry,
			cfg.JWT.RefreshExpiry,
		)

		// Validate the token
		claims, err := jwtUtil.ValidateAccessToken(tokenString)
		if err != nil {
			response.Unauthorized("Invalid or expired token", err.Error())
			return
		}

		// Set user ID in the context for use in handlers
		c.Set("userID", claims.UserID)

		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}

	return userID.(string), true
}
