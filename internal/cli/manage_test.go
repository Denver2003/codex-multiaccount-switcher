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

func TestAddNoInputFailsWhenConfirmationIsRequired(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"tokens":{"access":"abc"}}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	stdout, stderr, exitCode := runCLIWithInput(t, "", "--config-dir", configDir, "--auth-file", authFile, "add", "--no-input")
	if exitCode != exitUsage {
		t.Fatalf("expected usage exit, got %d stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
}

func TestAddSaveCurrentBacksUpAndRemovesLiveAuth(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "auth.json")
	resolver := config.NewResolver(configDir, authFile)

	if err := os.WriteFile(authFile, []byte(`{"email":"user@example.com","tokens":{"access":"abc"}}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	stdout, stderr, exitCode := runCLIWithInput(t, "", "--config-dir", configDir, "--auth-file", authFile, "add", "--save-current")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if _, err := os.Stat(authFile); !os.IsNotExist(err) {
		t.Fatalf("expected auth file to be removed, got err=%v", err)
	}

	profiles, err := store.New(resolver).VisibleProfiles()
	if err != nil {
		t.Fatalf("visible profiles: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected one saved profile, got %#v", profiles)
	}
}

func TestRenameRejectsCaseInsensitiveConflict(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	summaryA := domain.ProfileSummary{ID: "prof_a", Label: "Work", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: "sha256:a"}
	summaryB := domain.ProfileSummary{ID: "prof_b", Label: "Personal", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: "sha256:b"}

	if err := st.WriteProfile(store.ProfileRecord{Summary: summaryA}, []byte(`{"tokens":{"access":"a"}}`)); err != nil {
		t.Fatalf("write profile A: %v", err)
	}
	if err := st.WriteProfile(store.ProfileRecord{Summary: summaryB}, []byte(`{"tokens":{"access":"b"}}`)); err != nil {
		t.Fatalf("write profile B: %v", err)
	}
	if err := st.WriteMetadata(domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles:      []domain.ProfileSummary{summaryA, summaryB},
	}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	stdout, stderr, exitCode := runCLIWithInput(t, "", "--config-dir", configDir, "--auth-file", authFile, "rename", "prof_b", "work")
	if exitCode != exitFailure {
		t.Fatalf("expected failure, got %d stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
}

func TestRemoveDeletesProfileAndUpdatesMetadata(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	authData := []byte(`{"tokens":{"access":"abc"}}`)
	hash, err := auth.Hash(authData)
	if err != nil {
		t.Fatalf("hash auth: %v", err)
	}

	summary := domain.ProfileSummary{ID: "prof_a", Label: "Work", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: hash}
	if err := st.WriteProfile(store.ProfileRecord{Summary: summary}, authData); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	if err := st.WriteMetadata(domain.Metadata{
		SchemaVersion:    1,
		CreatedAt:        "2026-04-07T10:00:00Z",
		UpdatedAt:        "2026-04-07T10:00:00Z",
		CurrentProfileID: "prof_a",
		Profiles:         []domain.ProfileSummary{summary},
	}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	stdout, stderr, exitCode := runCLIWithInput(t, "", "--config-dir", configDir, "--auth-file", authFile, "remove", "--yes", "prof_a")
	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if strings.TrimSpace(stdout) == "" {
		t.Fatalf("expected output, got empty stdout")
	}

	if _, err := os.Stat(filepath.Join(configDir, "profiles", "prof_a")); !os.IsNotExist(err) {
		t.Fatalf("expected profile directory removed, got err=%v", err)
	}

	metadata, err := st.ReadMetadata()
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if len(metadata.Profiles) != 0 || metadata.CurrentProfileID != "" {
		t.Fatalf("unexpected metadata after remove: %#v", metadata)
	}
}

func TestRemoveTreatsEOFAsRefusal(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	summary := domain.ProfileSummary{ID: "prof_a", Label: "Work", CreatedAt: "2026-04-07T10:00:00Z", AuthHash: "sha256:a"}
	if err := st.WriteProfile(store.ProfileRecord{Summary: summary}, []byte(`{"tokens":{"access":"a"}}`)); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	if err := st.WriteMetadata(domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles:      []domain.ProfileSummary{summary},
	}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	stdout, stderr, exitCode := runCLIWithInput(t, "", "--config-dir", configDir, "--auth-file", authFile, "remove", "prof_a")
	if exitCode != exitUsage {
		t.Fatalf("expected usage exit, got %d stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
}
