package app

import (
	"context"
	"time"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
	"github.com/example/fintech-backend/shared/contracts/userpb"
	"github.com/example/fintech-backend/shared/grpcx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type GRPCServer struct {
	userpb.UserServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (g *GRPCServer) CreateUser(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	user, err := g.svc.CreateUser(ctx, CreateUserRequest{
		Email:    grpcx.GetString(m, "email"),
		FullName: grpcx.GetString(m, "full_name"),
		Phone:    grpcx.GetString(m, "phone"),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(userMap(user))
}

func (g *GRPCServer) GetUser(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	user, err := g.svc.GetUser(ctx, grpcx.GetString(m, "id"))
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return grpcx.ToStruct(userMap(user))
}

func (g *GRPCServer) UpdateUser(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	user, err := g.svc.UpdateUser(ctx, grpcx.GetString(m, "id"), UpdateUserRequest{
		FullName: grpcx.GetString(m, "full_name"),
		Phone:    grpcx.GetString(m, "phone"),
	})
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(userMap(user))
}

func (g *GRPCServer) DeleteUser(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	if err := g.svc.DeleteUser(ctx, grpcx.GetString(m, "id")); err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return grpcx.ToStruct(map[string]any{"ok": true})
}

func (g *GRPCServer) ListUsers(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	users, err := g.svc.ListUsers(ctx, int(grpcx.GetFloat64(m, "limit")), int(grpcx.GetFloat64(m, "offset")))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := make([]any, 0, len(users))
	for _, u := range users {
		result = append(result, userMap(u))
	}
	return grpcx.ToStruct(map[string]any{"data": result})
}

func userMap(user repository.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"phone":      user.Phone,
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	}
}
