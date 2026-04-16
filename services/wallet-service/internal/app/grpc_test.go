package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
)

func TestWalletGRPCFlow(t *testing.T) {
	repo := &fakeWalletRepo{wallets: map[string]repository.Wallet{}, txs: map[string][]repository.Transaction{}}
	svc := NewService(repo, nil, nil)
	server := NewGRPCServer(svc)

	u1 := uuid.NewString()
	u2 := uuid.NewString()
	createReq1, _ := structpb.NewStruct(map[string]any{"user_id": u1, "currency": "USD"})
	_, err := server.CreateWallet(context.Background(), createReq1)
	require.NoError(t, err)
	createReq2, _ := structpb.NewStruct(map[string]any{"user_id": u2, "currency": "USD"})
	_, err = server.CreateWallet(context.Background(), createReq2)
	require.NoError(t, err)

	depReq, _ := structpb.NewStruct(map[string]any{"user_id": u1, "amount": "10.00", "reference": "dep-ref"})
	_, err = server.Deposit(context.Background(), depReq)
	require.NoError(t, err)

	wdReq, _ := structpb.NewStruct(map[string]any{"user_id": u1, "amount": "2.00", "reference": "wd-ref"})
	_, err = server.Withdraw(context.Background(), wdReq)
	require.NoError(t, err)

	trReq, _ := structpb.NewStruct(map[string]any{
		"from_user_id": u1,
		"to_user_id":   u2,
		"amount":       "1.00",
		"reference":    "tx-ref",
	})
	_, err = server.Transfer(context.Background(), trReq)
	require.NoError(t, err)

	getReq, _ := structpb.NewStruct(map[string]any{"user_id": u1})
	_, err = server.GetWallet(context.Background(), getReq)
	require.NoError(t, err)

	listReq, _ := structpb.NewStruct(map[string]any{"user_id": u1, "limit": 10, "offset": 0})
	_, err = server.ListTransactions(context.Background(), listReq)
	require.NoError(t, err)

	failWdReq, _ := structpb.NewStruct(map[string]any{"user_id": u1, "amount": "999.00", "reference": "large-withdraw"})
	_, err = server.Withdraw(context.Background(), failWdReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}
