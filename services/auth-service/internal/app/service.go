package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/example/fintech-backend/services/auth-service/internal/repository"
	"github.com/example/fintech-backend/shared/events"
	"github.com/example/fintech-backend/shared/httpx"
	"github.com/example/fintech-backend/shared/security"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repo      repository.Repository
	jwt       *security.JWTManager
	publisher *events.Producer
}

func NewService(repo repository.Repository, jwt *security.JWTManager, publisher *events.Producer) *Service {
	return &Service{
		repo:      repo,
		jwt:       jwt,
		publisher: publisher,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := httpx.Validate(req); err != nil {
		return AuthResponse{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}
	user := repository.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err = s.repo.CreateUser(ctx, user); err != nil {
		return AuthResponse{}, err
	}

	token, err := s.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return AuthResponse{}, err
	}

	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, user.ID, events.NewNotificationEvent(
			"user_registered",
			user.ID,
			"Welcome to Fintech Platform",
			"Your account has been created successfully.",
		))
	}

	return AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Token:  token,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := httpx.Validate(req); err != nil {
		return AuthResponse{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if err = security.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
		Token:  token,
	}, nil
}

func (s *Service) ValidateToken(_ context.Context, token string) (*security.TokenClaims, error) {
	return s.jwt.Parse(token)
}
