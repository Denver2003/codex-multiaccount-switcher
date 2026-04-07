# Task 04 - Status and List Commands

## Goal

Implement read-only inspection commands first.

## Checklist

- [ ] Implement `status`.
- [ ] Implement `list`.
- [ ] Resolve the active profile by matching normalized auth hashes.
- [ ] Report active auth presence and validity.
- [ ] Report the profile store path.
- [ ] Report stored profile count.
- [ ] Report the current profile label and ID when determinable.
- [ ] Sort `list` output by label, then profile ID.
- [ ] Add tests for empty storage and populated storage.

## Done When

- [ ] Both commands work without mutating state.
- [ ] Output is concise and consistent.
- [ ] Status gracefully reports invalid active auth.

