package app

import (
	"io"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
)

type App struct {
	Paths       *config.Resolver
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Interactive bool
}

func New(paths *config.Resolver, stdin io.Reader, stdout, stderr io.Writer, interactive bool) *App {
	return &App{
		Paths:       paths,
		Stdin:       stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		Interactive: interactive,
	}
}
