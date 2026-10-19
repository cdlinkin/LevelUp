package core_auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-at-least-32-bytes!!"

func newTestTokenManager(ttl time.Duration) *TokenManager {
	return NewTokenManager(Config{
		JWTSecret: testSecret,
		TokenTTL:  ttl,
	})
}

func TestTokenManager_IssueAndParse(t *testing.T) {
	manager := newTestTokenManager(time.Hour)

	token, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}

	userID, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if userID != 42 {
		t.Errorf("Parse() userID = %d, want 42", userID)
	}
}

func TestTokenManager_ParseInvalid(t *testing.T) {
	manager := newTestTokenManager(time.Hour)

	otherManager := NewTokenManager(Config{
		JWTSecret: "another-secret-key-at-least-32-bytes",
		TokenTTL:  time.Hour,
	})
	foreignToken, err := otherManager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}

	expiredToken, err := newTestTokenManager(-time.Minute).Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}

	validToken, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	parts := strings.Split(validToken, ".")
	tamperedToken := parts[0] + "." + parts[1] + "." + strings.Repeat("A", len(parts[2]))

	noneToken, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(42),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}

	noExpToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: strconv.Itoa(42),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token without exp: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{name: "signed with another secret", token: foreignToken},
		{name: "expired", token: expiredToken},
		{name: "tampered signature", token: tamperedToken},
		{name: "alg none", token: noneToken},
		{name: "without exp", token: noExpToken},
		{name: "garbage", token: "not-a-token"},
		{name: "empty", token: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.Parse(tt.token)
			if !errors.Is(err, core_errors.ErrUnauthorized) {
				t.Errorf("Parse() error = %v, want %v", err, core_errors.ErrUnauthorized)
			}
		})
	}
}

func TestTokenManager_TTL(t *testing.T) {
	manager := newTestTokenManager(24 * time.Hour)

	if got := manager.TTL(); got != 24*time.Hour {
		t.Errorf("TTL() = %v, want %v", got, 24*time.Hour)
	}
}

func TestUserIDContext(t *testing.T) {
	ctx := ToContext(context.Background(), 5)

	userID, ok := UserIDFromContext(ctx)
	if !ok || userID != 5 {
		t.Errorf("UserIDFromContext() = (%d, %v), want (5, true)", userID, ok)
	}

	userID, ok = UserIDFromContext(context.Background())
	if ok || userID != 0 {
		t.Errorf("UserIDFromContext(empty) = (%d, %v), want (0, false)", userID, ok)
	}
}

func TestNewConfig(t *testing.T) {
	t.Run("valid secret", func(t *testing.T) {
		t.Setenv("AUTH_JWT_SECRET", testSecret)
		t.Setenv("AUTH_TOKEN_TTL", "2h")

		config, err := NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() error = %v, want nil", err)
		}
		if config.TokenTTL != 2*time.Hour {
			t.Errorf("TokenTTL = %v, want %v", config.TokenTTL, 2*time.Hour)
		}
	})

	t.Run("short secret", func(t *testing.T) {
		t.Setenv("AUTH_JWT_SECRET", "too-short")

		if _, err := NewConfig(); err == nil {
			t.Errorf("NewConfig() error = nil, want error for short secret")
		}
	})
}
