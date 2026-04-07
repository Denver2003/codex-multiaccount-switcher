package ops

import (
	"errors"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

type SaveCurrentResult struct {
	Created   bool
	Profile   domain.ProfileSummary
	Source    string
	Duplicate *domain.ProfileSummary
}

func SaveCurrent(paths *config.Resolver, label string) (SaveCurrentResult, error) {
	authPath, err := paths.AuthFile()
	if err != nil {
		return SaveCurrentResult{}, err
	}

	rawAuth, err := auth.ReadFile(authPath)
	if err != nil {
		return SaveCurrentResult{}, err
	}

	if err := auth.Validate(rawAuth); err != nil {
		return SaveCurrentResult{}, err
	}

	authHash, err := auth.Hash(rawAuth)
	if err != nil {
		return SaveCurrentResult{}, err
	}

	st := store.New(paths)
	metadata, err := st.ReadMetadata()
	if err != nil {
		if !errors.Is(err, domain.ErrMetadataNotFound) {
			return SaveCurrentResult{}, err
		}

		now := time.Now().UTC().Format(time.RFC3339)
		metadata = domain.Metadata{
			SchemaVersion: 1,
			CreatedAt:     now,
			UpdatedAt:     now,
			Profiles:      []domain.ProfileSummary{},
		}
	}

	for _, profile := range metadata.Profiles {
		if profile.AuthHash == authHash {
			profileCopy := profile
			return SaveCurrentResult{
				Created:   false,
				Profile:   profile,
				Source:    authPath,
				Duplicate: &profileCopy,
			}, nil
		}
	}

	email, err := auth.ExtractEmail(rawAuth)
	if err != nil {
		return SaveCurrentResult{}, err
	}

	resolvedLabel, err := resolveLabel(label, email, metadata.Profiles)
	if err != nil {
		return SaveCurrentResult{}, err
	}

	if err := ensureUniqueLabel(resolvedLabel, metadata.Profiles, ""); err != nil {
		return SaveCurrentResult{}, err
	}

	profileID, err := store.NewProfileID()
	if err != nil {
		return SaveCurrentResult{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	profile := domain.ProfileSummary{
		ID:        profileID,
		Label:     resolvedLabel,
		Email:     email,
		CreatedAt: now,
		AuthHash:  authHash,
	}

	if err := st.WriteProfile(store.ProfileRecord{Summary: profile}, rawAuth); err != nil {
		return SaveCurrentResult{}, err
	}

	metadata.Profiles = append(metadata.Profiles, profile)
	metadata.UpdatedAt = now
	if metadata.CreatedAt == "" {
		metadata.CreatedAt = now
	}

	if err := st.WriteMetadata(metadata); err != nil {
		return SaveCurrentResult{}, err
	}

	return SaveCurrentResult{
		Created: true,
		Profile: profile,
		Source:  authPath,
	}, nil
}
