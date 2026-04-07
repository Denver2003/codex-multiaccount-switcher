# Task 05 - Save Current and Switch

## Goal

Implement the two core state-changing account operations.

## Checklist

- [ ] Implement `save-current`.
- [ ] Detect and skip duplicate auth via normalized hash.
- [ ] Resolve labels using the spec rules.
- [ ] Enforce case-insensitive label uniqueness.
- [ ] Implement `switch <profile>`.
- [ ] Validate stored profile auth before activation.
- [ ] Backup the live auth file before replacement.
- [ ] Replace the live auth atomically.
- [ ] Update profile and metadata timestamps after switch.
- [ ] Print restart guidance after switch.
- [ ] Add tests for duplicate detection, successful switch, and corrupted profile handling.

## Done When

- [ ] `save-current` is safe to run repeatedly.
- [ ] `switch` activates the requested profile reliably.
- [ ] Both commands preserve backups and respect commit semantics.

