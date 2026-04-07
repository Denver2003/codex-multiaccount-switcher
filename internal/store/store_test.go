package store

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func TestEnsureLayoutCreatesDirectoriesWithExpectedPermissions(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	if err := st.EnsureLayout(); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	assertMode(t, baseDir, 0o700)
	assertMode(t, filepath.Join(baseDir, "profiles"), 0o700)
	assertMode(t, filepath.Join(baseDir, "backups"), 0o700)
}

func TestWriteMetadataCreatesManagedFileWithExpectedPermissions(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	metadata := domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles:      []domain.ProfileSummary{},
	}

	if err := st.WriteMetadata(metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	assertMode(t, filepath.Join(baseDir, "metadata.json"), 0o600)
}

func TestVisibleProfilesIgnoreOrphanedProfileDirectories(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	if err := st.EnsureLayout(); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	metadata := domain.Metadata{
		SchemaVersion: 1,
		CreatedAt:     "2026-04-07T10:00:00Z",
		UpdatedAt:     "2026-04-07T10:00:00Z",
		Profiles: []domain.ProfileSummary{
			{
				ID:       "prof_visible",
				Label:    "visible",
				AuthHash: "sha256:a",
			},
		},
	}

	if err := st.WriteMetadata(metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(baseDir, "profiles", "prof_orphan"), 0o700); err != nil {
		t.Fatalf("create orphan profile dir: %v", err)
	}

	profiles, err := st.VisibleProfiles()
	if err != nil {
		t.Fatalf("visible profiles: %v", err)
	}

	if len(profiles) != 1 || profiles[0].ID != "prof_visible" {
		t.Fatalf("expected only metadata-backed profile, got %#v", profiles)
	}
}

func TestVisibleProfilesFailsWhenMetadataMissingAndOrphanedProfilesExist(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	if err := st.EnsureLayout(); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(baseDir, "profiles", "prof_orphan"), 0o700); err != nil {
		t.Fatalf("create orphan profile dir: %v", err)
	}

	_, err := st.VisibleProfiles()
	if !errors.Is(err, domain.ErrMetadataNotFound) {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestProfileLocalFilesDoNotBecomeVisibleWithoutMetadataCommit(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	record := ProfileRecord{
		Summary: domain.ProfileSummary{
			ID:        "prof_pending",
			Label:     "pending",
			CreatedAt: "2026-04-07T10:00:00Z",
			AuthHash:  "sha256:test",
		},
	}

	if err := st.WriteProfile(record, []byte(`{"tokens":{"access":"abc"}}`)); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	_, err := st.VisibleProfiles()
	if !errors.Is(err, domain.ErrMetadataNotFound) {
		t.Fatalf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestReadWriteProfileRoundTrip(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver)

	record := ProfileRecord{
		Summary: domain.ProfileSummary{
			ID:         "prof_roundtrip",
			Label:      "roundtrip",
			Email:      "user@example.com",
			CreatedAt:  "2026-04-07T10:00:00Z",
			LastUsedAt: "2026-04-07T10:00:00Z",
			AuthHash:   "sha256:abc",
		},
	}
	authData := []byte("{\"tokens\":{\"access\":\"abc\"}}")

	if err := st.WriteProfile(record, authData); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	assertMode(t, filepath.Join(baseDir, "profiles", "prof_roundtrip", "profile.json"), 0o600)
	assertMode(t, filepath.Join(baseDir, "profiles", "prof_roundtrip", "auth.json"), 0o600)

	gotRecord, gotAuth, err := st.ReadProfile("prof_roundtrip")
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}

	if gotRecord.Summary.ID != record.Summary.ID || string(gotAuth) != string(authData) {
		t.Fatalf("unexpected roundtrip result: %#v %s", gotRecord, gotAuth)
	}
}

func TestCreateBackupWritesExpectedFilenameAndPermissions(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolver := config.NewResolver(baseDir, filepath.Join(baseDir, "auth.json"))
	st := New(resolver).WithClock(func() time.Time {
		return time.Date(2026, 4, 7, 10, 15, 30, 0, time.UTC)
	})

	backupID, backupPath, err := st.CreateBackup([]byte(`{"tokens":{"access":"abc"}}`))
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}

	matched, err := regexp.MatchString(`^bkp_20260407T101530Z_[0-9a-f]{6}$`, backupID)
	if err != nil {
		t.Fatalf("compile regexp: %v", err)
	}

	if !matched {
		t.Fatalf("unexpected backup ID: %s", backupID)
	}

	assertMode(t, backupPath, 0o600)
}

func TestNewProfileIDMatchesMVPFormat(t *testing.T) {
	t.Parallel()

	id, err := NewProfileID()
	if err != nil {
		t.Fatalf("new profile ID: %v", err)
	}

	matched, err := regexp.MatchString(`^prof_[0-9a-z]{16}$`, id)
	if err != nil {
		t.Fatalf("compile regexp: %v", err)
	}

	if !matched {
		t.Fatalf("unexpected profile ID: %s", id)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}

	if info.Mode().Perm() != want {
		t.Fatalf("expected mode %o for %s, got %o", want, path, info.Mode().Perm())
	}
}
