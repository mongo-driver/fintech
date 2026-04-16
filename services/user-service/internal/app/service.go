package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
	"github.com/example/fintech-backend/shared/httpx"
)

type Service struct {
	repo  repository.Repository
	cache *redis.Client
}

func NewService(repo repository.Repository, cache *redis.Client) *Service {
	return &Service{repo: repo, cache: cache}
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
	Phone    string `json:"phone" validate:"required,min=6,max=20"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
	Phone    string `json:"phone" validate:"required,min=6,max=20"`
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (repository.User, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := httpx.Validate(req); err != nil {
		return repository.User{}, err
	}
	now := time.Now().UTC()
	user := repository.User{
		ID:        uuid.NewString(),
		Email:     req.Email,
		FullName:  req.FullName,
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return repository.User{}, err
	}
	_ = s.cacheUser(ctx, user)
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (repository.User, error) {
	if s.cache != nil {
		key := s.cacheKey(id)
		v, err := s.cache.Get(ctx, key).Result()
		if err == nil {
			var user repository.User
			if unmarshalErr := json.Unmarshal([]byte(v), &user); unmarshalErr == nil {
				return user, nil
			}
		}
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return repository.User{}, err
	}
	_ = s.cacheUser(ctx, user)
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (repository.User, error) {
	if err := httpx.Validate(req); err != nil {
		return repository.User{}, err
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return repository.User{}, err
	}
	user.FullName = req.FullName
	user.Phone = req.Phone
	user.UpdatedAt = time.Now().UTC()
	if err = s.repo.Update(ctx, user); err != nil {
		return repository.User{}, err
	}
	_ = s.cacheUser(ctx, user)
	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Del(ctx, s.cacheKey(id)).Err()
	}
	return nil
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]repository.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) cacheKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (s *Service) cacheUser(ctx context.Context, user repository.User) error {
	if s.cache == nil {
		return nil
	}
	body, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, s.cacheKey(user.ID), body, 5*time.Minute).Err()
}
