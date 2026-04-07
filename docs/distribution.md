# Distribution Guide

## Release Artifacts

Tagged releases publish cross-compiled archives for:

- `linux-amd64`
- `linux-arm64`
- `darwin-amd64`
- `darwin-arm64`

Archive naming scheme:

```text
codex-switcher_<version>_<os>_<arch>.tar.gz
```

Example:

```text
codex-switcher_1.0.0_linux_amd64.tar.gz
```

Each release also publishes `SHA256SUMS.txt`.

## Maintainer Release Steps

1. Verify the working tree is clean.
2. Run:

```bash
export PATH=/home/den/.local/bin:$PATH
make fmt
make test
make vet
make lint
make smoke
```

3. Create and push a version tag:

```bash
git checkout main
git pull --ff-only
git tag v1.0.0
git push origin v1.0.0
```

4. Wait for the GitHub Actions `release` workflow to publish the archives and `SHA256SUMS.txt`.
5. Verify the release page contains all expected artifacts.

## Installer Script

The repository root includes `install.sh`. It downloads the matching artifact from the latest GitHub Release and installs `codex-switcher` into `~/.local/bin` by default.

Examples:

```bash
curl -fsSL https://raw.githubusercontent.com/Denver2003/codex-multiaccount-switcher/main/install.sh | sh
```

Pin a version:

```bash
curl -fsSL https://raw.githubusercontent.com/Denver2003/codex-multiaccount-switcher/main/install.sh | VERSION=v1.0.0 sh
```

## Homebrew

This project uses a dedicated tap approach.

Tap repository:

```text
Denver2003/homebrew-codex-switcher
```

Expected end-user install command:

```bash
brew tap Denver2003/codex-switcher
brew install codex-switcher
```

### Maintainer Formula Process

1. Create or maintain the dedicated tap repository named `homebrew-codex-switcher`.
2. After each tagged release, open `SHA256SUMS.txt` from the GitHub Release.
3. Copy the checksum for the target artifact you want Homebrew to install. The default formula should point to the macOS archive appropriate for Homebrew distribution policy.
4. Update the formula URL and SHA256 in the tap repository.
5. Commit and push the tap update.

Template formula:

```ruby
class CodexSwitcher < Formula
  desc "Local CLI utility for switching Codex auth profiles"
  homepage "https://github.com/Denver2003/codex-multiaccount-switcher"
  url "https://github.com/Denver2003/codex-multiaccount-switcher/releases/download/v1.0.0/codex-switcher_1.0.0_darwin_arm64.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_SHA256"
  version "1.0.0"

  def install
    bin.install "codex-switcher"
  end

  test do
    system "#{bin}/codex-switcher", "--help"
  end
end
```

Repository template file:

```text
packaging/homebrew/codex-switcher.rb.tmpl
```
