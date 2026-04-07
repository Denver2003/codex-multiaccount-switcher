# MVP Specification: Codex Account Switcher for Linux and macOS

## 1. Purpose

This document defines the MVP scope and product behavior for a local utility that allows a user to manually switch between multiple Codex accounts on Linux and macOS.

The primary business goal is to reduce the time needed to switch accounts when the current account reaches usage limits. The first version is focused on a reliable local CLI utility. A graphical UI is explicitly out of scope for this MVP.

## 2. Product Vision

The product manages multiple locally stored Codex authentication profiles and allows the user to:

- save the current authenticated Codex account as a reusable profile
- connect a new account through the standard Codex login flow
- switch the active Codex account by replacing the local auth state
- print clear restart guidance when a switch requires it

The utility does not replace Codex login. It orchestrates local state and uses Codex's existing authentication mechanism.

## 3. MVP Scope

### 3.1 In Scope

- Linux and macOS support
- CLI application
- local profile storage
- detection of the current Codex auth state
- save current account into profile storage
- add a new account through a guided flow
- switch between saved accounts
- remove and rename saved profiles
- basic status inspection
- safe backup and atomic replacement of active auth files
- restart recommendation messaging for Codex CLI and Codex desktop app on macOS

### 3.2 Out of Scope

- graphical UI
- automatic switching on limit detection
- usage/quota tracking
- cloud sync of accounts
- shared multi-user storage
- Windows support
- deep integration with internal Codex databases or session history
- recovery of partially broken desktop app state outside documented MVP behavior

## 4. Key Product Assumptions

### 4.1 Authentication Boundary

The MVP assumes that Codex CLI authentication is primarily represented by the local file:

- `~/.codex/auth.json`

The utility treats this file as the authoritative active auth artifact for CLI account switching.

### 4.2 Desktop App Boundary

The MVP assumes that the desktop app may keep additional state outside `~/.codex/auth.json`, especially on macOS. Because of that, desktop-app compatibility is best-effort in MVP and is limited to restart recommendation messaging.

The MVP will not directly modify macOS app storage under `~/Library/...` unless that becomes explicitly required in a later phase.

### 4.3 Session Behavior

The MVP assumes that already-running Codex processes may cache auth in memory. Therefore:

- switching account on disk does not guarantee hot reload inside already-running processes
- newly started Codex processes must use the newly active profile
- the tool should tell the user when a manual restart of Codex CLI or Codex desktop app is recommended

## 5. Users and Primary Use Cases

### 5.1 Target User

A technical user who:

- uses Codex CLI regularly
- has multiple accounts
- wants fast manual switching
- is comfortable with terminal workflows

### 5.2 Main Use Cases

#### Use Case A: Save the first account

The user is already logged into Codex. The tool detects active auth and offers to save it as the first profile.

#### Use Case B: Add a second or subsequent account

The user wants to connect a new Codex account without losing the current one. The tool saves the current account if needed, clears active auth, asks the user to log in with Codex, and then the user explicitly stores the newly authenticated state with `save-current`.

#### Use Case C: Switch account

The user selects a saved profile. The tool replaces the active Codex auth with the selected profile and tells the user whether manual restart of Codex CLI or the desktop app is recommended.

#### Use Case D: Inspect state

The user wants to see which profiles are stored and whether an active Codex auth currently exists.

#### Use Case E: Remove or rename profile

The user manages the local account list without changing active auth unless explicitly requested.

## 6. Product Principles

- Reliability over cleverness
- Explicit user actions over automation
- No destructive auth replacement without backup
- Do not depend on undocumented Codex internals beyond the active auth file in MVP
- Keep stored data portable and inspectable
- Keep secret handling conservative and simple
- Prefer explicit two-step flows over long interactive sessions when that reduces risk

## 7. Functional Requirements

### 7.1 Profile Management

The system must support creating, listing, renaming, removing, and reading account profiles.

Each profile must contain:

- stable internal profile ID
- user-visible label
- optional detected email, if it can be extracted safely
- creation timestamp
- last-used timestamp
- saved auth snapshot
- optional metadata for future extension

### 7.2 Active Auth Detection

The system must determine whether an active Codex authentication state exists by checking the active auth file path and validating its JSON structure.

The system must distinguish:

- no auth file
- unreadable auth file
- invalid auth file
- valid auth file

### 7.3 Save Current Account

The system must allow saving the current active auth state as a new reusable profile.

Behavior:

- read active auth
- validate JSON
- derive suggested label if possible
- allow user-provided label override
- compute `auth_hash` from normalized auth JSON
- if the same normalized auth is already stored, skip creating a duplicate and report the existing profile
- persist profile in profile storage
- do not modify the active auth file

### 7.4 Add New Account

The system must provide a guided flow for preparing the environment for connecting a new Codex account.

Behavior:

1. detect whether an active auth exists
2. if active auth exists, ask whether to save it before continuing
3. create safety backup of active auth if present
4. remove current active auth from the live location
5. instruct the user to authenticate manually in Codex CLI or Codex desktop app
6. instruct the user to verify the new login and then run `save-current`
7. complete successfully without waiting for a new auth file to appear

The tool must not attempt to fake or replace the Codex login flow itself.

### 7.5 Switch Profile

The system must allow switching the active Codex account to any saved profile.

Behavior:

1. verify selected profile exists
2. validate saved auth snapshot
3. back up current active auth if present
4. atomically replace active auth with selected profile snapshot
5. update `last_used_at`
6. print whether manual restart of Codex CLI or Codex desktop app is recommended

### 7.6 Status Inspection

The system must provide a command that reports:

- whether active auth exists
- path of active auth file
- whether the active auth matches a known saved profile, if determinable
- number of stored profiles
- last switch timestamp if known

### 7.7 Process Restart

The system must provide restart guidance after account switching.

Required MVP behavior:

- the tool does not attempt automatic process restart in MVP
- after a successful switch, the tool should state that already-running Codex processes may still use cached auth
- the tool should recommend restarting Codex CLI sessions and, on macOS, the Codex desktop app when relevant

Auth switching on disk must still complete successfully even when the tool can only provide a manual restart recommendation.

## 8. Command Surface for MVP

The exact naming can change at implementation time, but MVP must cover equivalent operations.

### 8.1 Required Commands

- `status`
- `list`
- `save-current`
- `add`
- `switch <profile>`
- `rename <profile> <new-label>`
- `remove <profile>`

### 8.2 Recommended Extra Commands

- `show <profile>`
- `backup list`
- `backup restore <id>`

These are not required for first delivery but strongly recommended if implementation cost is low.

## 9. Detailed User Flows

### 9.1 First Run Flow

Expected behavior on first run:

1. inspect active auth state
2. inspect profile storage
3. if no profiles exist and active auth is valid:
   - inform user that Codex is already authenticated
   - propose saving this state as the first profile
4. if no profiles exist and no active auth exists:
   - instruct user to authenticate with Codex first or run `add`

### 9.2 Add New Account Flow

Expected CLI interaction:

1. user runs `add`
2. tool checks current active auth
3. if present, tool asks whether to save current account first
4. tool backs up current auth
5. tool removes current active auth from the active location
6. tool asks user to authenticate via Codex CLI or Codex desktop app
7. tool tells the user to verify the login manually
8. tool tells the user to run `save-current` after the new account is authenticated
9. tool reports success for the preparation step

Expected follow-up:

1. user authenticates outside the tool
2. user runs `save-current`
3. tool validates current auth
4. tool asks for profile label, suggesting detected email if available
5. tool saves profile unless the same auth is already stored
6. tool reports whether it created a new profile or skipped because the auth already exists

### 9.3 Switch Flow

Expected CLI interaction:

1. user runs `switch work`
2. tool validates target profile
3. tool backs up current active auth
4. tool replaces active auth atomically
5. tool prints that restart of Codex CLI or Codex desktop app may be needed
6. tool does not attempt automatic restart in MVP

## 10. Storage Design

### 10.1 Active Auth Source

Default active auth path:

- `~/.codex/auth.json`

The implementation may allow overriding this path for testing, but the production default must target the standard Codex location.

### 10.2 Profile Store Location

Default profile storage location:

- macOS: `~/.config/codex-account-switcher/`
- Linux: `~/.config/codex-account-switcher/`

Alternative XDG-compliant handling is acceptable on Linux if it still resolves under the user's config directory.

### 10.3 Storage Layout

Recommended layout:

```text
~/.config/codex-account-switcher/
  metadata.json
  profiles/
    <profile-id>/
      auth.json
      profile.json
  backups/
    <timestamp>-<id>.auth.json
```

### 10.4 Metadata Model

Recommended global metadata fields:

- schema version
- created_at
- updated_at
- last_switch_at
- optional current_profile_id

Recommended per-profile metadata fields:

- id
- label
- email
- created_at
- last_used_at
- auth_hash

Profile labels must be unique in a case-insensitive comparison.

### 10.5 Validation

The system must validate:

- file existence
- JSON parseability
- basic expected auth structure before storing or activating

The MVP should not depend on a strict schema beyond basic sanity checks because the upstream auth format may evolve.

Email extraction is best-effort only. If no reliable email can be extracted, the tool must fall back to generated labels such as `account-1`.

For MVP, auth normalization rules are fixed as follows:

- parse the auth file as JSON
- remove only the top-level `last_refresh` field before hashing
- preserve all other fields exactly as parsed
- serialize the normalized JSON in canonical form before computing `auth_hash`

No additional fields may be ignored in MVP unless the specification is revised later.

## 11. Security Requirements

### 11.1 File Permissions

The system must create:

- storage directories with owner-only access where possible
- secret files with owner-only access where possible

On Unix-like systems, target permissions are:

- directories: `0700`
- auth and metadata files containing secrets: `0600`

### 11.2 Secret Handling

The MVP must not print auth tokens or secret values in stdout, logs, or error messages.

The MVP may store auth snapshots unencrypted on disk in the user's private config directory for initial delivery.

### 11.3 Encryption Strategy

Encryption at rest is not required in MVP.

Rationale:

- it adds UX complexity
- it adds key management requirements
- it is not necessary for the first proof of value

However, the design must leave room for a future encrypted storage mode.

### 11.4 Logging Rules

Logs must never contain:

- raw `auth.json` content
- API keys
- token payloads

Logs may contain:

- profile IDs
- labels
- file paths
- operation results
- validation failures with redacted context

## 12. Reliability Requirements

### 12.1 Atomic Write

Switch operations must use atomic replacement for the active auth file.

Recommended algorithm:

1. write new auth to temporary file in the same filesystem
2. fsync if supported and appropriate
3. rename temporary file over target path

### 12.2 Backup Before Replacement

Before replacing or removing active auth, the system must create a backup if an active auth file exists.

### 12.3 Recovery

If replacement fails:

- the previous auth must remain intact whenever possible
- the user must receive a clear recovery message
- the backup location must be reported

### 12.4 Corruption Handling

If a stored profile contains invalid auth data:

- the system must refuse activation
- the system must report which profile failed validation

## 13. Platform-Specific Behavior

### 13.1 Linux

Linux support in MVP is CLI-focused.

Required behavior:

- manage `~/.codex/auth.json`
- manage profile store under user config directory
- support standard shell usage

Optional behavior:

- print that manual restart of CLI sessions may be needed after switching

### 13.2 macOS

macOS support in MVP must include the same CLI behavior as Linux.

Optional macOS-only additions:

- detect whether app restart may be needed after switching

The MVP will not directly edit app-specific storage under `~/Library` unless that becomes a validated requirement later.

## 14. Non-Functional Requirements

### 14.1 Portability

The solution should be implemented as a single cross-platform CLI binary if practical.

### 14.2 Implementation Preference

Recommended language:

- Go

Rationale:

- simple static builds
- strong cross-platform file operations
- easy distribution
- good fit for CLI tooling

### 14.3 Performance

All core commands should complete in under one second on normal local hardware, excluding time spent waiting for direct user input inside the command itself.

### 14.4 UX

The CLI output must be concise and explicit.

Each destructive or state-changing operation should report:

- what was changed
- where backups were written
- whether restart is recommended

## 15. Error Handling Requirements

The system must provide actionable errors for:

- missing active auth
- unreadable active auth
- invalid JSON in active auth
- unknown profile
- duplicate profile label if labels must be unique
- failed backup
- failed atomic replacement
- attempt to save auth that is already stored

Error messages must suggest the next step when possible.

When the current auth is already stored, the message must identify the existing profile and explain that the save operation was skipped.

## 16. Suggested Label Resolution

When saving a profile, the system should derive a default label using the following priority:

1. explicit user input
2. detected email from auth data if safely available
3. fallback like `account-1`, `account-2`, and so on

Labels must be user-friendly and need not be globally immutable. Internal IDs must remain stable regardless of label changes.

## 17. Matching Active Auth to Saved Profiles

The MVP should attempt to identify whether the current active auth matches a saved profile.

Recommended strategy:

- compute a stable hash of normalized auth content
- compare active auth hash with stored profile hashes

This feature is optional if the upstream auth file contains non-deterministic fields that make stable matching unreliable. In that case, the tool should state that active profile cannot be determined with confidence.

The same matching logic should be reused by `save-current` to detect that the current auth is already stored and skip duplicate profile creation by default.

## 18. Acceptance Criteria

The MVP is considered complete when all items below are true.

### 18.1 Core Profile Operations

- user can save current active Codex auth as a profile
- user can list saved profiles
- user can rename a profile
- user can remove a profile

### 18.2 Add Flow

- user can prepare the environment for a new Codex account without manually copying auth files
- current auth is backed up before replacement
- the tool clearly instructs the user to authenticate outside the tool and then run `save-current`
- newly authenticated state can be saved as a named profile
- if the same auth was already saved earlier, `save-current` skips duplicate profile creation and reports the existing profile

### 18.3 Switch Flow

- user can activate any saved profile
- active `~/.codex/auth.json` is replaced atomically
- existing auth is backed up before replacement
- operation reports success or actionable failure

### 18.4 Safety

- secrets are not printed
- stored secret files are created with restrictive permissions
- invalid stored profiles are rejected during switch

### 18.5 Platform Support

- verified on macOS
- verified on Ubuntu

## 19. Test Scenarios for Later Development Phase

The future technical specification and implementation plan should include automated and manual tests for the following cases.

### 19.1 Happy Path

- save first account
- add second account
- switch between two valid profiles
- rename and remove inactive profile

### 19.2 Edge Cases

- switching when no active auth exists
- adding account when current auth is invalid
- saving profile with duplicate label
- saving profile when the same auth is already stored
- interrupted switch after backup but before replace
- corrupted stored profile auth
- permission denied in profile storage
- user stops after `add` and never completes a new login

### 19.3 Platform Checks

- Linux file permissions applied correctly
- macOS file permissions applied correctly
- restart recommendation messaging is explicit and predictable

## 20. Resolved MVP Decisions for Development Spec

These decisions are fixed for the MVP and should be treated as implementation requirements in the next document.

### 20.1 Add Command Strategy

The `add` command is a preparation step only.

It must:

- optionally save the currently active auth first
- always create a backup before removing active auth
- remove the active auth from the live location after backup
- instruct the user to authenticate manually in Codex CLI or Codex desktop app
- instruct the user to run `save-current` after login succeeds

It must not:

- automatically run `codex login`
- poll for a new auth file
- keep a long-running interactive session open waiting for login completion
- use a quarantine location for the active auth file in MVP

### 20.2 Restart Strategy

Automatic process restart is deferred out of MVP.

MVP behavior:

- `switch` completes only the auth replacement on disk
- command output must explain that existing Codex processes may still use cached auth
- command output must recommend manual restart of Codex CLI sessions and, on macOS, the Codex desktop app when relevant

### 20.3 Backup Retention

Backups are never pruned automatically in MVP.

Implications:

- every destructive auth change creates a backup when an active auth file exists
- cleanup, retention, or pruning commands may be added later

### 20.4 Email Extraction

Email extraction is best-effort only.

Rules:

- use detected email only when it can be extracted safely and confidently
- if email is unavailable, fall back to generated labels such as `account-1`
- email absence is not an error

### 20.5 Label Uniqueness Policy

Profile labels must be unique in a case-insensitive comparison.

Implications:

- `Work`, `work`, and `WORK` conflict
- when a requested label already exists, the tool must return an actionable error and suggest choosing another label
- profile IDs remain the stable internal identifier even though labels are unique

### 20.6 Existing Auth Deduplication

When `save-current` sees that the current normalized auth matches an already stored profile by `auth_hash`, the default behavior is to skip profile creation.

Implications:

- the tool must report which profile already contains that auth
- it must not create duplicate profiles for the same auth in MVP
- it must not silently overwrite the existing profile
- `auth_hash` must be computed from normalized JSON with only the top-level `last_refresh` field removed before canonical serialization

## 21. Recommended Next Step

The next document should be a development technical specification that converts this MVP definition into:

- concrete CLI command contract
- precise file schemas
- package/module structure
- implementation sequence
- test plan
- release checklist
