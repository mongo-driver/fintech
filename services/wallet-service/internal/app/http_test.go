package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
)

func TestWalletHTTPFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeWalletRepo{wallets: map[string]repository.Wallet{}, txs: map[string][]repository.Transaction{}}
	svc := NewService(repo, nil, nil)
	handler := NewHTTPHandler(svc)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1/wallets"))

	userID := uuid.NewString()
	createBody, _ := json.Marshal(map[string]string{"user_id": userID, "currency": "USD"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	depositBody, _ := json.Marshal(map[string]string{"amount": "20.00", "reference": "fund"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallets/"+userID+"/deposit", bytes.NewReader(depositBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+userID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	withdrawBody, _ := json.Marshal(map[string]string{"amount": "5.00", "reference": "atm"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallets/"+userID+"/withdraw", bytes.NewReader(withdrawBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	peerID := uuid.NewString()
	createPeer, _ := json.Marshal(map[string]string{"user_id": peerID, "currency": "USD"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallets/", bytes.NewReader(createPeer))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	transferBody, _ := json.Marshal(map[string]string{
		"from_user_id": userID,
		"to_user_id":   peerID,
		"amount":       "1.00",
		"reference":    "p2p",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(transferBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+userID+"/transactions?limit=10&offset=0", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	tooMuchBody, _ := json.Marshal(map[string]string{"amount": "500.00", "reference": "too-much"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/wallets/"+userID+"/withdraw", bytes.NewReader(tooMuchBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}
