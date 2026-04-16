package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Wallet struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Currency     string    `json:"currency"`
	BalanceCents int64     `json:"balance_cents"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	AmountCents int64     `json:"amount_cents"`
	Reference   string    `json:"reference"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository interface {
	Migrate(ctx context.Context) error
	CreateWallet(ctx context.Context, userID, currency string) (Wallet, error)
	GetWallet(ctx context.Context, userID string) (Wallet, error)
	Deposit(ctx context.Context, userID string, amountCents int64, reference string) (Wallet, error)
	Withdraw(ctx context.Context, userID string, amountCents int64, reference string) (Wallet, error)
	Transfer(ctx context.Context, fromUserID, toUserID string, amountCents int64, reference string) (Wallet, Wallet, error)
	ListTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Migrate(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    balance_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    type TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    reference TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_created_at ON transactions(wallet_id, created_at DESC);
`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *PostgresRepository) CreateWallet(ctx context.Context, userID, currency string) (Wallet, error) {
	now := time.Now().UTC()
	w := Wallet{
		ID:           uuid.NewString(),
		UserID:       userID,
		Currency:     currency,
		BalanceCents: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO wallets (id,user_id,currency,balance_cents,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		w.ID, w.UserID, w.Currency, w.BalanceCents, w.CreatedAt, w.UpdatedAt,
	)
	return w, err
}

func (r *PostgresRepository) GetWallet(ctx context.Context, userID string) (Wallet, error) {
	var w Wallet
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,currency,balance_cents,created_at,updated_at FROM wallets WHERE user_id=$1`,
		userID,
	).Scan(&w.ID, &w.UserID, &w.Currency, &w.BalanceCents, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Wallet{}, ErrWalletNotFound
		}
		return Wallet{}, err
	}
	return w, nil
}

func (r *PostgresRepository) Deposit(ctx context.Context, userID string, amountCents int64, reference string) (Wallet, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Wallet{}, err
	}
	defer tx.Rollback(ctx)

	w, err := getWalletForUpdate(ctx, tx, userID)
	if err != nil {
		return Wallet{}, err
	}
	w.BalanceCents += amountCents
	w.UpdatedAt = time.Now().UTC()
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance_cents=$2,updated_at=$3 WHERE id=$1`, w.ID, w.BalanceCents, w.UpdatedAt); err != nil {
		return Wallet{}, err
	}
	if err = insertTransaction(ctx, tx, w.ID, "deposit", amountCents, reference); err != nil {
		return Wallet{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return Wallet{}, err
	}
	return w, nil
}

func (r *PostgresRepository) Withdraw(ctx context.Context, userID string, amountCents int64, reference string) (Wallet, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Wallet{}, err
	}
	defer tx.Rollback(ctx)

	w, err := getWalletForUpdate(ctx, tx, userID)
	if err != nil {
		return Wallet{}, err
	}
	if w.BalanceCents < amountCents {
		return Wallet{}, ErrInsufficientFunds
	}
	w.BalanceCents -= amountCents
	w.UpdatedAt = time.Now().UTC()
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance_cents=$2,updated_at=$3 WHERE id=$1`, w.ID, w.BalanceCents, w.UpdatedAt); err != nil {
		return Wallet{}, err
	}
	if err = insertTransaction(ctx, tx, w.ID, "withdraw", amountCents, reference); err != nil {
		return Wallet{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return Wallet{}, err
	}
	return w, nil
}

func (r *PostgresRepository) Transfer(ctx context.Context, fromUserID, toUserID string, amountCents int64, reference string) (Wallet, Wallet, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Wallet{}, Wallet{}, err
	}
	defer tx.Rollback(ctx)

	fromWallet, err := getWalletForUpdate(ctx, tx, fromUserID)
	if err != nil {
		return Wallet{}, Wallet{}, err
	}
	toWallet, err := getWalletForUpdate(ctx, tx, toUserID)
	if err != nil {
		return Wallet{}, Wallet{}, err
	}
	if fromWallet.BalanceCents < amountCents {
		return Wallet{}, Wallet{}, ErrInsufficientFunds
	}

	now := time.Now().UTC()
	fromWallet.BalanceCents -= amountCents
	fromWallet.UpdatedAt = now
	toWallet.BalanceCents += amountCents
	toWallet.UpdatedAt = now

	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance_cents=$2,updated_at=$3 WHERE id=$1`, fromWallet.ID, fromWallet.BalanceCents, fromWallet.UpdatedAt); err != nil {
		return Wallet{}, Wallet{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance_cents=$2,updated_at=$3 WHERE id=$1`, toWallet.ID, toWallet.BalanceCents, toWallet.UpdatedAt); err != nil {
		return Wallet{}, Wallet{}, err
	}

	if err = insertTransaction(ctx, tx, fromWallet.ID, "transfer_out", amountCents, reference); err != nil {
		return Wallet{}, Wallet{}, err
	}
	if err = insertTransaction(ctx, tx, toWallet.ID, "transfer_in", amountCents, reference); err != nil {
		return Wallet{}, Wallet{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return Wallet{}, Wallet{}, err
	}
	return fromWallet, toWallet, nil
}

func (r *PostgresRepository) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.pool.Query(ctx, `
SELECT t.id, t.wallet_id, w.user_id, t.type, t.amount_cents, t.reference, t.created_at
FROM transactions t
JOIN wallets w ON w.id = t.wallet_id
WHERE w.user_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Transaction, 0, limit)
	for rows.Next() {
		var tx Transaction
		if err = rows.Scan(&tx.ID, &tx.WalletID, &tx.UserID, &tx.Type, &tx.AmountCents, &tx.Reference, &tx.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}

func getWalletForUpdate(ctx context.Context, tx pgx.Tx, userID string) (Wallet, error) {
	var w Wallet
	err := tx.QueryRow(ctx,
		`SELECT id,user_id,currency,balance_cents,created_at,updated_at FROM wallets WHERE user_id=$1 FOR UPDATE`,
		userID,
	).Scan(&w.ID, &w.UserID, &w.Currency, &w.BalanceCents, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Wallet{}, ErrWalletNotFound
		}
		return Wallet{}, err
	}
	return w, nil
}

func insertTransaction(ctx context.Context, tx pgx.Tx, walletID, txType string, amountCents int64, reference string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO transactions (id,wallet_id,type,amount_cents,reference,created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), walletID, txType, amountCents, reference, time.Now().UTC(),
	)
	return err
}
