package cli

import (
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runStatus(application *app.App, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: status does not accept positional arguments", errUsageMessage("status"))
	}

	status, err := ops.GetStatus(application.Paths)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(application.Stdout, "Active auth: %s\n", status.AuthState)
	_, _ = fmt.Fprintf(application.Stdout, "Auth file: %s\n", status.AuthFile)
	_, _ = fmt.Fprintf(application.Stdout, "Profile store: %s\n", status.ProfilesDir)
	_, _ = fmt.Fprintf(application.Stdout, "Saved profiles: %d\n", status.SavedCount)

	if status.MatchedProfile == nil {
		_, _ = fmt.Fprintln(application.Stdout, "Known profile match: no")
	} else {
		_, _ = fmt.Fprintln(application.Stdout, "Known profile match: yes")
		_, _ = fmt.Fprintf(application.Stdout, "Current profile: %s (%s)\n", status.MatchedProfile.Label, status.MatchedProfile.ID)
	}

	if status.LastSwitchAt != "" {
		_, _ = fmt.Fprintf(application.Stdout, "Last switch: %s\n", status.LastSwitchAt)
	}

	return nil
}
