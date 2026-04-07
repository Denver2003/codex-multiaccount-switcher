package cli

import (
	"strings"
	"testing"
)

func runCLIWithInput(t *testing.T, input string, args ...string) (string, string, int) {
	t.Helper()

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := run(args, strings.NewReader(input), &stdout, &stderr)
	return stdout.String(), stderr.String(), exitCode
}
