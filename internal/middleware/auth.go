package middleware

import (
	"net/http"
	"strings"

	"github.com/datmedevil17/gopher-uptime/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token (Bearer <token>)
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			// No Bearer prefix found
			token = authHeader
		}

		// Verify JWT
		userID, err := utils.VerifyJWT(token, jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token: "+err.Error())
			c.Abort()
			return
		}

		// Store userID in context
		c.Set("userID", userID)
		c.Next()
	}
}
