package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
	"github.com/example/fintech-backend/shared/events"
	"github.com/example/fintech-backend/shared/httpx"
)

type Service struct {
	repo      repository.Repository
	cache     *redis.Client
	publisher *events.Producer
}

func NewService(repo repository.Repository, cache *redis.Client, publisher *events.Producer) *Service {
	return &Service{repo: repo, cache: cache, publisher: publisher}
}

type CreateWalletRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid4"`
	Currency string `json:"currency" validate:"required,len=3"`
}

type MovementRequest struct {
	Amount    string `json:"amount" validate:"required"`
	Reference string `json:"reference" validate:"required,min=3,max=120"`
}

type TransferRequest struct {
	FromUserID string `json:"from_user_id" validate:"required,uuid4"`
	ToUserID   string `json:"to_user_id" validate:"required,uuid4"`
	Amount     string `json:"amount" validate:"required"`
	Reference  string `json:"reference" validate:"required,min=3,max=120"`
}

func (s *Service) CreateWallet(ctx context.Context, req CreateWalletRequest) (repository.Wallet, error) {
	if err := httpx.Validate(req); err != nil {
		return repository.Wallet{}, err
	}
	w, err := s.repo.CreateWallet(ctx, req.UserID, req.Currency)
	if err != nil {
		return repository.Wallet{}, err
	}
	_ = s.cacheWallet(ctx, w)
	return w, nil
}

func (s *Service) GetWallet(ctx context.Context, userID string) (repository.Wallet, error) {
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, s.walletKey(userID)).Result(); err == nil {
			var w repository.Wallet
			if jsonErr := json.Unmarshal([]byte(raw), &w); jsonErr == nil {
				return w, nil
			}
		}
	}
	w, err := s.repo.GetWallet(ctx, userID)
	if err != nil {
		return repository.Wallet{}, err
	}
	_ = s.cacheWallet(ctx, w)
	return w, nil
}

func (s *Service) Deposit(ctx context.Context, userID string, req MovementRequest) (repository.Wallet, error) {
	if err := httpx.Validate(req); err != nil {
		return repository.Wallet{}, err
	}
	amountCents, err := ParseAmountToCents(req.Amount)
	if err != nil {
		return repository.Wallet{}, err
	}
	w, err := s.repo.Deposit(ctx, userID, amountCents, req.Reference)
	if err != nil {
		return repository.Wallet{}, err
	}
	_ = s.cacheWallet(ctx, w)
	s.publishNotification(ctx, w.UserID, "wallet_deposit", fmt.Sprintf("Deposit %s successful", FormatCents(amountCents)))
	return w, nil
}

func (s *Service) Withdraw(ctx context.Context, userID string, req MovementRequest) (repository.Wallet, error) {
	if err := httpx.Validate(req); err != nil {
		return repository.Wallet{}, err
	}
	amountCents, err := ParseAmountToCents(req.Amount)
	if err != nil {
		return repository.Wallet{}, err
	}
	w, err := s.repo.Withdraw(ctx, userID, amountCents, req.Reference)
	if err != nil {
		return repository.Wallet{}, err
	}
	_ = s.cacheWallet(ctx, w)
	s.publishNotification(ctx, w.UserID, "wallet_withdraw", fmt.Sprintf("Withdraw %s successful", FormatCents(amountCents)))
	return w, nil
}

func (s *Service) Transfer(ctx context.Context, req TransferRequest) (repository.Wallet, repository.Wallet, error) {
	if err := httpx.Validate(req); err != nil {
		return repository.Wallet{}, repository.Wallet{}, err
	}
	amountCents, err := ParseAmountToCents(req.Amount)
	if err != nil {
		return repository.Wallet{}, repository.Wallet{}, err
	}
	fromW, toW, err := s.repo.Transfer(ctx, req.FromUserID, req.ToUserID, amountCents, req.Reference)
	if err != nil {
		return repository.Wallet{}, repository.Wallet{}, err
	}
	_ = s.cacheWallet(ctx, fromW)
	_ = s.cacheWallet(ctx, toW)
	s.publishNotification(ctx, fromW.UserID, "wallet_transfer_out", fmt.Sprintf("Transferred %s to user %s", FormatCents(amountCents), toW.UserID))
	s.publishNotification(ctx, toW.UserID, "wallet_transfer_in", fmt.Sprintf("Received %s from user %s", FormatCents(amountCents), fromW.UserID))
	return fromW, toW, nil
}

func (s *Service) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]repository.Transaction, error) {
	return s.repo.ListTransactions(ctx, userID, limit, offset)
}

func (s *Service) walletKey(userID string) string {
	return fmt.Sprintf("wallet:%s", userID)
}

func (s *Service) cacheWallet(ctx context.Context, wallet repository.Wallet) error {
	if s.cache == nil {
		return nil
	}
	body, err := json.Marshal(wallet)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, s.walletKey(wallet.UserID), body, 3*time.Minute).Err()
}

func (s *Service) publishNotification(ctx context.Context, userID, eventType, message string) {
	if s.publisher == nil {
		return
	}
	_ = s.publisher.Publish(ctx, userID, events.NewNotificationEvent(
		eventType,
		userID,
		"Wallet Update",
		message,
	))
}
