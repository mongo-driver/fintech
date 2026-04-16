package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHTTPManualSend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService(&fakeSender{}, zap.NewNop())
	h := NewHTTPHandler(svc)
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1/notifications"))

	body, _ := json.Marshal(map[string]string{
		"user_id": "u1",
		"subject": "hello",
		"message": "world",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)
}
