package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", domain.ErrAuthNotFound, path)
		}

		return nil, fmt.Errorf("%w: %s: %w", domain.ErrAuthUnreadable, path, err)
	}

	return data, nil
}
