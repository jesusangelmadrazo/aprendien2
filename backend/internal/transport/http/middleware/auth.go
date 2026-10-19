package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("role", claims.Role)
		c.Next()
	}
}
