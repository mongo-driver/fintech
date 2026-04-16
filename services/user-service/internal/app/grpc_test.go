package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
	"github.com/example/fintech-backend/shared/grpcx"
)

func TestUserGRPCFlow(t *testing.T) {
	repo := &fakeUserRepo{data: map[string]repository.User{}}
	svc := NewService(repo, nil)
	server := NewGRPCServer(svc)

	createReq, _ := structpb.NewStruct(map[string]any{
		"email":     "grpc-user@example.com",
		"full_name": "Grpc User",
		"phone":     "989111111111",
	})
	createResp, err := server.CreateUser(context.Background(), createReq)
	require.NoError(t, err)
	created := grpcx.ToMap(createResp)
	id := grpcx.GetString(created, "id")
	require.NotEmpty(t, id)

	getReq, _ := structpb.NewStruct(map[string]any{"id": id})
	_, err = server.GetUser(context.Background(), getReq)
	require.NoError(t, err)

	updateReq, _ := structpb.NewStruct(map[string]any{"id": id, "full_name": "Grpc Updated", "phone": "989122222222"})
	_, err = server.UpdateUser(context.Background(), updateReq)
	require.NoError(t, err)

	listReq, _ := structpb.NewStruct(map[string]any{"limit": 10, "offset": 0})
	_, err = server.ListUsers(context.Background(), listReq)
	require.NoError(t, err)

	deleteReq, _ := structpb.NewStruct(map[string]any{"id": id})
	_, err = server.DeleteUser(context.Background(), deleteReq)
	require.NoError(t, err)

	_, err = server.GetUser(context.Background(), deleteReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}
