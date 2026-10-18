package statistics_transport_http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	statistics_transport_http "github.com/cdlinkin/LevelUp/internal/features/statistics/transport/http"
	"go.uber.org/zap"
)

type fakeStatisticsService struct {
	called    bool
	gotUserID *int
	gotFrom   *time.Time
	gotTo     *time.Time
}

func (f *fakeStatisticsService) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, error) {
	f.called = true
	f.gotUserID, f.gotFrom, f.gotTo = userID, from, to
	return domain.Statistics{}, nil
}

func newTestRouter(service statistics_transport_http.StatisticsService) http.Handler {
	handler := statistics_transport_http.NewStatisticsHTTPHandler(service)

	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.HandleFunc(
			route.Method+" "+route.Path,
			route.Handler,
		)
	}

	return mux
}

func newRequest(method, target string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	log := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	return r.WithContext(core_logger.ToContext(r.Context(), log))
}

func date(year int, month time.Month, day int) *time.Time {
	d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &d
}

func assertTime(t *testing.T, field string, got, want *time.Time) {
	t.Helper()

	if got == nil || want == nil {
		if got != want {
			t.Errorf("%s = %v, want %v", field, got, want)
		}
		return
	}

	if !got.Equal(*want) {
		t.Errorf("%s = %v, want %v", field, *got, *want)
	}
}

func TestGetStatistics_QueryParams(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantCalled bool
		wantFrom   *time.Time
		wantTo     *time.Time
	}{
		{
			name:       "from and to",
			target:     "/statistics?from=2026-10-01&to=2026-10-20",
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantFrom:   date(2026, time.October, 1),
			wantTo:     date(2026, time.October, 20),
		},
		{
			name:       "only to",
			target:     "/statistics?to=2026-10-20",
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantFrom:   nil,
			wantTo:     date(2026, time.October, 20),
		},
		{
			name:       "no params",
			target:     "/statistics",
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantFrom:   nil,
			wantTo:     nil,
		},
		{
			name:       "invalid date format",
			target:     "/statistics?from=01.10.2026",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeStatisticsService{}
			router := newTestRouter(service)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, newRequest(http.MethodGet, tt.target))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if service.called != tt.wantCalled {
				t.Fatalf("service GetStatistics called = %v, want %v", service.called, tt.wantCalled)
			}
			if !tt.wantCalled {
				return
			}

			assertTime(t, "from", service.gotFrom, tt.wantFrom)
			assertTime(t, "to", service.gotTo, tt.wantTo)
		})
	}
}

func TestGetStatistics_UserID(t *testing.T) {
	service := &fakeStatisticsService{}
	router := newTestRouter(service)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, newRequest(http.MethodGet, "/statistics?user_id=3"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if service.gotUserID == nil || *service.gotUserID != 3 {
		t.Errorf("userID = %v, want 3", service.gotUserID)
	}
}
