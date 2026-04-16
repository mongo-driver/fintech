package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
)

func TestUserHTTPCreateAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeUserRepo{data: map[string]repository.User{}}
	svc := NewService(repo, nil)
	h := NewHTTPHandler(svc)

	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1/users"))

	body := map[string]string{
		"email":     "api-user@example.com",
		"full_name": "API User",
		"phone":     "989121212121",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var created repository.User
	err := json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	updatePayload, _ := json.Marshal(map[string]string{
		"full_name": "Updated Name",
		"phone":     "989122233344",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/users/"+created.ID, bytes.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/?limit=5&offset=0", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
