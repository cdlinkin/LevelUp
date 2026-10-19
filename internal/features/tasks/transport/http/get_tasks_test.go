package tasks_transport_http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	tasks_transport_http "github.com/cdlinkin/LevelUp/internal/features/tasks/transport/http"
	"go.uber.org/zap"
)

type fakeTasksService struct {
	getTasksCalled bool
	gotUserID      *int
	gotLimit       *int
	gotOffset      *int

	getTasksResult []domain.Task
}

func (f *fakeTasksService) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	return task, nil
}

func (f *fakeTasksService) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	f.getTasksCalled = true
	f.gotUserID, f.gotLimit, f.gotOffset = userID, limit, offset
	return f.getTasksResult, nil
}

func (f *fakeTasksService) GetTask(ctx context.Context, userID int, id int) (domain.Task, error) {
	return domain.Task{}, nil
}

func (f *fakeTasksService) DeleteTask(ctx context.Context, userID int, id int) error {
	return nil
}

func (f *fakeTasksService) PatchTask(ctx context.Context, userID int, id int, patch domain.TaskPatch) (domain.Task, error) {
	return domain.Task{}, nil
}

func newTestRouter(service tasks_transport_http.TasksService) http.Handler {
	handler := tasks_transport_http.NewTasksHTTPHandler(service)

	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.HandleFunc(
			route.Method+" "+route.Path,
			route.Handler,
		)
	}

	return mux
}

func newRequest(method, target string, userID int) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	log := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	ctx := core_logger.ToContext(r.Context(), log)
	ctx = core_auth.ToContext(ctx, userID)

	return r.WithContext(ctx)
}

func newAnonymousRequest(method, target string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	log := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	return r.WithContext(core_logger.ToContext(r.Context(), log))
}

func intValue(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestGetTasks_QueryParams(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantCalled bool
		wantUserID any
		wantLimit  any
		wantOffset any
	}{
		{
			name:       "user_id from token, query user_id ignored",
			target:     "/tasks?user_id=1&limit=10&offset=5",
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantUserID: 7,
			wantLimit:  10,
			wantOffset: 5,
		},
		{
			name:       "no params",
			target:     "/tasks",
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantUserID: 7,
			wantLimit:  nil,
			wantOffset: nil,
		},
		{
			name:       "limit is not a number",
			target:     "/tasks?limit=abc",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeTasksService{}
			router := newTestRouter(service)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, newRequest(http.MethodGet, tt.target, 7))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if service.getTasksCalled != tt.wantCalled {
				t.Fatalf("service GetTasks called = %v, want %v", service.getTasksCalled, tt.wantCalled)
			}
			if !tt.wantCalled {
				return
			}

			if got := intValue(service.gotUserID); got != tt.wantUserID {
				t.Errorf("userID = %v, want %v", got, tt.wantUserID)
			}
			if got := intValue(service.gotLimit); got != tt.wantLimit {
				t.Errorf("limit = %v, want %v", got, tt.wantLimit)
			}
			if got := intValue(service.gotOffset); got != tt.wantOffset {
				t.Errorf("offset = %v, want %v", got, tt.wantOffset)
			}
		})
	}
}

func TestGetTasks_Anonymous(t *testing.T) {
	service := &fakeTasksService{}
	router := newTestRouter(service)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, newAnonymousRequest(http.MethodGet, "/tasks"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if service.getTasksCalled {
		t.Errorf("service GetTasks was called without user, want not called")
	}
}

func TestGetTasks_ResponseBody(t *testing.T) {
	service := &fakeTasksService{
		getTasksResult: []domain.Task{
			{
				ID:           7,
				Title:        "do homework",
				AuthorUserID: 1,
			},
		},
	}
	router := newTestRouter(service)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, newRequest(http.MethodGet, "/tasks", 1))

	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}

	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("len(body) = %d, want 1", len(body))
	}
	if body[0]["id"] != float64(7) || body[0]["title"] != "do homework" {
		t.Errorf("body[0] = %v, want id=7 title=%q", body[0], "do homework")
	}
}
