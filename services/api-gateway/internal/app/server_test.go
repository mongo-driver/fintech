package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/example/fintech-backend/shared/security"
)

type fakeAuthClient struct {
	registerErr error
	loginErr    error
}

func (f fakeAuthClient) Register(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return structpb.NewStruct(map[string]any{
		"user_id": "u1",
		"email":   "a@b.com",
		"token":   "tkn",
	})
}
func (f fakeAuthClient) Login(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return structpb.NewStruct(map[string]any{
		"user_id": "u1",
		"email":   "a@b.com",
		"token":   "tkn",
	})
}
func (fakeAuthClient) ValidateToken(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"user_id": "u1"})
}

type fakeUserClient struct {
	createErr error
}

func (f fakeUserClient) CreateUser(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return structpb.NewStruct(map[string]any{"id": "u1"})
}
func (fakeUserClient) GetUser(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "u1"})
}
func (fakeUserClient) UpdateUser(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "u1"})
}
func (fakeUserClient) DeleteUser(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"ok": true})
}
func (fakeUserClient) ListUsers(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"data": []any{}})
}

type fakeWalletClient struct {
	withdrawErr error
}

func (fakeWalletClient) CreateWallet(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "w1"})
}
func (fakeWalletClient) GetWallet(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "w1"})
}
func (fakeWalletClient) Deposit(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "w1"})
}
func (f fakeWalletClient) Withdraw(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	if f.withdrawErr != nil {
		return nil, f.withdrawErr
	}
	return structpb.NewStruct(map[string]any{"id": "w1"})
}
func (fakeWalletClient) Transfer(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"id": "w1"})
}
func (fakeWalletClient) ListTransactions(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{"data": []any{}})
}

func setupServer(userClient fakeUserClient, walletClient fakeWalletClient, authClient fakeAuthClient) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	jwt := security.NewJWTManager("secret", time.Hour)
	token, _ := jwt.Generate("u1", "x@y.com")
	s := NewServer(authClient, userClient, walletClient, jwt)
	r := gin.New()
	s.RegisterRoutes(r)
	return r, token
}

func performJSON(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPublicAuthRoutes(t *testing.T) {
	r, _ := setupServer(fakeUserClient{}, fakeWalletClient{}, fakeAuthClient{})
	w := performJSON(r, http.MethodPost, "/api/v1/auth/register", "", map[string]string{"email": "a@b.com", "password": "Secure123"})
	require.Equal(t, http.StatusCreated, w.Code)
	w = performJSON(r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": "a@b.com", "password": "Secure123"})
	require.Equal(t, http.StatusOK, w.Code)
}

func TestProtectedRoutesWithoutToken(t *testing.T) {
	r, _ := setupServer(fakeUserClient{}, fakeWalletClient{}, fakeAuthClient{})
	w := performJSON(r, http.MethodGet, "/api/v1/users/u1", "", nil)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProtectedUserAndWalletRoutes(t *testing.T) {
	r, token := setupServer(fakeUserClient{}, fakeWalletClient{}, fakeAuthClient{})

	require.Equal(t, http.StatusCreated, performJSON(r, http.MethodPost, "/api/v1/users/", token, map[string]string{
		"email":     "u@x.com",
		"full_name": "U",
		"phone":     "123456",
	}).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodGet, "/api/v1/users/?limit=10&offset=0", token, nil).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodGet, "/api/v1/users/u1", token, nil).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodPut, "/api/v1/users/u1", token, map[string]string{
		"full_name": "N",
		"phone":     "987654",
	}).Code)
	require.Equal(t, http.StatusNoContent, performJSON(r, http.MethodDelete, "/api/v1/users/u1", token, nil).Code)

	require.Equal(t, http.StatusCreated, performJSON(r, http.MethodPost, "/api/v1/wallets/", token, map[string]string{
		"user_id":  "11111111-1111-1111-1111-111111111111",
		"currency": "USD",
	}).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodGet, "/api/v1/wallets/u1", token, nil).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodPost, "/api/v1/wallets/u1/deposit", token, map[string]string{
		"amount":    "10.00",
		"reference": "dep",
	}).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodPost, "/api/v1/wallets/u1/withdraw", token, map[string]string{
		"amount":    "5.00",
		"reference": "wd",
	}).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodPost, "/api/v1/wallets/transfer", token, map[string]string{
		"from_user_id": "u1",
		"to_user_id":   "u2",
		"amount":       "1.00",
		"reference":    "tx",
	}).Code)
	require.Equal(t, http.StatusOK, performJSON(r, http.MethodGet, "/api/v1/wallets/u1/transactions?limit=10&offset=0", token, nil).Code)
}

func TestGRPCErrorMapping(t *testing.T) {
	r, token := setupServer(fakeUserClient{
		createErr: status.Error(codes.InvalidArgument, "bad request"),
	}, fakeWalletClient{}, fakeAuthClient{})
	w := performJSON(r, http.MethodPost, "/api/v1/users/", token, map[string]string{
		"email":     "u@x.com",
		"full_name": "U",
		"phone":     "123456",
	})
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGRPCConflictAndUnauthorizedMapping(t *testing.T) {
	r, token := setupServer(fakeUserClient{}, fakeWalletClient{
		withdrawErr: status.Error(codes.FailedPrecondition, "insufficient"),
	}, fakeAuthClient{
		loginErr: status.Error(codes.Unauthenticated, "invalid"),
	})

	w := performJSON(r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": "a@b.com", "password": "wrong"})
	require.Equal(t, http.StatusUnauthorized, w.Code)

	w = performJSON(r, http.MethodPost, "/api/v1/wallets/u1/withdraw", token, map[string]string{
		"amount":    "10.00",
		"reference": "atm",
	})
	require.Equal(t, http.StatusConflict, w.Code)
}
