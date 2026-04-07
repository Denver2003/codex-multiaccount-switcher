package ops

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func resolveLabel(explicit, email string, profiles []domain.ProfileSummary) (string, error) {
	if explicit != "" {
		return validateLabel(explicit)
	}

	if email != "" {
		return validateLabel(email)
	}

	return nextFallbackLabel(profiles), nil
}

func validateLabel(label string) (string, error) {
	trimmed := strings.TrimSpace(label)
	switch {
	case trimmed == "":
		return "", fmt.Errorf("%w: label must not be empty", domain.ErrInvalidLabel)
	case len(trimmed) > 64:
		return "", fmt.Errorf("%w: label must be 1..64 characters", domain.ErrInvalidLabel)
	case strings.Contains(trimmed, "/") || strings.Contains(trimmed, `\`):
		return "", fmt.Errorf("%w: label must not contain path separators", domain.ErrInvalidLabel)
	case !utf8.ValidString(trimmed):
		return "", fmt.Errorf("%w: label must be valid UTF-8", domain.ErrInvalidLabel)
	default:
		return trimmed, nil
	}
}

func ensureUniqueLabel(label string, profiles []domain.ProfileSummary, exceptID string) error {
	normalized := strings.ToLower(label)
	for _, profile := range profiles {
		if exceptID != "" && profile.ID == exceptID {
			continue
		}

		if strings.ToLower(profile.Label) == normalized {
			return fmt.Errorf("%w: %s", domain.ErrLabelConflict, label)
		}
	}

	return nil
}

func nextFallbackLabel(profiles []domain.ProfileSummary) string {
	used := make(map[string]struct{}, len(profiles))
	for _, profile := range profiles {
		used[strings.ToLower(profile.Label)] = struct{}{}
	}

	for index := 1; ; index++ {
		label := fmt.Sprintf("account-%d", index)
		if _, exists := used[strings.ToLower(label)]; !exists {
			return label
		}
	}
}
