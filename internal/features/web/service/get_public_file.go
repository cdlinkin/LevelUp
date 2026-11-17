package web_service

import (
	"fmt"
	"os"
	"path"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

var publicFiles = map[string]struct{}{
	"app.css":            {},
	"tokens.css":         {},
	"base.css":           {},
	"cotton.svg":         {},
	"felt.svg":           {},
	"brand/mark.svg":     {},
	"brand/wordmark.svg": {},
}

func (s *WebService) GetPublicFile(rel string) ([]byte, error) {
	if _, ok := publicFiles[rel]; !ok {
		return nil, fmt.Errorf("public file %s: %w", rel, core_errors.ErrNotFound)
	}

	full := path.Join(os.Getenv("PROJECT_ROOT"), "public", rel)
	file, err := s.webRepository.GetFile(full)
	if err != nil {
		return nil, fmt.Errorf("get public file: %w", err)
	}

	return file, nil
}
