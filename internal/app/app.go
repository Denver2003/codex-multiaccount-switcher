package app

import (
	"io"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
)

type App struct {
	Paths  *config.Resolver
	Stdout io.Writer
	Stderr io.Writer
}

func New(paths *config.Resolver, stdout, stderr io.Writer) *App {
	return &App{
		Paths:  paths,
		Stdout: stdout,
		Stderr: stderr,
	}
}
