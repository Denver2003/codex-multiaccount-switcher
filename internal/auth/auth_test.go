package auth

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func TestReadFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")
	authBody := []byte(`{"tokens":{"access":"abc"}}`)

	if err := os.WriteFile(authPath, authBody, 0o600); err != nil {
		t.Fatalf("write auth fixture: %v", err)
	}

	got, err := ReadFile(authPath)
	if err != nil {
		t.Fatalf("read auth file: %v", err)
	}

	if string(got) != string(authBody) {
		t.Fatalf("expected %q, got %q", authBody, got)
	}
}

func TestReadFileMissing(t *testing.T) {
	t.Parallel()

	_, err := ReadFile(filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, domain.ErrAuthNotFound) {
		t.Fatalf("expected ErrAuthNotFound, got %v", err)
	}
}

func TestValidateAcceptsTokens(t *testing.T) {
	t.Parallel()

	err := Validate([]byte(`{"tokens":{"access":"abc"},"last_refresh":"2026-04-07T10:00:00Z"}`))
	if err != nil {
		t.Fatalf("expected valid auth, got %v", err)
	}
}

func TestValidateAcceptsAPIKey(t *testing.T) {
	t.Parallel()

	err := Validate([]byte(`{"OPENAI_API_KEY":"secret"}`))
	if err != nil {
		t.Fatalf("expected valid auth, got %v", err)
	}
}

func TestValidateRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	err := Validate([]byte(`{"tokens":`))
	if !errors.Is(err, domain.ErrInvalidAuth) {
		t.Fatalf("expected ErrInvalidAuth, got %v", err)
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	err := Validate([]byte(`{"email":"user@example.com"}`))
	if !errors.Is(err, domain.ErrInvalidAuth) {
		t.Fatalf("expected ErrInvalidAuth, got %v", err)
	}
}

func TestNormalizeRemovesOnlyTopLevelLastRefresh(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"tokens":{"access":"abc","last_refresh":"keep-me"},
		"nested":{"last_refresh":"keep-me-too"},
		"last_refresh":"remove-me"
	}`)

	got, err := Normalize(raw)
	if err != nil {
		t.Fatalf("normalize auth: %v", err)
	}

	want := `{"nested":{"last_refresh":"keep-me-too"},"tokens":{"access":"abc","last_refresh":"keep-me"}}`
	if string(got) != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestHashIsStableAcrossKeyOrderAndLastRefresh(t *testing.T) {
	t.Parallel()

	left := []byte(`{"tokens":{"refresh":"r","access":"a"},"last_refresh":"2026-04-07T10:00:00Z","email":"user@example.com"}`)
	right := []byte(`{"email":"user@example.com","tokens":{"access":"a","refresh":"r"},"last_refresh":"2026-04-07T10:05:00Z"}`)

	leftHash, err := Hash(left)
	if err != nil {
		t.Fatalf("hash left auth: %v", err)
	}

	rightHash, err := Hash(right)
	if err != nil {
		t.Fatalf("hash right auth: %v", err)
	}

	if leftHash != rightHash {
		t.Fatalf("expected equal hashes, got %s and %s", leftHash, rightHash)
	}
}

func TestHashChangesWhenAuthChangesOutsideLastRefresh(t *testing.T) {
	t.Parallel()

	left := []byte(`{"tokens":{"access":"a"}}`)
	right := []byte(`{"tokens":{"access":"b"}}`)

	leftHash, err := Hash(left)
	if err != nil {
		t.Fatalf("hash left auth: %v", err)
	}

	rightHash, err := Hash(right)
	if err != nil {
		t.Fatalf("hash right auth: %v", err)
	}

	if leftHash == rightHash {
		t.Fatalf("expected different hashes, got %s", leftHash)
	}
}
