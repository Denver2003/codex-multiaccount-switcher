package auth

import (
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func Validate(raw []byte) error {
	value, err := parseJSONObject(raw)
	if err != nil {
		return err
	}

	if _, ok := value["tokens"]; ok {
		return nil
	}

	if _, ok := value["OPENAI_API_KEY"]; ok {
		return nil
	}

	return fmt.Errorf("%w: auth must contain top-level \"tokens\" or \"OPENAI_API_KEY\"", domain.ErrInvalidAuth)
}
