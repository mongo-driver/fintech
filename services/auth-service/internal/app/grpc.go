package app

import (
	"context"
	"errors"

	"github.com/example/fintech-backend/shared/contracts/authpb"
	"github.com/example/fintech-backend/shared/grpcx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type GRPCServer struct {
	authpb.AuthServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (g *GRPCServer) Register(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	resp, err := g.svc.Register(ctx, RegisterRequest{
		Email:    grpcx.GetString(m, "email"),
		Password: grpcx.GetString(m, "password"),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(map[string]any{
		"user_id": resp.UserID,
		"email":   resp.Email,
		"token":   resp.Token,
	})
}

func (g *GRPCServer) Login(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	resp, err := g.svc.Login(ctx, LoginRequest{
		Email:    grpcx.GetString(m, "email"),
		Password: grpcx.GetString(m, "password"),
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(map[string]any{
		"user_id": resp.UserID,
		"email":   resp.Email,
		"token":   resp.Token,
	})
}

func (g *GRPCServer) ValidateToken(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	claims, err := g.svc.ValidateToken(ctx, grpcx.GetString(m, "token"))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return grpcx.ToStruct(map[string]any{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt.Time.Unix(),
	})
}
