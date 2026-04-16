package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/auth-service/internal/repository"
	"github.com/example/fintech-backend/shared/security"
)

func TestHTTPRegisterAndLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService(&fakeRepo{users: map[string]repository.User{}}, security.NewJWTManager("secret", time.Hour), nil)
	h := NewHTTPHandler(svc)
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1/auth"))

	registerBody := map[string]string{
		"email":    "api@example.com",
		"password": "SecurePass123",
	}
	data, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody := map[string]string{
		"email":    "api@example.com",
		"password": "SecurePass123",
	}
	data, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token := loginResp["token"].(string)

	validateData, _ := json.Marshal(map[string]string{"token": token})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate", bytes.NewReader(validateData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/token", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotImplemented, w.Code)

	badLogin, _ := json.Marshal(map[string]string{
		"email":    "api@example.com",
		"password": "wrongpass",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLogin))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
