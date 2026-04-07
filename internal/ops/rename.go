package ops

import (
	"fmt"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

func RenameProfile(paths *config.Resolver, selector, newLabel string) (domain.ProfileSummary, error) {
	st := store.New(paths)
	metadata, err := st.ReadMetadata()
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	target, index, err := resolveProfileSelector(metadata.Profiles, selector)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	resolvedLabel, err := validateLabel(newLabel)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	if err := ensureUniqueLabel(resolvedLabel, metadata.Profiles, target.ID); err != nil {
		return domain.ProfileSummary{}, err
	}

	record, authData, err := st.ReadProfile(target.ID)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	record.Summary.Label = resolvedLabel
	metadata.Profiles[index] = record.Summary
	metadata.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := st.WriteProfile(record, authData); err != nil {
		return domain.ProfileSummary{}, fmt.Errorf("update profile %s: %w", target.ID, err)
	}

	if err := st.WriteMetadata(metadata); err != nil {
		return domain.ProfileSummary{}, err
	}

	return record.Summary, nil
}
