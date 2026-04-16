package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
)

type fakeUserRepo struct {
	data map[string]repository.User
}

func (f *fakeUserRepo) Migrate(context.Context) error { return nil }
func (f *fakeUserRepo) Create(_ context.Context, user repository.User) error {
	f.data[user.ID] = user
	return nil
}
func (f *fakeUserRepo) GetByID(_ context.Context, id string) (repository.User, error) {
	v, ok := f.data[id]
	if !ok {
		return repository.User{}, repository.ErrNotFound
	}
	return v, nil
}
func (f *fakeUserRepo) Update(_ context.Context, user repository.User) error {
	if _, ok := f.data[user.ID]; !ok {
		return repository.ErrNotFound
	}
	f.data[user.ID] = user
	return nil
}
func (f *fakeUserRepo) Delete(_ context.Context, id string) error {
	if _, ok := f.data[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.data, id)
	return nil
}
func (f *fakeUserRepo) List(_ context.Context, _, _ int) ([]repository.User, error) {
	out := make([]repository.User, 0, len(f.data))
	for _, u := range f.data {
		out = append(out, u)
	}
	return out, nil
}

func TestCreateUpdateDeleteUser(t *testing.T) {
	repo := &fakeUserRepo{data: map[string]repository.User{}}
	svc := NewService(repo, nil)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Email:    "one@example.com",
		FullName: "Test User",
		Phone:    "989111111111",
	})
	require.NoError(t, err)
	require.NotEmpty(t, user.ID)

	updated, err := svc.UpdateUser(context.Background(), user.ID, UpdateUserRequest{
		FullName: "Updated",
		Phone:    "989122222222",
	})
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.FullName)

	err = svc.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)
}

func TestGetAndListUsers(t *testing.T) {
	repo := &fakeUserRepo{data: map[string]repository.User{}}
	svc := NewService(repo, nil)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Email:    "list@example.com",
		FullName: "List User",
		Phone:    "989100000000",
	})
	require.NoError(t, err)

	got, err := svc.GetUser(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)

	users, err := svc.ListUsers(context.Background(), 10, 0)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestGetUserNotFound(t *testing.T) {
	repo := &fakeUserRepo{data: map[string]repository.User{}}
	svc := NewService(repo, nil)
	_, err := svc.GetUser(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)
}
