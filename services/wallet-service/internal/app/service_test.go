package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
)

type fakeWalletRepo struct {
	wallets map[string]repository.Wallet
	txs     map[string][]repository.Transaction
}

func (f *fakeWalletRepo) Migrate(context.Context) error { return nil }
func (f *fakeWalletRepo) CreateWallet(_ context.Context, userID, currency string) (repository.Wallet, error) {
	w := repository.Wallet{
		ID:           uuid.NewString(),
		UserID:       userID,
		Currency:     currency,
		BalanceCents: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if f.wallets == nil {
		f.wallets = map[string]repository.Wallet{}
	}
	f.wallets[userID] = w
	return w, nil
}
func (f *fakeWalletRepo) GetWallet(_ context.Context, userID string) (repository.Wallet, error) {
	w, ok := f.wallets[userID]
	if !ok {
		return repository.Wallet{}, repository.ErrWalletNotFound
	}
	return w, nil
}
func (f *fakeWalletRepo) Deposit(_ context.Context, userID string, amountCents int64, reference string) (repository.Wallet, error) {
	w, err := f.GetWallet(context.Background(), userID)
	if err != nil {
		return repository.Wallet{}, err
	}
	w.BalanceCents += amountCents
	f.wallets[userID] = w
	f.txs[userID] = append(f.txs[userID], repository.Transaction{
		ID:          uuid.NewString(),
		WalletID:    w.ID,
		UserID:      userID,
		Type:        "deposit",
		AmountCents: amountCents,
		Reference:   reference,
		CreatedAt:   time.Now(),
	})
	return w, nil
}
func (f *fakeWalletRepo) Withdraw(_ context.Context, userID string, amountCents int64, reference string) (repository.Wallet, error) {
	w, err := f.GetWallet(context.Background(), userID)
	if err != nil {
		return repository.Wallet{}, err
	}
	if w.BalanceCents < amountCents {
		return repository.Wallet{}, repository.ErrInsufficientFunds
	}
	w.BalanceCents -= amountCents
	f.wallets[userID] = w
	f.txs[userID] = append(f.txs[userID], repository.Transaction{
		ID:          uuid.NewString(),
		WalletID:    w.ID,
		UserID:      userID,
		Type:        "withdraw",
		AmountCents: amountCents,
		Reference:   reference,
		CreatedAt:   time.Now(),
	})
	return w, nil
}
func (f *fakeWalletRepo) Transfer(_ context.Context, fromUserID, toUserID string, amountCents int64, reference string) (repository.Wallet, repository.Wallet, error) {
	from, err := f.Withdraw(context.Background(), fromUserID, amountCents, reference)
	if err != nil {
		return repository.Wallet{}, repository.Wallet{}, err
	}
	to, err := f.Deposit(context.Background(), toUserID, amountCents, reference)
	if err != nil {
		return repository.Wallet{}, repository.Wallet{}, err
	}
	return from, to, nil
}
func (f *fakeWalletRepo) ListTransactions(_ context.Context, userID string, _, _ int) ([]repository.Transaction, error) {
	return f.txs[userID], nil
}

func TestDepositWithdrawTransfer(t *testing.T) {
	repo := &fakeWalletRepo{wallets: map[string]repository.Wallet{}, txs: map[string][]repository.Transaction{}}
	svc := NewService(repo, nil, nil)

	u1 := uuid.NewString()
	u2 := uuid.NewString()
	_, err := svc.CreateWallet(context.Background(), CreateWalletRequest{UserID: u1, Currency: "USD"})
	require.NoError(t, err)
	_, err = svc.CreateWallet(context.Background(), CreateWalletRequest{UserID: u2, Currency: "USD"})
	require.NoError(t, err)

	w1, err := svc.Deposit(context.Background(), u1, MovementRequest{Amount: "100.00", Reference: "init"})
	require.NoError(t, err)
	require.Equal(t, int64(10000), w1.BalanceCents)

	w1, err = svc.Withdraw(context.Background(), u1, MovementRequest{Amount: "40.00", Reference: "bill"})
	require.NoError(t, err)
	require.Equal(t, int64(6000), w1.BalanceCents)

	from, to, err := svc.Transfer(context.Background(), TransferRequest{
		FromUserID: u1,
		ToUserID:   u2,
		Amount:     "10.00",
		Reference:  "p2p",
	})
	require.NoError(t, err)
	require.Equal(t, int64(5000), from.BalanceCents)
	require.Equal(t, int64(1000), to.BalanceCents)

	got, err := svc.GetWallet(context.Background(), u1)
	require.NoError(t, err)
	require.Equal(t, int64(5000), got.BalanceCents)

	txs, err := svc.ListTransactions(context.Background(), u1, 20, 0)
	require.NoError(t, err)
	require.NotEmpty(t, txs)
}

func TestParseAmountToCents(t *testing.T) {
	tests := []struct {
		in     string
		want   int64
		hasErr bool
	}{
		{"10", 1000, false},
		{"10.5", 1050, false},
		{"10.55", 1055, false},
		{"0", 0, true},
		{"-1", 0, true},
		{"1.555", 0, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("amount_%s", tt.in), func(t *testing.T) {
			got, err := ParseAmountToCents(tt.in)
			if tt.hasErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDepositInvalidAmount(t *testing.T) {
	repo := &fakeWalletRepo{wallets: map[string]repository.Wallet{}, txs: map[string][]repository.Transaction{}}
	svc := NewService(repo, nil, nil)
	userID := uuid.NewString()
	_, err := svc.CreateWallet(context.Background(), CreateWalletRequest{UserID: userID, Currency: "USD"})
	require.NoError(t, err)

	_, err = svc.Deposit(context.Background(), userID, MovementRequest{Amount: "bad", Reference: "invalid"})
	require.Error(t, err)
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	repo := &fakeWalletRepo{wallets: map[string]repository.Wallet{}, txs: map[string][]repository.Transaction{}}
	svc := NewService(repo, nil, nil)
	userID := uuid.NewString()
	_, err := svc.CreateWallet(context.Background(), CreateWalletRequest{UserID: userID, Currency: "USD"})
	require.NoError(t, err)

	_, err = svc.Withdraw(context.Background(), userID, MovementRequest{Amount: "1.00", Reference: "withdraw"})
	require.ErrorIs(t, err, repository.ErrInsufficientFunds)
}
