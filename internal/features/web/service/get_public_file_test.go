package web_service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

type fakeWebRepository struct {
	called bool
}

func (f *fakeWebRepository) GetFile(filepath string) ([]byte, error) {
	f.called = true
	return nil, errors.New("should not read")
}

func TestGetPublicFileRejectsUnknown(t *testing.T) {
	repo := &fakeWebRepository{}
	service := NewWebService(repo)

	_, err := service.GetPublicFile("../.env")
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetPublicFile() error = %v, want not found", err)
	}
	if repo.called {
		t.Fatal("repository was called for a rejected path")
	}
}

func TestGetPublicFileReadsStylesheet(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROJECT_ROOT", root)

	service := NewWebService(new(realRepo))
	file, err := service.GetPublicFile("tokens.css")
	if err != nil {
		t.Fatalf("GetPublicFile() error = %v", err)
	}
	if len(file) == 0 {
		t.Fatal("tokens.css is empty")
	}
}

type realRepo struct{}

func (r *realRepo) GetFile(filepath string) ([]byte, error) {
	return os.ReadFile(filepath)
}
