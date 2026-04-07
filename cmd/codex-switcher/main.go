package main

import (
	"os"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
