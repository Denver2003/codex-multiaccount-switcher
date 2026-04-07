# Task 06 - Add, Rename, and Remove

## Goal

Implement the remaining management commands and their interactive behavior.

## Checklist

- [ ] Implement `add`.
- [ ] Require confirmation when needed.
- [ ] Back up live auth before deleting it.
- [ ] Remove the live auth file from the active location.
- [ ] Print manual login instructions for the next step.
- [ ] Implement `rename <profile> <new-label>`.
- [ ] Implement `remove <profile>`.
- [ ] Handle `--yes` and `--no-input` behavior.
- [ ] Treat EOF as refusal during confirmation.
- [ ] Add tests for confirmation behavior, label conflicts, and profile deletion.

## Done When

- [ ] `add` prepares the environment for a new login without polling.
- [ ] `rename` and `remove` update storage cleanly.
- [ ] Confirmation behavior matches the developer rules and spec.

