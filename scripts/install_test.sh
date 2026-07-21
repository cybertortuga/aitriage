#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
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
mkdir -p "$release_dir" "$payload_dir" "$install_dir"

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

test "$("$install_dir/aitriage" version)" = "AITriage $version"

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
