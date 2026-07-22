#!/bin/sh

set -eu

repository_root=$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd)
test_root=$(mktemp -d "${TMPDIR:-/tmp}/aitriage-installer-test.XXXXXX")
trap 'rm -rf "$test_root"' EXIT HUP INT TERM

case $(uname -s) in
  Darwin) release_os=Darwin ;;
  Linux) release_os=Linux ;;
  *) exit 0 ;;
esac

case $(uname -m) in
  arm64 | aarch64) release_arch=arm64 ;;
  x86_64 | amd64) release_arch=x86_64 ;;
  *) exit 0 ;;
esac

version=9.9.9
tag=v$version
asset=aitriage_${version}_${release_os}_${release_arch}.tar.gz
release_dir=$test_root/releases/download/$tag
payload_dir=$test_root/payload
install_dir=$test_root/bin
mkdir -p "$release_dir" "$payload_dir"

cat >"$payload_dir/aitriage" <<'EOF'
#!/bin/sh
if [ "${1:-}" = version ]; then
  echo "AITriage 9.9.9"
  exit 0
fi
exit 1
EOF
chmod +x "$payload_dir/aitriage"
tar -czf "$release_dir/$asset" -C "$payload_dir" aitriage

if command -v sha256sum >/dev/null 2>&1; then
  checksum=$(sha256sum "$release_dir/$asset" | awk '{print $1}')
else
  checksum=$(shasum -a 256 "$release_dir/$asset" | awk '{print $1}')
fi
printf '%s  %s\n' "$checksum" "$asset" >"$release_dir/checksums.txt"

AITRIAGE_VERSION=$version \
AITRIAGE_RELEASE_BASE_URL=file://$test_root/releases \
AITRIAGE_INSTALL_DIR=$install_dir \
AITRIAGE_SKIP_SETUP=1 \
  sh "$repository_root/scripts/install.sh"

test -d "$install_dir"
test "$("$install_dir/aitriage" version)" = "AITriage $version"

# AI IDEs commonly have no controlling terminal. If the system destination is
# not writable and sudo cannot run non-interactively, installation must fall
# back to a per-user path instead of hanging or failing at a password prompt.
restricted_dir=$test_root/restricted-bin
fallback_dir=$test_root/user-bin
fake_bin=$test_root/fake-bin
mkdir -p "$restricted_dir" "$fake_bin"
chmod 0555 "$restricted_dir"
cat >"$fake_bin/sudo" <<'EOF'
#!/bin/sh
exit 1
EOF
chmod +x "$fake_bin/sudo"
PATH="$fake_bin:$PATH" \
AITRIAGE_VERSION=$version \
AITRIAGE_RELEASE_BASE_URL=file://$test_root/releases \
AITRIAGE_INSTALL_DIR=$restricted_dir \
AITRIAGE_USER_INSTALL_DIR=$fallback_dir \
AITRIAGE_NONINTERACTIVE=1 \
AITRIAGE_SKIP_SETUP=1 \
  sh "$repository_root/scripts/install.sh"
test "$("$fallback_dir/aitriage" version)" = "AITriage $version"

printf 'corrupt' >>"$release_dir/$asset"
if AITRIAGE_VERSION=$version \
  AITRIAGE_RELEASE_BASE_URL=file://$test_root/releases \
  AITRIAGE_INSTALL_DIR=$install_dir \
  AITRIAGE_SKIP_SETUP=1 \
    sh "$repository_root/scripts/install.sh" >/dev/null 2>&1; then
  echo "installer accepted a corrupted release asset" >&2
  exit 1
fi

echo "installer E2E passed"
