package cli

import (
	"flag"
	"fmt"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/ops"
)

func runRemove(application *app.App, args []string) error {
	flags := flag.NewFlagSet("remove", flag.ContinueOnError)
	flags.SetOutput(application.Stderr)

	var yes bool
	var noInput bool

	flags.BoolVar(&yes, "yes", false, "skip confirmation")
	flags.BoolVar(&noInput, "no-input", false, "disable interactive prompts")

	if err := flags.Parse(args); err != nil {
		return errUsageMessage("remove")
	}

	if len(flags.Args()) != 1 {
		return fmt.Errorf("%w: remove requires exactly one profile selector", errUsageMessage("remove"))
	}

	profile, err := ops.RemoveProfile(application.Paths, flags.Args()[0], ops.RemoveOptions{
		Yes:         yes,
		NoInput:     noInput,
		Interactive: application.Interactive && !noInput,
		Input:       application.Stdin,
	})
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(application.Stdout, "Removed profile: %s\n", profile.Label)
	_, _ = fmt.Fprintf(application.Stdout, "ID: %s\n", profile.ID)
	return nil
}
