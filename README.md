# codex-switcher

Local CLI utility for saving, switching, and managing multiple Codex auth profiles on Linux and macOS.

## Install

Install paths are documented in this order:

1. direct release artifact
2. installer script
3. Homebrew

### 1. Direct Release Artifact

Download the archive for your platform from GitHub Releases, extract it, and move `codex-switcher` into a directory in your `PATH`.

Release archives follow this naming scheme:

```text
codex-switcher_<version>_<os>_<arch>.tar.gz
```

### 2. Installer Script

Install the latest release into `~/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/Denver2003/codex-multiaccount-switcher/main/install.sh | sh
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/Denver2003/codex-multiaccount-switcher/main/install.sh | VERSION=v1.0.0 sh
```

### 3. Homebrew

Dedicated tap:

```bash
brew tap Denver2003/codex-switcher
brew install codex-switcher
```

## Development

Requirements:

- Go in `PATH`
- `golangci-lint` in `PATH` for linting

Common commands:

```bash
make build
make test
make vet
make lint
make smoke
make install
```

## Notes

- Active auth defaults to `~/.codex/auth.json`
- Switcher storage defaults to `~/.config/codex-account-switcher`
- `make smoke` uses a temporary local workspace under `.tmp/smoke`
- Maintainer release flow is documented in [docs/distribution.md](./docs/distribution.md)
