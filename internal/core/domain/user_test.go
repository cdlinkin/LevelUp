package domain

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "secret123",
			wantErr:  nil,
		},
		{
			name:     "7 symbols",
			password: strings.Repeat("a", 7),
			wantErr:  core_errors.ErrInvalidArgument,
		},
		{
			name:     "8 symbols",
			password: strings.Repeat("a", 8),
			wantErr:  nil,
		},
		{
			name:     "72 bytes",
			password: strings.Repeat("a", 72),
			wantErr:  nil,
		},
		{
			name:     "73 bytes",
			password: strings.Repeat("a", 73),
			wantErr:  core_errors.ErrInvalidArgument,
		},
		{
			name:     "40 cyrillic letters is 80 bytes",
			password: strings.Repeat("я", 40),
			wantErr:  core_errors.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HashPassword(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("HashPassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if hash == "secret123" {
		t.Fatalf("hash equals plain password")
	}

	user := User{PasswordHash: hash}

	if !user.CheckPassword("secret123") {
		t.Errorf("CheckPassword(correct) = false, want true")
	}
	if user.CheckPassword("wrong-password") {
		t.Errorf("CheckPassword(wrong) = true, want false")
	}

	emptyUser := User{}
	if emptyUser.CheckPassword("secret123") {
		t.Errorf("CheckPassword() for user without hash = true, want false")
	}
}

func TestHashPasswordUsesSalt(t *testing.T) {
	first, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}
	second, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if first == second {
		t.Errorf("two hashes of the same password are equal, want different (salt)")
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{email: "den@mail.ru", want: "den@mail.ru"},
		{email: "Den@Mail.RU", want: "den@mail.ru"},
		{email: "  den@mail.ru  ", want: "den@mail.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := NormalizeEmail(tt.email); got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestUserValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "den@mail.ru",
			wantErr: nil,
		},
		{
			name:    "not normalized",
			email:   "Den@Mail.ru",
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name:    "not an email",
			email:   "not-an-email",
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name:    "with display name",
			email:   "Den <den@mail.ru>",
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name:    "empty",
			email:   "",
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name:    "251 symbols",
			email:   strings.Repeat("a", 245) + "@x.ru" + "a",
			wantErr: core_errors.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				FullName: "Denis Test",
				Email:    tt.email,
			}

			err := user.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
