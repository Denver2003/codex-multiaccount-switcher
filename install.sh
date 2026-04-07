#!/usr/bin/env sh
set -eu

OWNER_REPO="${OWNER_REPO:-Denver2003/codex-multiaccount-switcher}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Linux) printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *) fail "Unsupported operating system: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) fail "Unsupported architecture: $(uname -m)" ;;
  esac
}

download() {
  url="$1"
  output="$2"

  if need_cmd curl; then
    curl -fsSL "$url" -o "$output"
    return
  fi

  if need_cmd wget; then
    wget -qO "$output" "$url"
    return
  fi

  fail "curl or wget is required"
}

resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    printf '%s' "$VERSION"
    return
  fi

  api_url="https://api.github.com/repos/$OWNER_REPO/releases/latest"
  metadata_file="$TMP_DIR/latest-release.json"
  download "$api_url" "$metadata_file"

  version_line=$(sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\(v[^"]*\)".*$/\1/p' "$metadata_file" | head -n 1)
  [ -n "$version_line" ] || fail "Unable to determine latest release version"
  printf '%s' "$version_line"
}

path_contains() {
  case ":$PATH:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

need_cmd tar || fail "tar is required"
mkdir -p "$INSTALL_DIR"
[ -w "$INSTALL_DIR" ] || fail "Install directory is not writable: $INSTALL_DIR"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

os=$(detect_os)
arch=$(detect_arch)
tag=$(resolve_version)
version="${tag#v}"
archive="codex-switcher_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$OWNER_REPO/releases/download/$tag/$archive"

printf 'Installing %s from %s\n' "$archive" "$url"
download "$url" "$TMP_DIR/$archive"
tar -xzf "$TMP_DIR/$archive" -C "$TMP_DIR"
install -m 755 "$TMP_DIR/codex-switcher" "$INSTALL_DIR/codex-switcher"

printf 'Installed codex-switcher to %s/codex-switcher\n' "$INSTALL_DIR"
if ! path_contains "$INSTALL_DIR"; then
  printf 'Add %s to PATH to run codex-switcher directly.\n' "$INSTALL_DIR"
fi
printf 'Run: codex-switcher --help\n'
