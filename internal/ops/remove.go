package ops

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

type RemoveOptions struct {
	Yes         bool
	NoInput     bool
	Interactive bool
	Input       io.Reader
}

func RemoveProfile(paths *config.Resolver, selector string, options RemoveOptions) (domain.ProfileSummary, error) {
	st := store.New(paths)
	metadata, err := st.ReadMetadata()
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	target, index, err := resolveProfileSelector(metadata.Profiles, selector)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	if !options.Yes {
		if !options.Interactive || options.NoInput {
			return domain.ProfileSummary{}, fmt.Errorf("%w: remove requires confirmation; rerun with --yes or interactive mode", domain.ErrUsage)
		}

		confirmed, confirmErr := confirm(options.Input)
		if confirmErr != nil {
			return domain.ProfileSummary{}, confirmErr
		}
		if !confirmed {
			return domain.ProfileSummary{}, fmt.Errorf("%w: remove aborted by user", domain.ErrUsage)
		}
	}

	profileDir, err := st.ProfileDir(target.ID)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	if err := os.RemoveAll(profileDir); err != nil {
		return domain.ProfileSummary{}, fmt.Errorf("remove profile directory %s: %w", profileDir, err)
	}

	metadata.Profiles = append(metadata.Profiles[:index], metadata.Profiles[index+1:]...)
	metadata.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if metadata.CurrentProfileID == target.ID {
		metadata.CurrentProfileID = ""
	}

	if err := st.WriteMetadata(metadata); err != nil {
		return domain.ProfileSummary{}, err
	}

	return target, nil
}
