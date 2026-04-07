package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfigDirName = "codex-account-switcher"
	DefaultAuthDirName   = ".codex"
	DefaultAuthFileName  = "auth.json"
)

type Resolver struct {
	configDirOverride string
	authFileOverride  string
}

func NewResolver(configDirOverride, authFileOverride string) *Resolver {
	return &Resolver{
		configDirOverride: configDirOverride,
		authFileOverride:  authFileOverride,
	}
}

func (r *Resolver) ConfigDir() (string, error) {
	if r.configDirOverride != "" {
		return r.configDirOverride, nil
	}

	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(baseDir, DefaultConfigDirName), nil
}

func (r *Resolver) AuthFile() (string, error) {
	if r.authFileOverride != "" {
		return r.authFileOverride, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(homeDir, DefaultAuthDirName, DefaultAuthFileName), nil
}

func (r *Resolver) ProfilesDir() (string, error) {
	configDir, err := r.ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "profiles"), nil
}

func (r *Resolver) BackupsDir() (string, error) {
	configDir, err := r.ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "backups"), nil
}

func (r *Resolver) MetadataFile() (string, error) {
	configDir, err := r.ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "metadata.json"), nil
}
