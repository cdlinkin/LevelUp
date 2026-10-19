package core_http_middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_middleware "github.com/cdlinkin/LevelUp/internal/core/transport/http/middleware"
	"go.uber.org/zap"
)

type fakeTokenParser struct {
	userID int
	err    error

	gotToken string
}

func (f *fakeTokenParser) Parse(token string) (int, error) {
	f.gotToken = token
	return f.userID, f.err
}

func newRequest(authorization string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}

	log := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	return r.WithContext(core_logger.ToContext(r.Context(), log))
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		parser        *fakeTokenParser
		wantStatus    int
		wantNext      bool
		wantUserID    int
		wantToken     string
	}{
		{
			name:          "valid token",
			authorization: "Bearer good-token",
			parser:        &fakeTokenParser{userID: 7},
			wantStatus:    http.StatusOK,
			wantNext:      true,
			wantUserID:    7,
			wantToken:     "good-token",
		},
		{
			name:          "no header",
			authorization: "",
			parser:        &fakeTokenParser{userID: 7},
			wantStatus:    http.StatusUnauthorized,
			wantNext:      false,
		},
		{
			name:          "not bearer scheme",
			authorization: "Basic dXNlcjpwYXNz",
			parser:        &fakeTokenParser{userID: 7},
			wantStatus:    http.StatusUnauthorized,
			wantNext:      false,
		},
		{
			name:          "bearer without token",
			authorization: "Bearer ",
			parser:        &fakeTokenParser{userID: 7},
			wantStatus:    http.StatusUnauthorized,
			wantNext:      false,
		},
		{
			name:          "invalid token",
			authorization: "Bearer bad-token",
			parser:        &fakeTokenParser{err: fmt.Errorf("parse token: %w", core_errors.ErrUnauthorized)},
			wantStatus:    http.StatusUnauthorized,
			wantNext:      false,
			wantToken:     "bad-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			gotUserID := 0

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID, _ = core_auth.UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := core_http_middleware.Auth(tt.parser)(next)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, newRequest(tt.authorization))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if nextCalled != tt.wantNext {
				t.Fatalf("next handler called = %v, want %v", nextCalled, tt.wantNext)
			}
			if tt.parser.gotToken != tt.wantToken {
				t.Errorf("parser got token = %q, want %q", tt.parser.gotToken, tt.wantToken)
			}
			if tt.wantNext && gotUserID != tt.wantUserID {
				t.Errorf("userID in context = %d, want %d", gotUserID, tt.wantUserID)
			}
		})
	}
}
