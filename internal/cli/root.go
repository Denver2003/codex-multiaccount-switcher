package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/app"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

type command struct {
	name        string
	usage       string
	description string
	run         func(*app.App, []string) error
}

func Main(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	rootFlags := flag.NewFlagSet("codex-switcher", flag.ContinueOnError)
	rootFlags.SetOutput(io.Discard)

	var configDir string
	var authFile string
	var verbose bool

	rootFlags.StringVar(&configDir, "config-dir", "", "override config directory")
	rootFlags.StringVar(&authFile, "auth-file", "", "override active auth file path")
	rootFlags.BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootFlags.Usage = func() {}

	if err := rootFlags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout, registeredCommands())
			return exitSuccess
		}

		printUsage(stderr, registeredCommands())
		return exitUsage
	}

	resolver := config.NewResolver(configDir, authFile)
	application := app.New(resolver, stdout, stderr)
	_ = verbose

	remainingArgs := rootFlags.Args()
	if len(remainingArgs) == 0 {
		printUsage(stdout, registeredCommands())
		return exitSuccess
	}

	commands := registeredCommands()
	selected, ok := commands[remainingArgs[0]]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n\n", remainingArgs[0])
		printUsage(stderr, commands)
		return exitUsage
	}

	if err := selected.run(application, remainingArgs[1:]); err != nil {
		if errors.Is(err, domain.ErrUsage) {
			_, _ = fmt.Fprintln(stderr, err)
			_, _ = fmt.Fprintln(stderr)
			_, _ = fmt.Fprintf(stderr, "Usage: codex-switcher %s\n", selected.usage)
			return exitUsage
		}

		_, _ = fmt.Fprintln(stderr, err)
		return exitFailure
	}

	return exitSuccess
}

func registeredCommands() map[string]command {
	return map[string]command{
		"add": {
			name:        "add",
			usage:       "add",
			description: "Prepare the environment for logging into another Codex account.",
			run:         notImplemented("add"),
		},
		"list": {
			name:        "list",
			usage:       "list",
			description: "List saved profiles.",
			run:         notImplemented("list"),
		},
		"remove": {
			name:        "remove",
			usage:       "remove <profile>",
			description: "Remove a saved profile.",
			run:         notImplemented("remove"),
		},
		"rename": {
			name:        "rename",
			usage:       "rename <profile> <new-label>",
			description: "Rename a saved profile.",
			run:         notImplemented("rename"),
		},
		"save-current": {
			name:        "save-current",
			usage:       "save-current [--label <value>]",
			description: "Save the current active auth as a reusable profile.",
			run:         notImplemented("save-current"),
		},
		"status": {
			name:        "status",
			usage:       "status",
			description: "Inspect the current auth state and local profile store.",
			run:         notImplemented("status"),
		},
		"switch": {
			name:        "switch",
			usage:       "switch <profile>",
			description: "Switch the active Codex auth to a saved profile.",
			run:         notImplemented("switch"),
		},
	}
}

func notImplemented(name string) func(*app.App, []string) error {
	return func(application *app.App, args []string) error {
		_ = application
		_ = args
		return fmt.Errorf("%s: %w", name, domain.ErrNotImplemented)
	}
}

func printUsage(w io.Writer, commands map[string]command) {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  codex-switcher [--verbose] [--config-dir <path>] [--auth-file <path>] <command> [args]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Commands:")

	for _, name := range names {
		cmd := commands[name]
		_, _ = fmt.Fprintf(w, "  %-14s %s\n", cmd.name, cmd.description)
	}
}
