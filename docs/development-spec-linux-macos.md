# Development Specification: Codex Account Switcher MVP for Linux and macOS

## 1. Document Purpose

This document translates the approved MVP product specification into an implementation-ready technical specification.

Its purpose is to define:

- the CLI contract
- the storage schema
- the internal architecture
- the algorithms for account save/add/switch flows
- validation and error rules
- the implementation sequence
- the testing scope

This document is intended to be the direct source for breaking the work into engineering tasks.

## 2. Scope

This specification covers only the MVP for:

- Linux
- macOS

It covers only the CLI product. UI work is explicitly out of scope.

## 3. Implementation Goals

The implementation must optimize for:

- predictable behavior
- local safety
- minimal coupling to undocumented Codex internals
- portability across Linux and macOS
- low operational complexity

The implementation must not optimize for:

- background automation
- silent magic behavior
- long-lived interactive sessions

## 4. Technology Decision

### 4.1 Primary Language

The project should be implemented in Go.

### 4.2 Go Version

Recommended baseline:

- Go 1.23 or newer

Exact version can be finalized during project bootstrap, but the implementation should target a recent stable Go release with standard library support for:

- JSON handling
- file permissions
- atomic rename
- path operations
- CLI execution

### 4.3 External Dependencies

MVP should prefer the Go standard library wherever practical.

Allowed categories of dependencies:

- CLI framework, if it materially improves command structure
- canonical JSON serializer, if standard library output is not sufficient for stable hashing

Preferred dependency posture:

- zero or minimal runtime dependencies
- no crypto dependency beyond standard hash functions needed for dedup
- no OS credential store integration in MVP

## 5. Product Binary

### 5.1 Binary Name

Working binary name for development:

- `codex-switcher`

This can be changed later if naming requirements evolve.

### 5.2 Invocation Style

Expected invocation style:

```bash
codex-switcher <command> [flags]
```

## 6. CLI Contract

### 6.1 Global Behavior

All commands must:

- print concise human-readable output
- avoid printing secrets
- return non-zero exit codes on failure

For MVP, human-readable output is required.

Structured `--json` output is explicitly out of scope for MVP and may be added later.

### 6.2 Global Flags

Recommended global flags:

- `--verbose`
- `--config-dir <path>`
- `--auth-file <path>`

Rules:

- `--config-dir` is primarily for tests and advanced users
- `--auth-file` is primarily for tests and advanced users
- default behavior must not require flags

### 6.3 Required Commands

Required commands:

- `status`
- `list`
- `save-current`
- `add`
- `switch <profile>`
- `rename <profile> <new-label>`
- `remove <profile>`

### 6.4 Optional Commands

Recommended if cost is low:

- `show <profile>`
- `backup list`
- `backup restore <backup-id>`

These should not block MVP completion.

## 7. Command Specifications

### 7.1 `status`

Purpose:

- inspect active auth state
- inspect whether it matches a saved profile
- summarize local switcher state

Inputs:

- no positional arguments

Success output must include:

- active auth presence
- active auth file path
- profile store path
- saved profile count
- whether active auth matches a known profile
- current profile label and ID if match is known
- last switch timestamp if available

Failure conditions:

- config directory unreadable
- metadata unreadable and unrecoverable

Behavior notes:

- missing metadata file is not an error if profile storage has not been initialized yet
- invalid active auth should be reported as invalid state, not necessarily as command failure

### 7.2 `list`

Purpose:

- print stored profiles

Inputs:

- no positional arguments

Success output per profile should include:

- label
- ID
- email if available
- created_at
- last_used_at

Sorting:

- default sort by label ascending
- ties are resolved by profile ID ascending

Failure conditions:

- unreadable profile store
- unreadable metadata that prevents listing

### 7.3 `save-current`

Purpose:

- save the currently active auth as a reusable profile

Inputs:

- optional `--label <value>`

Preconditions:

- active auth file exists
- active auth file is readable
- active auth file is valid enough to pass minimal validation

Success behavior:

1. read active auth
2. normalize auth JSON
3. compute `auth_hash`
4. check whether the same `auth_hash` already exists
5. if duplicate exists:
   - do not create a new profile
   - report existing profile label and ID
   - return success
6. resolve label
7. validate label uniqueness
8. create profile directory
9. write `auth.json`
10. write `profile.json`
11. update global metadata if needed
12. report created profile

Failure conditions:

- no active auth file
- unreadable active auth file
- invalid auth JSON
- label conflict
- profile store write failure

Commit rule:

- a newly created profile is considered committed only after `metadata.json` has been written successfully
- if profile-local files were written but `metadata.json` update failed, the command must report failure and the profile must not be treated as visible by normal read commands

### 7.4 `add`

Purpose:

- prepare the environment for logging into a different Codex account

Inputs:

- optional `--save-current`
- optional `--label <value>` only if the implementation chooses to immediately pass a suggested label into the final messaging

Preconditions:

- none

Success behavior:

1. detect active auth state
2. if active auth exists:
   - if `--save-current` is passed, run the same internal flow as `save-current`
   - otherwise ask for confirmation in interactive mode or fail with an explicit instruction in non-interactive mode
3. create backup of active auth if it exists
4. delete live active auth file
5. print exact next steps:
   - authenticate manually in Codex CLI or Codex desktop app
   - verify login completed successfully
   - run `codex-switcher save-current`
6. return success

Failure conditions:

- active auth exists and user refuses confirmation
- backup creation fails
- active auth deletion fails

Non-goals:

- must not launch `codex login`
- must not poll for new auth
- must not keep the command running waiting for login
- must not use quarantine storage

### 7.5 `switch <profile>`

Purpose:

- activate a saved profile as the live Codex auth

Inputs:

- one required profile selector

Selector behavior:

- first try exact label match, case-insensitive
- if not found, try exact ID match
- partial matches are not allowed in MVP

Preconditions:

- target profile exists
- stored profile auth is readable and valid

Success behavior:

1. resolve target profile
2. read and validate target auth
3. back up current active auth if present
4. atomically write target auth to active auth path
5. update target profile `last_used_at`
6. update global metadata:
   - `updated_at`
   - `last_switch_at`
   - `current_profile_id`
7. print restart guidance
8. return success

Failure conditions:

- unknown profile
- invalid stored auth
- backup failure
- atomic replacement failure
- metadata update failure after switch

Important rule:

- the active auth replacement is the primary operation
- if metadata update fails after a successful auth replacement, output must clearly state that switch succeeded but metadata update failed

### 7.6 `rename <profile> <new-label>`

Purpose:

- rename a saved profile label

Success behavior:

1. resolve target profile
2. validate `new-label`
3. ensure no case-insensitive label conflict
4. update profile metadata
5. update global metadata timestamp

Failure conditions:

- unknown profile
- invalid new label
- duplicate label
- write failure

### 7.7 `remove <profile>`

Purpose:

- delete a saved profile from the switcher store

Success behavior:

1. resolve target profile
2. ask for confirmation in interactive mode, unless `--yes`
3. remove profile directory
4. update global metadata
5. if removed profile was `current_profile_id`, clear that field

Failure conditions:

- unknown profile
- deletion failure
- metadata update failure

Behavior note:

- removing a saved profile does not modify the live active auth file

## 7.8 Interactive Behavior

Interactive mode must be determined as:

- `stdin` is a TTY
- `stdout` is a TTY
- `--no-input` is not set

Non-interactive mode applies when any of the above conditions is false.

Confirmation parsing rules:

- accept `y` and `yes` in any letter case as confirmation
- any other non-empty input is treated as refusal
- EOF is treated as refusal

Prompt text does not need to be byte-for-byte fixed in MVP, but must clearly state:

- what action is about to happen
- whether auth will be removed or a profile deleted
- how the user can avoid the action

## 8. Output and Exit Codes

### 8.1 Human Output

Default output should be concise.

Example style:

```text
Saved profile: work
ID: prof_01J...
Source: /Users/name/.codex/auth.json
```

### 8.2 Exit Code Convention

Recommended exit codes:

- `0`: success
- `1`: generic runtime failure
- `2`: invalid arguments or usage
- `3`: state/precondition error
- `4`: validation error
- `5`: storage/write error

Exact values may change, but categories should remain distinct internally.

## 9. Storage Layout

### 9.1 Default Paths

Default active auth path:

- `~/.codex/auth.json`

Default config directory:

- `~/.config/codex-account-switcher`

### 9.2 File Layout

Required layout:

```text
~/.config/codex-account-switcher/
  metadata.json
  profiles/
    <profile-id>/
      auth.json
      profile.json
  backups/
    <backup-id>.auth.json
```

### 9.3 Directory Initialization

On first write operation, the tool must ensure:

- config directory exists with `0700`
- `profiles` exists with `0700`
- `backups` exists with `0700`
- `metadata.json`, when created, uses `0600`

## 10. File Schemas

### 10.1 `metadata.json`

Required fields:

```json
{
  "schema_version": 1,
  "created_at": "2026-04-07T10:00:00Z",
  "updated_at": "2026-04-07T10:00:00Z",
  "last_switch_at": "2026-04-07T10:00:00Z",
  "current_profile_id": "prof_01ABCDEF",
  "profiles": [
    {
      "id": "prof_01ABCDEF",
      "label": "work",
      "email": "name@example.com",
      "created_at": "2026-04-07T10:00:00Z",
      "last_used_at": "2026-04-07T10:00:00Z",
      "auth_hash": "sha256:..."
    }
  ]
}
```

Notes:

- `current_profile_id` may be empty or omitted if unknown
- `last_switch_at` may be empty or omitted before the first switch
- `profiles` is the authoritative profile index
- `metadata.json` must be written with file mode `0600`

### 10.2 `profile.json`

Recommended per-profile file:

```json
{
  "id": "prof_01ABCDEF",
  "label": "work",
  "email": "name@example.com",
  "created_at": "2026-04-07T10:00:00Z",
  "last_used_at": "2026-04-07T10:00:00Z",
  "auth_hash": "sha256:..."
}
```

Rule:

- `profile.json` and the metadata entry for that profile must contain the same values
- `profile.json` must be written with file mode `0600`

### 10.3 `auth.json`

This file stores the exact active auth snapshot as captured from the live Codex auth file.

Rules:

- content must be preserved exactly as read, except for trailing newline normalization if required by file writing implementation
- no secret fields may be removed from stored profile auth
- no migration of upstream auth format should happen in MVP
- stored profile `auth.json` must be written with file mode `0600`

### 10.4 Backup File Naming

Recommended backup ID format:

- `bkp_<UTC compact timestamp>_<short random suffix>`

Example:

- `bkp_20260407T101530Z_a1b2c3.auth.json`

Backup metadata may be encoded in filename only for MVP.

If implementation cost is low, a separate `backups.json` index may be added later, but it is not required.

Backup files must be written with file mode `0600`.

### 10.5 Profile ID Format

`profile-id` format is fixed for MVP.

Rules:

- prefix: `prof_`
- suffix: 16 lowercase base32 characters without padding
- suffix must be generated from cryptographically secure random bytes

Example:

- `prof_8f3k2m1q7t9v4x6z`

The implementation must treat profile IDs as opaque identifiers outside of validation and storage.

## 11. Auth Normalization and Hashing

### 11.1 Purpose

Normalization exists only for:

- deduplication in `save-current`
- matching active auth against saved profiles in `status`

Normalization does not change stored auth snapshots.

### 11.2 Rules

Normalization algorithm for MVP:

1. parse auth JSON into a generic JSON object
2. remove the top-level `last_refresh` field if present
3. preserve every other field
4. serialize into canonical JSON
5. compute SHA-256 hash
6. format hash as `sha256:<hex>`

### 11.3 Canonical JSON Requirement

The implementation must use deterministic serialization for hashing.

Acceptable approaches:

- recursively sort object keys before marshaling
- use a dependency that guarantees canonical JSON ordering

The chosen approach must be covered by tests.

### 11.4 Invalid JSON

If auth JSON cannot be parsed:

- `save-current` must fail
- `status` must report invalid auth state
- `switch` must refuse to activate that stored profile

## 12. Label Rules

### 12.1 Allowed Labels

MVP label rules:

- length: `1..64`
- trim leading and trailing whitespace
- must not be empty after trim
- must not contain path separators `/` or `\`
- must be valid UTF-8

ASCII-only labels are not required.

### 12.2 Uniqueness

Labels must be unique by case-insensitive comparison.

Examples of conflicts:

- `work` vs `Work`
- `Account A` vs `account a`

### 12.3 Default Label Resolution

Resolution order:

1. explicit `--label`
2. detected email
3. generated fallback `account-N`

### 12.4 Fallback Label Generation

Fallback labels must be generated deterministically from the current visible profile set.

Algorithm:

1. build the set of existing labels using case-insensitive comparison
2. start from `account-1`
3. increment `N` until a free label is found
4. use the first free label

Rules:

- deleted profiles do not reserve labels
- hidden or orphaned profile directories do not reserve labels unless they are present in `metadata.json`

### 12.5 Email Extraction

Email extraction is best-effort only.

MVP rule:

- parse auth JSON generically
- only use a value as email if there is an explicitly confirmed top-level string field named `email`
- if such a field is absent, do not infer email from any other field
- if such a field is present but does not look like a basic email address, ignore it

This intentionally keeps email extraction narrow in MVP.

## 13. Validation Rules

### 13.1 Active Auth Validation

Minimum validation for active auth:

- file exists
- readable
- valid JSON object at top level
- contains at least one of:
  - `tokens`
  - `OPENAI_API_KEY`

This is intentionally permissive to avoid coupling to internal format changes.

### 13.2 Stored Profile Validation

When switching to a stored profile, validate:

- `profile.json` readable
- `auth.json` readable
- `profile.json` contains required fields
- `auth.json` passes minimum auth validation
- `auth_hash` in `profile.json` matches recomputed hash from stored auth

Hash mismatch behavior:

- refuse switch
- report profile corruption or drift

### 13.3 Metadata Validation

If `metadata.json` is missing:

- initialize lazily when first write occurs

If `metadata.json` exists but is invalid:

- read-only commands should fail with actionable error
- write commands should fail unless safe recovery logic is explicitly implemented

## 14. File IO and Safety

### 14.1 Atomic Write Strategy

All managed JSON files must be written atomically.

Algorithm:

1. create temp file in the same directory
2. write contents
3. flush and close file
4. set required permissions
5. rename over destination

Recommended implementation detail:

- call file `fsync` before rename when supported in the chosen implementation path

Out of scope for MVP:

- directory `fsync` after rename

### 14.2 Active Auth Replacement

Switching active auth must use atomic replace at the final live path.

Recommended algorithm:

1. ensure parent directory exists
2. write target auth to temp file in `~/.codex`
3. chmod `0600`
4. rename temp file to `auth.json`

### 14.3 Backup Creation

Before deleting or replacing live auth:

1. check whether live auth exists
2. if yes, copy contents to backup file
3. ensure backup file permissions are `0600`

Backup creation must happen before destructive change.

### 14.4 Permission Repair

When reading existing files not created by the tool:

- do not fail only because permissions are broader than desired

When writing files managed by the tool:

- always enforce target permissions

## 15. Metadata Synchronization Rules

### 15.1 Source of Truth

For MVP:

- `metadata.json` is the primary index
- `profile.json` is duplicated profile-local metadata for portability and recovery

### 15.2 Sync Rule

Whenever a profile changes, the tool must update:

- the entry in `metadata.json`
- the corresponding `profile.json`

Write ordering rules:

- for profile creation, write profile-local files first and `metadata.json` last
- for profile update, write profile-local files first and `metadata.json` last
- for profile deletion, delete the profile directory first and write `metadata.json` last if the deletion succeeds

Commit semantics:

- `metadata.json` is the commit point for visibility in MVP
- normal read commands must trust only the profiles listed in `metadata.json`
- orphaned profile directories must be ignored by normal reads

### 15.3 Recovery Rule

If `metadata.json` is missing but profile directories exist, automatic rebuild is not required in MVP.

Recommended behavior:

- fail with explicit error
- provide a future task for repair tooling

This keeps MVP logic simpler and safer.

## 16. Interactivity Rules

### 16.1 Default Mode

Commands may be interactive in a terminal session where confirmation is required.

### 16.2 Non-Interactive Mode

Recommended flags:

- `--yes`
- `--no-input`

If `--no-input` is set and confirmation would be needed:

- fail with actionable error
- return a state/precondition error

### 16.3 Confirmation Points

Confirmation should be required for:

- `add` when active auth exists and `--save-current` is not provided
- `remove` unless `--yes`

`switch` should not require confirmation in MVP.

## 17. Restart Guidance Rules

### 17.1 Linux Messaging

After a successful switch on Linux, output should say that:

- existing Codex CLI sessions may still use cached auth
- new Codex CLI processes will use the switched account
- user should restart active Codex sessions if behavior does not reflect the new account

### 17.2 macOS Messaging

After a successful switch on macOS, output should say that:

- existing Codex CLI sessions may still use cached auth
- the Codex desktop app may also require manual restart
- new processes should use the switched account

### 17.3 Restart Detection

The tool does not need to detect running processes in MVP.

Static guidance is sufficient.

## 18. Suggested Project Structure

Recommended repository layout:

```text
.
  cmd/
    codex-switcher/
      main.go
  internal/
    app/
      app.go
    cli/
      root.go
      status.go
      list.go
      save_current.go
      add.go
      switch.go
      rename.go
      remove.go
    config/
      paths.go
    domain/
      model.go
      errors.go
    store/
      metadata_store.go
      profile_store.go
      backup_store.go
    auth/
      reader.go
      validator.go
      normalize.go
      hash.go
      live_auth.go
    ops/
      save_current.go
      add.go
      switch.go
      rename.go
      remove.go
      status.go
      list.go
    fsx/
      atomic_write.go
      permissions.go
      temp.go
    ux/
      output.go
      confirm.go
```

This layout is guidance, not a hard requirement, but the codebase should preserve separation between:

- CLI parsing
- domain model
- storage
- auth logic
- operations

## 19. Internal Module Responsibilities

### 19.1 `config`

Responsibilities:

- resolve default paths
- apply overrides from flags
- expand `~`

### 19.2 `domain`

Responsibilities:

- core structs
- typed errors
- constants

### 19.3 `auth`

Responsibilities:

- read active auth
- validate auth
- normalize auth for hashing
- compute hash

### 19.4 `store`

Responsibilities:

- load and save metadata
- create/read/update/delete profile files
- create backups

### 19.5 `ops`

Responsibilities:

- implement high-level use cases
- compose auth and store services
- contain business flow, not raw CLI parsing

### 19.6 `cli`

Responsibilities:

- parse command-line arguments
- call operation layer
- format terminal output

## 20. Core Data Models

### 20.1 `ProfileRecord`

Required fields:

- `ID string`
- `Label string`
- `Email string`
- `CreatedAt time.Time`
- `LastUsedAt *time.Time`
- `AuthHash string`

Constraints:

- `ID` must match the MVP `profile-id` format
- `Label` must follow label validation rules

### 20.2 `Metadata`

Required fields:

- `SchemaVersion int`
- `CreatedAt time.Time`
- `UpdatedAt time.Time`
- `LastSwitchAt *time.Time`
- `CurrentProfileID string`
- `Profiles []ProfileRecord`

### 20.3 `StatusResult`

Suggested fields:

- `ActiveAuthExists bool`
- `ActiveAuthValid bool`
- `ActiveAuthPath string`
- `ConfigDir string`
- `ProfileCount int`
- `MatchedProfileID string`
- `MatchedProfileLabel string`
- `LastSwitchAt *time.Time`

## 21. Error Model

### 21.1 Error Categories

Recommended typed categories:

- usage error
- validation error
- state error
- storage error
- corruption error

### 21.2 Named Conditions

Recommended internal errors:

- `ErrNoActiveAuth`
- `ErrInvalidActiveAuth`
- `ErrDuplicateLabel`
- `ErrProfileNotFound`
- `ErrAuthAlreadyStored`
- `ErrProfileCorrupted`
- `ErrMetadataCorrupted`
- `ErrConfirmationRequired`

These names are illustrative; exact identifiers may vary.

## 22. Algorithm Specs

### 22.1 Save Current Algorithm

Algorithm:

1. resolve live auth path
2. read live auth bytes
3. validate auth
4. normalize auth
5. compute hash
6. load metadata, or initialize empty metadata if absent
7. search metadata profiles for matching hash
8. if found, return success with existing profile reference
9. resolve label
10. validate label and uniqueness
11. create new profile ID
12. create profile directory
13. write stored `auth.json`
14. write `profile.json`
15. append profile to metadata
16. update metadata timestamps
17. write `metadata.json`
18. return created profile

### 22.2 Add Algorithm

Algorithm:

1. detect live auth presence
2. if absent:
   - print manual login instructions
   - return success
3. if present and `--save-current`:
   - execute save-current subflow
4. else request confirmation or fail in non-interactive mode
5. create backup of live auth
6. delete live auth file
7. print exact next steps
8. return success

### 22.3 Switch Algorithm

Algorithm:

1. resolve target profile
2. read target `profile.json`
3. read target `auth.json`
4. validate target auth
5. normalize target auth and recompute hash
6. compare recomputed hash to stored `auth_hash`
7. fail if mismatch
8. create backup of live auth if it exists
9. atomically replace live auth with target auth bytes
10. update target `last_used_at`
11. update metadata timestamps and `current_profile_id`
12. persist metadata and `profile.json`
13. print restart guidance
14. return success

### 22.4 Rename Algorithm

Algorithm:

1. resolve profile
2. normalize new label
3. validate label
4. check uniqueness
5. update metadata entry
6. update `profile.json`
7. persist metadata

### 22.6 Confirmation Algorithm

Algorithm:

1. detect interactive mode
2. if interactive mode is false:
   - if command supports `--yes` and it is set, continue
   - otherwise fail with confirmation-required error
3. print confirmation prompt
4. read a single line from stdin
5. if EOF, treat as refusal
6. normalize input by trim + lowercase
7. continue only for `y` or `yes`
8. otherwise abort without side effects

### 22.5 Remove Algorithm

Algorithm:

1. resolve profile
2. confirm if needed
3. delete profile directory recursively
4. remove profile from metadata
5. clear `current_profile_id` if it matches removed ID
6. persist metadata

## 23. Testing Strategy

### 23.1 Unit Tests

Required unit test areas:

- auth validation
- auth normalization
- canonical hashing
- label validation
- profile ID generation and validation
- fallback label generation
- duplicate detection
- path resolution
- metadata read/write
- atomic write helper

### 23.2 Integration Tests

Required integration test areas:

- save-current creates profile and metadata
- save-current skips duplicate auth
- save-current leaves orphaned files ignored if metadata commit fails
- add creates backup and deletes live auth
- switch replaces live auth
- switch updates timestamps
- rename updates both metadata and profile file
- remove deletes profile and updates metadata
- non-interactive confirmation failure behavior

### 23.3 Fixture Strategy

Recommended test fixtures:

- valid auth with `OPENAI_API_KEY`
- valid auth with `tokens`
- auth including `last_refresh`
- invalid JSON auth
- auth missing expected top-level indicators

### 23.4 Platform Tests

Minimum manual platform verification:

- macOS: save/add/switch/list/status
- Ubuntu: save/add/switch/list/status

Permissions to verify:

- config dir `0700`
- secret files `0600`

## 24. Logging and Observability

### 24.1 Logging

MVP logging should be minimal.

Recommended:

- no persistent logs by default
- verbose diagnostics only under `--verbose`

### 24.2 Redaction

Any debug output must redact:

- API keys
- token values
- raw auth payload

### 24.3 Auditability

Operation success messages should include enough context for manual audit:

- affected profile
- target paths
- backup file path when created

## 25. Packaging and Build

### 25.1 Build Targets

Initial build targets:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64` if cost is low

### 25.2 Build Output

Single binary per target.

### 25.3 Distribution

MVP can be distributed as:

- local build artifact
- tarball or zip
- GitHub release artifact later

Installer packaging is not required.

## 26. Implementation Phases

### Phase 1: Bootstrap

- initialize Go module
- add CLI skeleton
- add path/config resolution
- add domain models

### Phase 2: Core Storage and Auth

- implement metadata store
- implement profile store
- implement backup store
- implement auth validation
- implement normalization and hashing

### Phase 3: Core Commands

- implement `status`
- implement `list`
- implement `save-current`
- implement `switch`

### Phase 4: Management Commands

- implement `add`
- implement `rename`
- implement `remove`

### Phase 5: Verification

- unit tests
- integration tests
- manual verification on macOS
- manual verification on Ubuntu

## 27. Suggested Task Breakdown

This section is intentionally phrased so it can later be converted directly into implementation tasks.

### Task Group A: Project Bootstrap

- create Go module and binary entrypoint
- add CLI framework and command registration
- define domain types and error types

### Task Group B: Path and Config Layer

- implement default path resolution
- implement support for `--config-dir`
- implement support for `--auth-file`

### Task Group C: Auth Layer

- implement live auth reader
- implement active auth validator
- implement auth normalization
- implement canonical auth hashing
- add tests for auth edge cases

### Task Group D: Storage Layer

- implement metadata load/save
- implement profile create/read/update/delete
- implement backup create/list
- implement atomic JSON write helper
- implement permission enforcement for created files

### Task Group E: Operation Layer

- implement `save-current` flow
- implement `switch` flow
- implement `add` flow
- implement `rename` flow
- implement `remove` flow
- implement `status` and `list`

### Task Group F: CLI UX

- human-readable output formatting
- confirmation prompts
- non-interactive behavior
- restart guidance messaging

### Task Group G: Quality

- unit test suite
- integration test suite with temp directories
- manual verification checklist for macOS and Ubuntu

## 28. Deferred Work

Explicitly deferred from this development spec:

- automatic process restart
- Windows support
- encrypted at-rest profile storage
- metadata rebuild tooling
- backup pruning
- desktop app storage integration
- auto-detection of account limits
- GUI

## 29. Definition of Done

The implementation is done when:

- all required commands exist
- command behavior matches this document
- save/add/switch flows work on macOS and Ubuntu
- duplicates are prevented using normalized auth hashing
- backups are created before destructive auth changes
- file writes are atomic where required
- managed files get restrictive permissions
- interactive and non-interactive confirmation behavior is predictable
- tests cover core logic and error paths
- manual verification has been completed on both supported platforms
