package core_auth

import (
	"fmt"
	"strconv"
	"time"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(config Config) *TokenManager {
	return &TokenManager{
		secret: []byte(config.JWTSecret),
		ttl:    config.TokenTTL,
	}
}

func (m *TokenManager) TTL() time.Duration {
	return m.ttl
}

func (m *TokenManager) Issue(userID int) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}

func (m *TokenManager) Parse(tokenString string) (int, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse token: %v: %w",
			err,
			core_errors.ErrUnauthorized,
		)
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid token subject %q: %w",
			claims.Subject,
			core_errors.ErrUnauthorized,
		)
	}

	return userID, err
}
