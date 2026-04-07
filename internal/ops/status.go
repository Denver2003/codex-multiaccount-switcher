package ops

import (
	"errors"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

type StatusResult struct {
	AuthState      string
	AuthFile       string
	ProfilesDir    string
	SavedCount     int
	MatchedProfile *domain.ProfileSummary
	LastSwitchAt   string
}

func GetStatus(paths *config.Resolver) (StatusResult, error) {
	result := StatusResult{
		AuthState: "missing",
	}

	authFile, err := paths.AuthFile()
	if err != nil {
		return StatusResult{}, err
	}
	result.AuthFile = authFile

	profilesDir, err := paths.ProfilesDir()
	if err != nil {
		return StatusResult{}, err
	}
	result.ProfilesDir = profilesDir

	st := store.New(paths)
	metadata, err := st.ReadMetadata()
	switch {
	case err == nil:
		result.SavedCount = len(metadata.Profiles)
		result.LastSwitchAt = metadata.LastSwitchAt
	case errors.Is(err, domain.ErrMetadataNotFound):
		result.SavedCount = 0
	default:
		return StatusResult{}, err
	}

	rawAuth, err := auth.ReadFile(authFile)
	switch {
	case err == nil:
	case errors.Is(err, domain.ErrAuthNotFound):
		return result, nil
	default:
		result.AuthState = "unreadable"
		return result, nil
	}

	if err := auth.Validate(rawAuth); err != nil {
		result.AuthState = "invalid"
		return result, nil
	}

	result.AuthState = "valid"

	if metadata.Profiles == nil {
		return result, nil
	}

	authHash, err := auth.Hash(rawAuth)
	if err != nil {
		result.AuthState = "invalid"
		return result, nil
	}

	for _, profile := range metadata.Profiles {
		if profile.AuthHash == authHash {
			profileCopy := profile
			result.MatchedProfile = &profileCopy
			break
		}
	}

	return result, nil
}
