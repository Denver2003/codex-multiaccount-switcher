package ops

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/auth"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

func TestSaveCurrentSkipsDuplicateAuth(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	authData := []byte(`{"email":"user@example.com","tokens":{"access":"abc"},"last_refresh":"one"}`)
	if err := os.WriteFile(authFile, authData, 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	hash, err := auth.Hash(authData)
	if err != nil {
		t.Fatalf("hash auth: %v", err)
	}

	summary := domain.ProfileSummary{
		ID:        "prof_existing",
		Label:     "work",
		Email:     "user@example.com",
		CreatedAt: "2026-04-07T10:00:00Z",
		AuthHash:  hash,
	}

	if err := st.WriteProfile(store.ProfileRecord{Summary: summary}, authData); err != nil {
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

	result, err := SaveCurrent(resolver, "")
	if err != nil {
		t.Fatalf("save current: %v", err)
	}

	if result.Created || result.Duplicate == nil || result.Duplicate.ID != "prof_existing" {
		t.Fatalf("unexpected duplicate result: %#v", result)
	}
}

func TestSaveCurrentCreatesProfileAndMetadata(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)

	if err := os.WriteFile(authFile, []byte(`{"email":"user@example.com","tokens":{"access":"abc"}}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	result, err := SaveCurrent(resolver, "")
	if err != nil {
		t.Fatalf("save current: %v", err)
	}

	if !result.Created || result.Profile.Label != "user@example.com" {
		t.Fatalf("unexpected save result: %#v", result)
	}

	metadata, err := store.New(resolver).ReadMetadata()
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}

	if len(metadata.Profiles) != 1 || metadata.Profiles[0].ID != result.Profile.ID {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestSwitchProfileReplacesLiveAuthAndUpdatesMetadata(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	oldAuth := []byte(`{"tokens":{"access":"old"}}`)
	if err := os.WriteFile(authFile, oldAuth, 0o600); err != nil {
		t.Fatalf("write live auth: %v", err)
	}

	newAuth := []byte(`{"tokens":{"access":"new"}}`)
	hash, err := auth.Hash(newAuth)
	if err != nil {
		t.Fatalf("hash new auth: %v", err)
	}

	summary := domain.ProfileSummary{
		ID:        "prof_target",
		Label:     "work",
		CreatedAt: "2026-04-07T10:00:00Z",
		AuthHash:  hash,
	}

	if err := st.WriteProfile(store.ProfileRecord{Summary: summary}, newAuth); err != nil {
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

	result, err := SwitchProfile(resolver, "work")
	if err != nil {
		t.Fatalf("switch profile: %v", err)
	}

	liveAuth, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("read live auth: %v", err)
	}

	if string(liveAuth) != string(newAuth) {
		t.Fatalf("expected switched auth, got %s", liveAuth)
	}

	if !result.BackupCreated || result.BackupID == "" {
		t.Fatalf("expected backup, got %#v", result)
	}

	metadata, err := st.ReadMetadata()
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}

	if metadata.CurrentProfileID != "prof_target" || metadata.LastSwitchAt == "" || metadata.Profiles[0].LastUsedAt == "" {
		t.Fatalf("unexpected metadata after switch: %#v", metadata)
	}
}

func TestSwitchProfileRejectsCorruptedStoredProfile(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	authData := []byte(`{"tokens":{"access":"new"}}`)
	summary := domain.ProfileSummary{
		ID:        "prof_target",
		Label:     "work",
		CreatedAt: "2026-04-07T10:00:00Z",
		AuthHash:  "sha256:deadbeef",
	}

	if err := st.WriteProfile(store.ProfileRecord{Summary: summary}, authData); err != nil {
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

	_, err := SwitchProfile(resolver, "work")
	if !errors.Is(err, domain.ErrProfileCorrupt) {
		t.Fatalf("expected ErrProfileCorrupt, got %v", err)
	}
}

func TestSaveCurrentRejectsLabelConflictCaseInsensitive(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	authFile := filepath.Join(configDir, "live-auth.json")
	resolver := config.NewResolver(configDir, authFile)
	st := store.New(resolver)

	if err := os.WriteFile(authFile, []byte(`{"tokens":{"access":"abc"}}`), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	existing := domain.ProfileSummary{
		ID:        "prof_existing",
		Label:     "Work",
		CreatedAt: "2026-04-07T10:00:00Z",
		AuthHash:  "sha256:other",
	}

	if err := st.WriteMetadata(domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles:      []domain.ProfileSummary{existing},
	}); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	_, err := SaveCurrent(resolver, "work")
	if !errors.Is(err, domain.ErrLabelConflict) {
		t.Fatalf("expected ErrLabelConflict, got %v", err)
	}
}

func TestRestartGuidanceIsNonEmpty(t *testing.T) {
	t.Parallel()

	if strings.TrimSpace(restartGuidance()) == "" {
		t.Fatal("expected non-empty restart guidance")
	}
}
