package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

func TestStatusEmptyStorage(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")

	stdout, stderr, exitCode := runCLI(t, configDir, authFile, "status")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stderr=%q", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Active auth: missing") {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	if !strings.Contains(stdout, "Saved profiles: 0") {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestStatusReportsMatchingProfile(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	authData := []byte(`{"tokens":{"access":"abc"},"last_refresh":"2026-04-07T10:00:00Z"}`)
	if err := os.WriteFile(authFile, authData, 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	authHash, err := auth.Hash(authData)
	if err != nil {
		t.Fatalf("hash auth: %v", err)
	}

	record := store.ProfileRecord{
		Summary: domain.ProfileSummary{
			ID:         "prof_match",
			Label:      "work",
			CreatedAt:  "2026-04-07T10:00:00Z",
			LastUsedAt: "2026-04-07T10:00:00Z",
			AuthHash:   authHash,
		},
	}

	if err := st.WriteProfile(record, authData); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if err := st.WriteMetadata(domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		LastSwitchAt:  "2026-04-07T10:00:00Z",
		Profiles:      []domain.ProfileSummary{record.Summary},
	}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	stdout, stderr, exitCode := runCLI(t, configDir, authFile, "status")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stderr=%q", exitCode, stderr)
	}

	for _, snippet := range []string{
		"Active auth: valid",
		"Known profile match: yes",
		"Current profile: work (prof_match)",
		"Last switch: 2026-04-07T10:00:00Z",
	} {
		if !strings.Contains(stdout, snippet) {
			t.Fatalf("expected %q in stdout, got %q", snippet, stdout)
		}
	}
}

func TestStatusGracefullyReportsInvalidAuth(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")

	if err := os.WriteFile(authFile, []byte(`{"tokens":`), 0o600); err != nil {
		t.Fatalf("write invalid auth: %v", err)
	}

	stdout, stderr, exitCode := runCLI(t, configDir, authFile, "status")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stderr=%q", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Active auth: invalid") {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestListEmptyStorage(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")

	stdout, stderr, exitCode := runCLI(t, configDir, authFile, "list")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stderr=%q", exitCode, stderr)
	}

	if strings.TrimSpace(stdout) != "No saved profiles." {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestListOutputsProfilesSortedByLabelThenID(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	metadata := domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles: []domain.ProfileSummary{
			{ID: "prof_b", Label: "Zulu", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: "sha256:1"},
			{ID: "prof_a2", Label: "alpha", CreatedAt: "2026-04-07T10:00:00Z", LastUsedAt: "2026-04-07T10:05:00Z", Email: "a@example.com", AuthHash: "sha256:2"},
			{ID: "prof_a1", Label: "Alpha", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: "sha256:3"},
		},
	}

	if err := st.WriteMetadata(metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	stdout, stderr, exitCode := runCLI(t, configDir, authFile, "list")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stderr=%q", exitCode, stderr)
	}

	alphaIndex := strings.Index(stdout, "Alpha (prof_a1)")
	alphaLowerIndex := strings.Index(stdout, "alpha (prof_a2)")
	zuluIndex := strings.Index(stdout, "Zulu (prof_b)")
	if alphaIndex < 0 || alphaLowerIndex < 0 || zuluIndex < 0 {
		t.Fatalf("missing expected profiles in stdout: %q", stdout)
	}

	if alphaIndex >= alphaLowerIndex || alphaLowerIndex >= zuluIndex {
		t.Fatalf("unexpected sort order: %q", stdout)
	}

	if !strings.Contains(stdout, "Email: a@example.com") || !strings.Contains(stdout, "Last used: 2026-04-07T10:05:00Z") {
		t.Fatalf("missing expected profile details: %q", stdout)
	}
}

func runCLI(t *testing.T, configDir, authFile string, args ...string) (string, string, int) {
	t.Helper()

	fullArgs := append([]string{"--config-dir", configDir, "--auth-file", authFile}, args...)
	var stdout strings.Builder
	var stderr strings.Builder

	exitCode := run(fullArgs, strings.NewReader(""), &stdout, &stderr)
	return stdout.String(), stderr.String(), exitCode
}
