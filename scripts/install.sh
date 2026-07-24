#!/bin/sh

set -eu

repository=${AITRIAGE_REPOSITORY:-dodobrands/aitriage}
install_dir=${AITRIAGE_INSTALL_DIR:-/usr/local/bin}
user_install_dir=${AITRIAGE_USER_INSTALL_DIR:-${HOME:?HOME is required}/.local/bin}
skip_setup=${AITRIAGE_SKIP_SETUP:-0}

fail() {
  printf 'AITriage install error: %s\n' "$1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

case $(uname -s) in
  Darwin) release_os=Darwin ;;
  Linux) release_os=Linux ;;
  *) fail "unsupported operating system; use the GitHub release page" ;;
esac

case $(uname -m) in
  arm64 | aarch64) release_arch=arm64 ;;
  x86_64 | amd64) release_arch=x86_64 ;;
  *) fail "unsupported CPU architecture; use the GitHub release page" ;;
esac

if [ -n "${AITRIAGE_VERSION:-}" ]; then
  version=${AITRIAGE_VERSION#v}
  tag=v$version
else
  latest_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' \
    "https://github.com/$repository/releases/latest") || \
    fail "could not resolve the latest official release"
  tag=${latest_url##*/}
  case $tag in
    v[0-9]*) version=${tag#v} ;;
    *) fail "the latest release tag is invalid: $tag" ;;
  esac
fi

asset=aitriage_${version}_${release_os}_${release_arch}.tar.gz
base_url=${AITRIAGE_RELEASE_BASE_URL:-https://github.com/$repository/releases}
download_url=$base_url/download/$tag
temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/aitriage-install.XXXXXX")
trap 'rm -rf "$temporary_dir"' EXIT HUP INT TERM

printf 'Downloading AITriage %s for %s/%s...\n' "$version" "$release_os" "$release_arch"
curl -fsSL "$download_url/$asset" -o "$temporary_dir/$asset" || \
  fail "release asset is unavailable: $asset"
curl -fsSL "$download_url/checksums.txt" -o "$temporary_dir/checksums.txt" || \
  fail "release checksum file is unavailable"

expected=$(awk -v name="$asset" '$2 == name || $2 == "*" name { print $1; exit }' \
  "$temporary_dir/checksums.txt")
[ -n "$expected" ] || fail "checksum for $asset is missing"

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$temporary_dir/$asset" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$temporary_dir/$asset" | awk '{print $1}')
else
  fail "sha256sum or shasum is required to verify the download"
fi
[ "$actual" = "$expected" ] || fail "release checksum verification failed"

tar -xzf "$temporary_dir/$asset" -C "$temporary_dir"
[ -f "$temporary_dir/aitriage" ] || fail "release archive does not contain aitriage"

# A caller-provided install directory is commonly a new path under a writable
# temporary or workspace directory. Create it as the current user first; the
# old flow treated every missing directory as privileged and invoked sudo even
# when its parent was writable.
if [ ! -d "$install_dir" ]; then
  install -d "$install_dir" 2>/dev/null || true
fi

if [ -d "$install_dir" ] && [ -w "$install_dir" ]; then
  install -m 0755 "$temporary_dir/aitriage" "$install_dir/aitriage"
elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
  printf 'Using pre-authorized administrator access for %s.\n' "$install_dir"
  sudo -n install -d "$install_dir"
  sudo -n install -m 0755 "$temporary_dir/aitriage" "$install_dir/aitriage"
elif command -v sudo >/dev/null 2>&1 &&
  [ "${AITRIAGE_NONINTERACTIVE:-0}" != 1 ] &&
  (: </dev/tty) 2>/dev/null; then
  printf 'Administrator access is required to install into %s.\n' "$install_dir"
  sudo install -d "$install_dir"
  sudo install -m 0755 "$temporary_dir/aitriage" "$install_dir/aitriage"
else
  install_dir=$user_install_dir
  install -d "$install_dir" || fail "could not create user install directory: $install_dir"
  install -m 0755 "$temporary_dir/aitriage" "$install_dir/aitriage" || \
    fail "could not install into user directory: $install_dir"
  printf 'No interactive administrator access is available.\n'
  printf 'AITriage was installed for the current user: %s\n' "$install_dir/aitriage"
  case :${PATH:-}: in
    *:"$install_dir":*) ;;
    *) printf 'For future shells, add %s to PATH.\n' "$install_dir" ;;
  esac
fi

binary=$install_dir/aitriage
installed_version=$("$binary" version)
case $installed_version in
  "AITriage $version" | "AITriage $tag") ;;
  *) fail "release binary reported an unexpected version: $installed_version" ;;
esac
printf '%s\n' "$installed_version"

if [ "$skip_setup" = 1 ]; then
  printf 'AITriage CLI installed successfully.\n'
  exit 0
fi

printf 'Preparing the complete scanner bundle...\n'
"$binary" setup --full
printf 'AITriage is ready.\n'
