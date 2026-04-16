package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/fintech-backend/shared/security"
)

const (
	contextUserIDKey = "user_id"
	contextEmailKey  = "email"
)

func JWTAuth(manager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid bearer token"})
			return
		}

		claims, err := manager.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextEmailKey, claims.Email)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	v, _ := c.Get(contextUserIDKey)
	s, _ := v.(string)
	return s
}
