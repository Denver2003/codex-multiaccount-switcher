package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithoutCommandPrintsUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, strings.NewReader(""), &stdout, &stderr)
	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage in stdout, got %q", stdout.String())
	}
}

func TestRunHelpFlagPrintsUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if exitCode != exitSuccess {
		t.Fatalf("expected exit code %d, got %d", exitSuccess, exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("expected commands list in stdout, got %q", stdout.String())
	}
}
