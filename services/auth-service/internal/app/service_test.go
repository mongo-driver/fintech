package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/auth-service/internal/repository"
	"github.com/example/fintech-backend/shared/security"
)

type fakeRepo struct {
	users map[string]repository.User
}

func (f *fakeRepo) Migrate(context.Context) error { return nil }
func (f *fakeRepo) CreateUser(_ context.Context, user repository.User) error {
	if f.users == nil {
		f.users = map[string]repository.User{}
	}
	f.users[user.Email] = user
	return nil
}
func (f *fakeRepo) GetByEmail(_ context.Context, email string) (repository.User, error) {
	u, ok := f.users[email]
	if !ok {
		return repository.User{}, repository.ErrNotFound
	}
	return u, nil
}
func (f *fakeRepo) GetByID(_ context.Context, id string) (repository.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return repository.User{}, repository.ErrNotFound
}

func TestRegisterAndLogin(t *testing.T) {
	repo := &fakeRepo{users: map[string]repository.User{}}
	jwt := security.NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, jwt, nil)

	registerResp, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Password: "SecurePass123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.Token)
	require.NotEmpty(t, registerResp.UserID)

	loginResp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "SecurePass123",
	})
	require.NoError(t, err)
	require.Equal(t, registerResp.UserID, loginResp.UserID)
	require.NotEmpty(t, loginResp.Token)
}

func TestLoginInvalidPassword(t *testing.T) {
	repo := &fakeRepo{users: map[string]repository.User{}}
	jwt := security.NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, jwt, nil)

	_, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Password: "SecurePass123",
	})
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRegisterInvalidInput(t *testing.T) {
	repo := &fakeRepo{users: map[string]repository.User{}}
	jwt := security.NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, jwt, nil)

	_, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "invalid-email",
		Password: "short",
	})
	require.Error(t, err)
}

func TestValidateToken(t *testing.T) {
	repo := &fakeRepo{users: map[string]repository.User{}}
	jwt := security.NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, jwt, nil)

	resp, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "token@example.com",
		Password: "SecurePass123",
	})
	require.NoError(t, err)

	claims, err := svc.ValidateToken(context.Background(), resp.Token)
	require.NoError(t, err)
	require.Equal(t, resp.UserID, claims.UserID)
	require.Equal(t, resp.Email, claims.Email)
}
