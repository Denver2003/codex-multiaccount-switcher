package cli

import (
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runRename(application *app.App, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("%w: rename requires <profile> <new-label>", errUsageMessage("rename"))
	}

	profile, err := ops.RenameProfile(application.Paths, args[0], args[1])
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(application.Stdout, "Renamed profile: %s\n", profile.Label)
	_, _ = fmt.Fprintf(application.Stdout, "ID: %s\n", profile.ID)
	return nil
}
