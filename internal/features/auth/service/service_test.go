package auth_service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	auth_service "github.com/cdlinkin/LevelUp/internal/features/auth/service"
)

type fakeUsersRepository struct {
	createUserCalled bool
	createUserArg    domain.User
	createUserErr    error

	getUserByEmailArg    string
	getUserByEmailResult domain.User
	getUserByEmailErr    error

	getUserResult domain.User
	getUserErr    error
}

func (f *fakeUsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	f.createUserCalled = true
	f.createUserArg = user

	if f.createUserErr != nil {
		return domain.User{}, f.createUserErr
	}

	user.ID = 1
	return user, nil
}

func (f *fakeUsersRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	f.getUserByEmailArg = email
	return f.getUserByEmailResult, f.getUserByEmailErr
}

func (f *fakeUsersRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	return f.getUserResult, f.getUserErr
}

type fakeTokenManager struct {
	issueCalled bool
	issueArg    int
}

func (f *fakeTokenManager) Issue(userID int) (string, error) {
	f.issueCalled = true
	f.issueArg = userID
	return "test-token", nil
}

func (f *fakeTokenManager) TTL() time.Duration {
	return time.Hour
}

func existingUser(t *testing.T) domain.User {
	t.Helper()

	hash, err := domain.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	return domain.User{
		ID:           1,
		Version:      1,
		FullName:     "Denis Test",
		Email:        "den@mail.ru",
		PasswordHash: hash,
	}
}

func TestRegister_Success(t *testing.T) {
	repo := &fakeUsersRepository{}
	service := auth_service.NewAuthService(repo, &fakeTokenManager{})

	user, err := service.Register(context.Background(), "Denis Test", nil, "  Den@Mail.RU ", "secret123")
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}
	if !repo.createUserCalled {
		t.Fatalf("repository CreateUser was not called, want called")
	}

	saved := repo.createUserArg
	if saved.Email != "den@mail.ru" {
		t.Errorf("saved Email = %q, want %q", saved.Email, "den@mail.ru")
	}
	if saved.PasswordHash == "secret123" {
		t.Errorf("password saved as plain text, want hash")
	}
	if !saved.CheckPassword("secret123") {
		t.Errorf("saved hash does not match password")
	}
	if user.ID != 1 {
		t.Errorf("Register() ID = %d, want 1", user.ID)
	}
}

func TestRegister_InvalidData(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		email    string
		password string
	}{
		{name: "short password", fullName: "Denis Test", email: "den@mail.ru", password: "123"},
		{name: "invalid email", fullName: "Denis Test", email: "not-an-email", password: "secret123"},
		{name: "short name", fullName: "De", email: "den@mail.ru", password: "secret123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUsersRepository{}
			service := auth_service.NewAuthService(repo, &fakeTokenManager{})

			_, err := service.Register(context.Background(), tt.fullName, nil, tt.email, tt.password)

			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Errorf("Register() error = %v, want %v", err, core_errors.ErrInvalidArgument)
			}
			if repo.createUserCalled {
				t.Errorf("repository CreateUser was called with invalid data, want not called")
			}
		})
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	repo := &fakeUsersRepository{
		createUserErr: fmt.Errorf("user with email=den@mail.ru already exists: %w", core_errors.ErrConflict),
	}
	service := auth_service.NewAuthService(repo, &fakeTokenManager{})

	_, err := service.Register(context.Background(), "Denis Test", nil, "den@mail.ru", "secret123")

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Errorf("Register() error = %v, want %v", err, core_errors.ErrConflict)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := &fakeUsersRepository{
		getUserByEmailResult: existingUser(t),
	}
	tokens := &fakeTokenManager{}
	service := auth_service.NewAuthService(repo, tokens)

	token, ttl, err := service.Login(context.Background(), " DEN@Mail.ru", "secret123")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}

	if repo.getUserByEmailArg != "den@mail.ru" {
		t.Errorf("repository got email = %q, want normalized %q", repo.getUserByEmailArg, "den@mail.ru")
	}
	if !tokens.issueCalled || tokens.issueArg != 1 {
		t.Errorf("token issued for user %d (called=%v), want user 1", tokens.issueArg, tokens.issueCalled)
	}
	if token != "test-token" {
		t.Errorf("Login() token = %q, want %q", token, "test-token")
	}
	if ttl != time.Hour {
		t.Errorf("Login() ttl = %v, want %v", ttl, time.Hour)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	wrongPasswordRepo := &fakeUsersRepository{
		getUserByEmailResult: existingUser(t),
	}
	unknownEmailRepo := &fakeUsersRepository{
		getUserByEmailErr: fmt.Errorf("user with email=nobody@mail.ru: %w", core_errors.ErrNotFound),
	}

	wrongPasswordTokens := &fakeTokenManager{}
	_, _, wrongPasswordErr := auth_service.NewAuthService(wrongPasswordRepo, wrongPasswordTokens).
		Login(context.Background(), "den@mail.ru", "wrong-password")

	unknownEmailTokens := &fakeTokenManager{}
	_, _, unknownEmailErr := auth_service.NewAuthService(unknownEmailRepo, unknownEmailTokens).
		Login(context.Background(), "nobody@mail.ru", "secret123")

	for name, err := range map[string]error{
		"wrong password": wrongPasswordErr,
		"unknown email":  unknownEmailErr,
	} {
		if !errors.Is(err, core_errors.ErrUnauthorized) {
			t.Errorf("%s: Login() error = %v, want %v", name, err, core_errors.ErrUnauthorized)
		}
	}

	if wrongPasswordErr.Error() != unknownEmailErr.Error() {
		t.Errorf("errors differ, attacker can tell which emails exist:\n wrong password: %v\n unknown email:  %v",
			wrongPasswordErr, unknownEmailErr)
	}

	if wrongPasswordTokens.issueCalled || unknownEmailTokens.issueCalled {
		t.Errorf("token was issued for invalid credentials")
	}
}

func TestLogin_RepositoryFailure(t *testing.T) {
	repo := &fakeUsersRepository{
		getUserByEmailErr: errors.New("connection refused"),
	}
	service := auth_service.NewAuthService(repo, &fakeTokenManager{})

	_, _, err := service.Login(context.Background(), "den@mail.ru", "secret123")

	if err == nil {
		t.Fatalf("Login() error = nil, want error")
	}
	if errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("Login() error = %v, database failure must not look like wrong password", err)
	}
}
