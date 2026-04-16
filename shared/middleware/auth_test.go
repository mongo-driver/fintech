package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/shared/security"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := security.NewJWTManager("secret", time.Hour)
	token, err := manager.Generate("u1", "x@y.com")
	require.NoError(t, err)

	r := gin.New()
	r.Use(JWTAuth(manager))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(200, gin.H{"user_id": UserID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
