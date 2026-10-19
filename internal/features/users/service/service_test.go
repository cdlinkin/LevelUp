package users_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	users_service "github.com/cdlinkin/LevelUp/internal/features/users/service"
)

type fakeUsersRepository struct {
	called bool
}

func (f *fakeUsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	f.called = true
	return user, nil
}

func (f *fakeUsersRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	f.called = true
	return domain.User{
		ID:       id,
		Version:  1,
		FullName: "Roma Test",
		Email:    "rom@mail.ru",
	}, nil
}

func (f *fakeUsersRepository) DeleteUser(ctx context.Context, id int) error {
	f.called = true
	return nil
}

func (f *fakeUsersRepository) PatchUser(ctx context.Context, id int, user domain.User) (domain.User, error) {
	f.called = true
	return user, nil
}

func ptr[T any](v T) *T { return &v }

func TestUsersService_OtherUserForbidden(t *testing.T) {
	const currentUserID, otherUserID = 1, 2

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: ptr("Hacker"), Set: true},
	}

	tests := []struct {
		name string
		call func(s *users_service.UsersService) error
	}{
		{
			name: "get",
			call: func(s *users_service.UsersService) error {
				_, err := s.GetUser(context.Background(), currentUserID, otherUserID)
				return err
			},
		},
		{
			name: "patch",
			call: func(s *users_service.UsersService) error {
				_, err := s.PatchUser(context.Background(), currentUserID, otherUserID, patch)
				return err
			},
		},
		{
			name: "delete",
			call: func(s *users_service.UsersService) error {
				return s.DeleteUser(context.Background(), currentUserID, otherUserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUsersRepository{}
			service := users_service.NewUsersService(repo)

			err := tt.call(service)

			if !errors.Is(err, core_errors.ErrForbidden) {
				t.Errorf("error = %v, want %v", err, core_errors.ErrForbidden)
			}
			if repo.called {
				t.Errorf("repository was called for other user, want not called")
			}
		})
	}
}

func TestUsersService_Self(t *testing.T) {
	repo := &fakeUsersRepository{}
	service := users_service.NewUsersService(repo)

	user, err := service.GetUser(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("GetUser() error = %v, want nil", err)
	}
	if user.ID != 1 {
		t.Errorf("GetUser() ID = %d, want 1", user.ID)
	}

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: ptr("New Name"), Set: true},
	}

	patched, err := service.PatchUser(context.Background(), 1, 1, patch)
	if err != nil {
		t.Fatalf("PatchUser() error = %v, want nil", err)
	}
	if patched.FullName != "New Name" {
		t.Errorf("PatchUser() FullName = %q, want %q", patched.FullName, "New Name")
	}

	if err := service.DeleteUser(context.Background(), 1, 1); err != nil {
		t.Errorf("DeleteUser() error = %v, want nil", err)
	}
}
