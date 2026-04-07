# Developer Agent Rules

This repository is being built from the approved MVP and development specifications in `docs/`.

## General Rules

- Keep changes aligned with `docs/mvp-spec-linux-macos.md` and `docs/development-spec-linux-macos.md`.
- Prefer small, reviewable changes.
- Do not introduce features that are not required by the current task.
- Do not silently change product decisions without updating the specs first.
- Keep the codebase clean and readable.
- Avoid unused variables, unused imports, dead code, and warning-producing patterns.

## Code Quality

- Follow idiomatic Go style.
- Keep functions small and focused.
- Use clear names for packages, types, and functions.
- Prefer simple implementations over clever ones.
- Add comments only when the code is not obvious.
- Preserve ASCII unless a file already uses non-ASCII text or non-ASCII is required by the task.

## Formatting and Checks

Before finishing work, the developer should run:

- `gofmt -w` on modified Go files
- `go test ./...`
- `go vet ./...`
- `golangci-lint run ./...` if `golangci-lint` is available in the environment

If `golangci-lint` is not available, the developer should still run the available checks and explicitly mention that the linter binary is missing.

## Lint Expectations

- No unused variables.
- No unused imports.
- No obvious nil dereference risks.
- No ignored errors unless they are intentionally and explicitly handled.
- No warning-level issues from the configured linters.
- No TODOs unless the task explicitly requires a follow-up marker.

## Implementation Rules

- Prefer standard library code first.
- Keep the CLI behavior explicit and predictable.
- Use atomic file writes for managed files.
- Do not weaken file permissions unless the spec explicitly allows it.
- Do not add automatic restart logic unless the task or spec explicitly requires it.
- Respect the storage layout and commit semantics defined in the development spec.

## Verification Rules

- Every meaningful change should have tests where practical.
- When adding or changing filesystem behavior, include tests for success and failure paths.
- When touching normalization, hashing, or deduplication logic, include fixtures for the edge cases.
- When adding CLI commands, verify argument parsing and exit behavior.

## Task Execution

- Work through the task list in `tasks/` in order unless a task explicitly says it can be done in parallel.
- Do not skip implementation steps that are listed as prerequisites.
- If a task exposes a spec ambiguity, stop and raise it before inventing behavior.

