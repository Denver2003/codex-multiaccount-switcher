package ops

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/fsx"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

type SwitchResult struct {
	Profile       domain.ProfileSummary
	BackupCreated bool
	BackupID      string
	RestartAdvice string
}

func SwitchProfile(paths *config.Resolver, selector string) (SwitchResult, error) {
	if strings.TrimSpace(selector) == "" {
		return SwitchResult{}, fmt.Errorf("%w: missing profile selector", domain.ErrUsage)
	}

	st := store.New(paths)
	metadata, err := st.ReadMetadata()
	if err != nil {
		return SwitchResult{}, err
	}

	target, index, err := resolveProfileSelector(metadata.Profiles, selector)
	if err != nil {
		return SwitchResult{}, err
	}

	record, authData, err := st.ReadProfile(target.ID)
	if err != nil {
		return SwitchResult{}, err
	}

	if err := auth.Validate(authData); err != nil {
		return SwitchResult{}, err
	}

	recomputedHash, err := auth.Hash(authData)
	if err != nil {
		return SwitchResult{}, err
	}

	if recomputedHash != record.Summary.AuthHash {
		return SwitchResult{}, fmt.Errorf("%w: stored auth hash mismatch for profile %s", domain.ErrProfileCorrupt, target.ID)
	}

	authPath, err := paths.AuthFile()
	if err != nil {
		return SwitchResult{}, err
	}

	result := SwitchResult{
		Profile:       record.Summary,
		RestartAdvice: restartGuidance(),
	}

	if liveAuth, readErr := auth.ReadFile(authPath); readErr == nil {
		backupID, _, backupErr := st.CreateBackup(liveAuth)
		if backupErr != nil {
			return SwitchResult{}, backupErr
		}
		result.BackupCreated = true
		result.BackupID = backupID
	} else if !errors.Is(readErr, domain.ErrAuthNotFound) {
		return SwitchResult{}, readErr
	}

	if err := fsx.AtomicWriteFile(authPath, authData, 0o600); err != nil {
		return SwitchResult{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	record.Summary.LastUsedAt = now
	metadata.Profiles[index] = record.Summary
	metadata.UpdatedAt = now
	metadata.LastSwitchAt = now
	metadata.CurrentProfileID = record.Summary.ID

	if err := st.WriteProfile(record, authData); err != nil {
		return SwitchResult{}, fmt.Errorf("switch succeeded but profile update failed: %w", err)
	}

	if err := st.WriteMetadata(metadata); err != nil {
		return SwitchResult{}, fmt.Errorf("switch succeeded but metadata update failed: %w", err)
	}

	result.Profile = record.Summary
	return result, nil
}

func resolveProfileSelector(profiles []domain.ProfileSummary, selector string) (domain.ProfileSummary, int, error) {
	lowerSelector := strings.ToLower(selector)
	for index, profile := range profiles {
		if strings.ToLower(profile.Label) == lowerSelector {
			return profile, index, nil
		}
	}

	for index, profile := range profiles {
		if profile.ID == selector {
			return profile, index, nil
		}
	}

	return domain.ProfileSummary{}, -1, fmt.Errorf("%w: %s", domain.ErrProfileNotFound, selector)
}

func restartGuidance() string {
	if runtime.GOOS == "darwin" {
		return "Existing Codex CLI sessions may still use cached auth. Restart active Codex CLI sessions and the Codex desktop app if needed."
	}

	return "Existing Codex CLI sessions may still use cached auth. Restart active Codex CLI sessions if behavior does not reflect the new account."
}
