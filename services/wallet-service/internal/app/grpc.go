package app

import (
	"context"
	"time"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
	"github.com/example/fintech-backend/shared/contracts/walletpb"
	"github.com/example/fintech-backend/shared/grpcx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type GRPCServer struct {
	walletpb.WalletServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (g *GRPCServer) CreateWallet(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	w, err := g.svc.CreateWallet(ctx, CreateWalletRequest{
		UserID:   grpcx.GetString(m, "user_id"),
		Currency: grpcx.GetString(m, "currency"),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(walletMap(w))
}

func (g *GRPCServer) GetWallet(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	w, err := g.svc.GetWallet(ctx, grpcx.GetString(m, "user_id"))
	if err != nil {
		if err == repository.ErrWalletNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return grpcx.ToStruct(walletMap(w))
}

func (g *GRPCServer) Deposit(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	w, err := g.svc.Deposit(ctx, grpcx.GetString(m, "user_id"), MovementRequest{
		Amount:    grpcx.GetString(m, "amount"),
		Reference: grpcx.GetString(m, "reference"),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(walletMap(w))
}

func (g *GRPCServer) Withdraw(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	w, err := g.svc.Withdraw(ctx, grpcx.GetString(m, "user_id"), MovementRequest{
		Amount:    grpcx.GetString(m, "amount"),
		Reference: grpcx.GetString(m, "reference"),
	})
	if err != nil {
		if err == repository.ErrInsufficientFunds {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(walletMap(w))
}

func (g *GRPCServer) Transfer(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	fromWallet, toWallet, err := g.svc.Transfer(ctx, TransferRequest{
		FromUserID: grpcx.GetString(m, "from_user_id"),
		ToUserID:   grpcx.GetString(m, "to_user_id"),
		Amount:     grpcx.GetString(m, "amount"),
		Reference:  grpcx.GetString(m, "reference"),
	})
	if err != nil {
		if err == repository.ErrInsufficientFunds {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return grpcx.ToStruct(map[string]any{
		"from_wallet": walletMap(fromWallet),
		"to_wallet":   walletMap(toWallet),
	})
}

func (g *GRPCServer) ListTransactions(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	m := grpcx.ToMap(in)
	txs, err := g.svc.ListTransactions(
		ctx,
		grpcx.GetString(m, "user_id"),
		int(grpcx.GetFloat64(m, "limit")),
		int(grpcx.GetFloat64(m, "offset")),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	data := make([]any, 0, len(txs))
	for _, tx := range txs {
		data = append(data, map[string]any{
			"id":         tx.ID,
			"user_id":    tx.UserID,
			"type":       tx.Type,
			"amount":     FormatCents(tx.AmountCents),
			"reference":  tx.Reference,
			"created_at": tx.CreatedAt.Format(time.RFC3339),
		})
	}
	return grpcx.ToStruct(map[string]any{"data": data})
}

func walletMap(w repository.Wallet) map[string]any {
	return map[string]any{
		"id":         w.ID,
		"user_id":    w.UserID,
		"currency":   w.Currency,
		"balance":    FormatCents(w.BalanceCents),
		"created_at": w.CreatedAt.Format(time.RFC3339),
		"updated_at": w.UpdatedAt.Format(time.RFC3339),
	}
}
