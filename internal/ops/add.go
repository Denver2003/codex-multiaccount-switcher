package ops

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

type AddOptions struct {
	SaveCurrent bool
	Label       string
	NoInput     bool
	Interactive bool
	Input       io.Reader
}

type AddResult struct {
	AuthRemoved      bool
	BackupCreated    bool
	BackupID         string
	SavedCurrent     bool
	SavedDuplicate   *domain.ProfileSummary
	SavedProfile     *domain.ProfileSummary
	NextInstructions []string
}

func PrepareAdd(paths *config.Resolver, options AddOptions) (AddResult, error) {
	authPath, err := paths.AuthFile()
	if err != nil {
		return AddResult{}, err
	}

	rawAuth, err := auth.ReadFile(authPath)
	if err != nil {
		if errors.Is(err, domain.ErrAuthNotFound) {
			return AddResult{NextInstructions: addInstructions()}, nil
		}

		return AddResult{}, err
	}

	result := AddResult{}
	if options.SaveCurrent {
		saveResult, saveErr := SaveCurrent(paths, options.Label)
		if saveErr != nil {
			return AddResult{}, saveErr
		}

		result.SavedCurrent = saveResult.Created
		if saveResult.Duplicate != nil {
			result.SavedDuplicate = saveResult.Duplicate
		} else {
			result.SavedProfile = &saveResult.Profile
		}
	} else {
		if !options.Interactive || options.NoInput {
			return AddResult{}, fmt.Errorf("%w: active auth exists; rerun with --save-current or in interactive mode", domain.ErrUsage)
		}

		confirmed, confirmErr := confirm(options.Input)
		if confirmErr != nil {
			return AddResult{}, confirmErr
		}
		if !confirmed {
			return AddResult{}, fmt.Errorf("%w: add aborted by user", domain.ErrUsage)
		}
	}

	st := store.New(paths)
	backupID, _, err := st.CreateBackup(rawAuth)
	if err != nil {
		return AddResult{}, err
	}

	if err := os.Remove(authPath); err != nil {
		return AddResult{}, fmt.Errorf("remove live auth %s: %w", authPath, err)
	}

	result.AuthRemoved = true
	result.BackupCreated = true
	result.BackupID = backupID
	result.NextInstructions = addInstructions()
	return result, nil
}

func confirm(input io.Reader) (bool, error) {
	if input == nil {
		return false, fmt.Errorf("%w: confirmation input unavailable", domain.ErrUsage)
	}

	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	if errors.Is(err, io.EOF) && line == "" {
		return false, nil
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func addInstructions() []string {
	return []string{
		"Authenticate manually in Codex CLI or the Codex desktop app.",
		"Verify login completed successfully.",
		"Run `codex-switcher save-current` to store the new account.",
	}
}
