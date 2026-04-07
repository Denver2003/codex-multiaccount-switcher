# Task 08 - Distribution Without Go

## Goal

Make the CLI installable without requiring Go by shipping release binaries, then adding a one-command installer, then exposing a Homebrew tap.

## Scope

This task is intentionally sequential and should be implemented in three stages:

1. GitHub Releases with attached binaries
2. `install.sh` that downloads the correct release artifact
3. Homebrew tap support

Do not start a later stage before the previous stage is working and documented.

## Checklist

- [ ] Add a GitHub Actions release workflow that builds `codex-switcher` for at least:
- [ ] `linux-amd64`
- [ ] `linux-arm64`
- [ ] `darwin-amd64`
- [ ] `darwin-arm64`
- [ ] Package release artifacts in a consistent archive format and naming scheme.
- [ ] Ensure tagged releases publish downloadable artifacts to GitHub Releases.
- [ ] Document how maintainers cut a release and push tags.
- [ ] Update `README.md` with installation from release artifacts.

- [ ] Add `install.sh` at the repo root.
- [ ] Detect supported OS and architecture in the installer.
- [ ] Download the matching artifact from the latest GitHub Release.
- [ ] Verify the required tools exist (`curl` or `wget`, archive extraction, writable install path).
- [ ] Install the binary to a sensible default such as `~/.local/bin`.
- [ ] Print clear post-install guidance if the install directory is not in `PATH`.
- [ ] Update `README.md` with the one-line installer command.

- [ ] Prepare Homebrew distribution as the final stage.
- [ ] Choose whether to use a dedicated tap or submit to a broader formula repository.
- [ ] Add maintainer documentation for generating the Homebrew formula metadata.
- [ ] Ensure the release process includes checksums needed by Homebrew.
- [ ] Document the final Homebrew install command in `README.md`.

## Done When

- [ ] A user on supported macOS or Linux systems can install `codex-switcher` without Go.
- [ ] The release flow is documented well enough for another maintainer to run it.
- [ ] The repository documents all three install paths in the order:
- [ ] direct release artifact
- [ ] installer script
- [ ] Homebrew
- [ ] Each stage is shipped only after the previous one has been verified manually.
