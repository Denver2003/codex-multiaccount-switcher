package cli

import (
	"flag"
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runAdd(application *app.App, args []string) error {
	flags := flag.NewFlagSet("add", flag.ContinueOnError)
	flags.SetOutput(application.Stderr)

	var saveCurrent bool
	var label string
	var noInput bool

	flags.BoolVar(&saveCurrent, "save-current", false, "save current auth before removing it")
	flags.StringVar(&label, "label", "", "label to use with --save-current")
	flags.BoolVar(&noInput, "no-input", false, "disable interactive prompts")

	if err := flags.Parse(args); err != nil {
		return errUsageMessage("add")
	}

	if len(flags.Args()) != 0 {
		return fmt.Errorf("%w: add does not accept positional arguments", errUsageMessage("add"))
	}

	result, err := ops.PrepareAdd(application.Paths, ops.AddOptions{
		SaveCurrent: saveCurrent,
		Label:       label,
		NoInput:     noInput,
		Interactive: application.Interactive && !noInput,
		Input:       application.Stdin,
	})
	if err != nil {
		return err
	}

	if result.SavedProfile != nil {
		_, _ = fmt.Fprintf(application.Stdout, "Saved current profile: %s (%s)\n", result.SavedProfile.Label, result.SavedProfile.ID)
	}
	if result.SavedDuplicate != nil {
		_, _ = fmt.Fprintf(application.Stdout, "Current auth already stored as: %s (%s)\n", result.SavedDuplicate.Label, result.SavedDuplicate.ID)
	}
	if result.BackupCreated {
		_, _ = fmt.Fprintf(application.Stdout, "Backup: %s\n", result.BackupID)
	}
	if result.AuthRemoved {
		_, _ = fmt.Fprintln(application.Stdout, "Removed live auth from the active location.")
	}
	for _, line := range result.NextInstructions {
		_, _ = fmt.Fprintln(application.Stdout, line)
	}

	return nil
}
