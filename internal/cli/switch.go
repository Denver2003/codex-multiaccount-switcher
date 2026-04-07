package cli

import (
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runSwitch(application *app.App, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: switch requires exactly one profile selector", errUsageMessage("switch"))
	}

	result, err := ops.SwitchProfile(application.Paths, args[0])
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(application.Stdout, "Switched to profile: %s\n", result.Profile.Label)
	_, _ = fmt.Fprintf(application.Stdout, "ID: %s\n", result.Profile.ID)
	if result.BackupCreated {
		_, _ = fmt.Fprintf(application.Stdout, "Backup: %s\n", result.BackupID)
	}
	_, _ = fmt.Fprintln(application.Stdout, result.RestartAdvice)
	return nil
}
