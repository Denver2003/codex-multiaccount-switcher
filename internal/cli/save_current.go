package cli

import (
	"flag"
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runSaveCurrent(application *app.App, args []string) error {
	flags := flag.NewFlagSet("save-current", flag.ContinueOnError)
	flags.SetOutput(application.Stderr)

	var label string
	flags.StringVar(&label, "label", "", "profile label override")

	if err := flags.Parse(args); err != nil {
		return errUsageMessage("save-current")
	}

	if len(flags.Args()) != 0 {
		return fmt.Errorf("%w: save-current does not accept positional arguments", errUsageMessage("save-current"))
	}

	result, err := ops.SaveCurrent(application.Paths, label)
	if err != nil {
		return err
	}

	if result.Duplicate != nil {
		_, _ = fmt.Fprintf(application.Stdout, "Profile already saved: %s\n", result.Duplicate.Label)
		_, _ = fmt.Fprintf(application.Stdout, "ID: %s\n", result.Duplicate.ID)
		return nil
	}

	_, _ = fmt.Fprintf(application.Stdout, "Saved profile: %s\n", result.Profile.Label)
	_, _ = fmt.Fprintf(application.Stdout, "ID: %s\n", result.Profile.ID)
	_, _ = fmt.Fprintf(application.Stdout, "Source: %s\n", result.Source)
	return nil
}
