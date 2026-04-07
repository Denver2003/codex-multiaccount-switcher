package store

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/fsx"
)

const (
	dirPerm    = 0o700
	filePerm   = 0o600
	schemaV1   = 1
	profPrefix = "prof_"
)

var profileIDEncoding = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

type Clock func() time.Time

type Store struct {
	paths *config.Resolver
	clock Clock
}

type ProfileRecord struct {
	Summary domain.ProfileSummary `json:",inline"`
}

func New(paths *config.Resolver) *Store {
	return &Store{
		paths: paths,
		clock: time.Now().UTC,
	}
}

func (s *Store) WithClock(clock Clock) *Store {
	copy := *s
	copy.clock = clock
	return &copy
}

func (s *Store) EnsureLayout() error {
	configDir, err := s.paths.ConfigDir()
	if err != nil {
		return err
	}

	profilesDir, err := s.paths.ProfilesDir()
	if err != nil {
		return err
	}

	backupsDir, err := s.paths.BackupsDir()
	if err != nil {
		return err
	}

	for _, path := range []string{configDir, profilesDir, backupsDir} {
		if err := fsx.EnsureDir(path, dirPerm); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ReadMetadata() (domain.Metadata, error) {
	path, err := s.paths.MetadataFile()
	if err != nil {
		return domain.Metadata{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Metadata{}, fmt.Errorf("%w: %s", domain.ErrMetadataNotFound, path)
		}

		return domain.Metadata{}, fmt.Errorf("read metadata %s: %w", path, err)
	}

	var metadata domain.Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return domain.Metadata{}, fmt.Errorf("%w: %s: %v", domain.ErrInvalidMetadata, path, err)
	}

	if metadata.SchemaVersion == 0 {
		metadata.SchemaVersion = schemaV1
	}

	if metadata.Profiles == nil {
		metadata.Profiles = []domain.ProfileSummary{}
	}

	return metadata, nil
}

func (s *Store) WriteMetadata(metadata domain.Metadata) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}

	if metadata.SchemaVersion == 0 {
		metadata.SchemaVersion = schemaV1
	}

	if metadata.Profiles == nil {
		metadata.Profiles = []domain.ProfileSummary{}
	}

	path, err := s.paths.MetadataFile()
	if err != nil {
		return err
	}

	return fsx.AtomicWriteJSON(path, metadata, filePerm)
}

func (s *Store) WriteProfile(record ProfileRecord, authData []byte) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}

	dir, err := s.profileDir(record.Summary.ID)
	if err != nil {
		return err
	}

	if err := fsx.EnsureDir(dir, dirPerm); err != nil {
		return err
	}

	if err := fsx.AtomicWriteJSON(filepath.Join(dir, "profile.json"), record.Summary, filePerm); err != nil {
		return err
	}

	if err := fsx.AtomicWriteFile(filepath.Join(dir, "auth.json"), authData, filePerm); err != nil {
		return err
	}

	return nil
}

func (s *Store) ReadProfile(id string) (ProfileRecord, []byte, error) {
	dir, err := s.profileDir(id)
	if err != nil {
		return ProfileRecord{}, nil, err
	}

	profileData, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProfileRecord{}, nil, fmt.Errorf("%w: %s", domain.ErrProfileNotFound, id)
		}

		return ProfileRecord{}, nil, fmt.Errorf("read profile %s: %w", id, err)
	}

	var summary domain.ProfileSummary
	if err := json.Unmarshal(profileData, &summary); err != nil {
		return ProfileRecord{}, nil, fmt.Errorf("parse profile %s: %w", id, err)
	}

	authData, err := os.ReadFile(filepath.Join(dir, "auth.json"))
	if err != nil {
		return ProfileRecord{}, nil, fmt.Errorf("read profile auth %s: %w", id, err)
	}

	return ProfileRecord{Summary: summary}, authData, nil
}

func (s *Store) VisibleProfiles() ([]domain.ProfileSummary, error) {
	metadata, err := s.ReadMetadata()
	if err != nil {
		if errors.Is(err, domain.ErrMetadataNotFound) {
			profilesDir, dirErr := s.paths.ProfilesDir()
			if dirErr != nil {
				return nil, dirErr
			}

			entries, readErr := os.ReadDir(profilesDir)
			if readErr == nil && len(entries) > 0 {
				return nil, fmt.Errorf("%w: metadata is missing while profile directories exist", domain.ErrMetadataNotFound)
			}

			return []domain.ProfileSummary{}, nil
		}

		return nil, err
	}

	profiles := make([]domain.ProfileSummary, len(metadata.Profiles))
	copy(profiles, metadata.Profiles)
	return profiles, nil
}

func (s *Store) CreateBackup(authData []byte) (string, string, error) {
	if err := s.EnsureLayout(); err != nil {
		return "", "", err
	}

	name, err := newBackupFilename(s.clock())
	if err != nil {
		return "", "", err
	}

	backupsDir, err := s.paths.BackupsDir()
	if err != nil {
		return "", "", err
	}

	path := filepath.Join(backupsDir, name)
	if err := fsx.AtomicWriteFile(path, authData, filePerm); err != nil {
		return "", "", err
	}

	return strings.TrimSuffix(name, ".auth.json"), path, nil
}

func NewProfileID() (string, error) {
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate profile ID: %w", err)
	}

	return profPrefix + strings.ToLower(profileIDEncoding.EncodeToString(randomBytes)), nil
}

func newBackupFilename(now time.Time) (string, error) {
	randomBytes := make([]byte, 3)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate backup suffix: %w", err)
	}

	return fmt.Sprintf("bkp_%s_%s.auth.json", now.UTC().Format("20060102T150405Z"), hex.EncodeToString(randomBytes)), nil
}

func (s *Store) profileDir(id string) (string, error) {
	profilesDir, err := s.paths.ProfilesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(profilesDir, id), nil
}

func (s *Store) ProfileDir(id string) (string, error) {
	return s.profileDir(id)
}
