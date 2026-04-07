package cli

import (
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runList(application *app.App, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: list does not accept positional arguments", errUsageMessage("list"))
	}

	profiles, err := ops.ListProfiles(application.Paths)
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		_, _ = fmt.Fprintln(application.Stdout, "No saved profiles.")
		return nil
	}

	for _, profile := range profiles {
		_, _ = fmt.Fprintf(application.Stdout, "%s (%s)\n", profile.Label, profile.ID)
		if profile.Email != "" {
			_, _ = fmt.Fprintf(application.Stdout, "  Email: %s\n", profile.Email)
		}
		_, _ = fmt.Fprintf(application.Stdout, "  Created: %s\n", profile.CreatedAt)
		if profile.LastUsedAt != "" {
			_, _ = fmt.Fprintf(application.Stdout, "  Last used: %s\n", profile.LastUsedAt)
		}
	}

	return nil
}
