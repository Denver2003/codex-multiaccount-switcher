# codex-switcher

Local CLI utility for saving, switching, and managing multiple Codex auth profiles on Linux and macOS.

## Quick Start

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
```

Run the CLI directly:

```bash
./.bin/codex-switcher status
./.bin/codex-switcher list
./.bin/codex-switcher save-current
./.bin/codex-switcher switch work
```

Install into `~/.local/bin`:

```bash
make install
```

## Notes

- Active auth defaults to `~/.codex/auth.json`
- Switcher storage defaults to `~/.config/codex-account-switcher`
- `make smoke` uses a temporary local workspace under `.tmp/smoke`
