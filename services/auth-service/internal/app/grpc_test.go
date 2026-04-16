package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/example/fintech-backend/services/auth-service/internal/repository"
	"github.com/example/fintech-backend/shared/grpcx"
	"github.com/example/fintech-backend/shared/security"
)

func TestGRPCRegisterLoginValidate(t *testing.T) {
	svc := NewService(&fakeRepo{users: map[string]repository.User{}}, security.NewJWTManager("secret", time.Hour), nil)
	grpcServer := NewGRPCServer(svc)

	registerReq, _ := structpb.NewStruct(map[string]any{"email": "grpc@example.com", "password": "SecurePass123"})
	registerResp, err := grpcServer.Register(context.Background(), registerReq)
	require.NoError(t, err)
	registerMap := grpcx.ToMap(registerResp)
	require.NotEmpty(t, registerMap["token"])

	loginReq, _ := structpb.NewStruct(map[string]any{"email": "grpc@example.com", "password": "SecurePass123"})
	loginResp, err := grpcServer.Login(context.Background(), loginReq)
	require.NoError(t, err)
	loginMap := grpcx.ToMap(loginResp)
	require.NotEmpty(t, loginMap["token"])

	validateReq, _ := structpb.NewStruct(map[string]any{"token": loginMap["token"]})
	_, err = grpcServer.ValidateToken(context.Background(), validateReq)
	require.NoError(t, err)

	badLoginReq, _ := structpb.NewStruct(map[string]any{"email": "grpc@example.com", "password": "WrongPass999"})
	_, err = grpcServer.Login(context.Background(), badLoginReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}
