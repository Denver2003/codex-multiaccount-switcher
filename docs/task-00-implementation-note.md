# Task 00 Implementation Note

## Confirmed Scope

- MVP is CLI-only.
- Supported platforms are Linux and macOS.
- The active Codex auth file is `~/.codex/auth.json`.
- The default switcher config directory is `~/.config/codex-account-switcher`.
- Required MVP commands are `status`, `list`, `save-current`, `add`, `switch`, `rename`, and `remove`.
- `metadata.json` is the authoritative profile index and the commit point for profile visibility.
- Profile-local files are written first, and `metadata.json` is written last.

## Remaining Ambiguities

No blocking spec ambiguity was found for Task 00.

Implementation should continue strictly in task order and avoid adding optional commands before the required MVP command set is complete.
