# Task 02 - Auth Reader, Validator, Normalizer, Hashing

## Goal

Implement Codex auth reading and the normalized hash logic used by `save-current`, `status`, and `switch`.

## Checklist

- [ ] Implement live auth file reading from the configured auth path.
- [ ] Implement minimum auth validation for active auth.
- [ ] Implement auth normalization rules from the development spec.
- [ ] Remove only the top-level `last_refresh` field during normalization.
- [ ] Serialize normalized JSON canonically before hashing.
- [ ] Compute `sha256:<hex>` auth hashes.
- [ ] Add tests for valid auth fixtures.
- [ ] Add tests for invalid JSON.
- [ ] Add tests for normalization stability.
- [ ] Add tests for dedup matching behavior.

## Done When

- [ ] Hashing is deterministic for the same normalized auth.
- [ ] Invalid auth is rejected with actionable errors.
- [ ] Tests cover the canonicalization and normalization rules.

