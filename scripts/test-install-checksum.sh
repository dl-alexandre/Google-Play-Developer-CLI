#!/usr/bin/env bash
# Fixture tests for install.sh checksum verification and release asset naming.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRATCH="${1:-}"
if [ -z "${SCRATCH}" ]; then
  echo "usage: $0 <scratch-dir>" >&2
  exit 2
fi
mkdir -p "${SCRATCH}"
LOG="${SCRATCH}/install-verify.log"
: >"${LOG}"

log() { echo "$@" | tee -a "${LOG}"; }

# shellcheck source=../install.sh
GPD_INSTALL_LIB_ONLY=1 source "${ROOT}/install.sh"

WORK="$(mktemp -d "${SCRATCH}/install-fixture.XXXXXX")"
trap 'rm -rf "${WORK}"' EXIT

ASSET_NAME="gpd_test_asset.bin"
ASSET_PATH="${WORK}/${ASSET_NAME}"
echo "hello-gpd-install-fixture" >"${ASSET_PATH}"

if command -v shasum >/dev/null 2>&1; then
  GOOD_SUM="$(shasum -a 256 "${ASSET_PATH}" | awk '{print $1}')"
else
  GOOD_SUM="$(sha256sum "${ASSET_PATH}" | awk '{print $1}')"
fi
BAD_SUM="0000000000000000000000000000000000000000000000000000000000000000"

GOOD_CHECKSUMS="${WORK}/checksums-good.txt"
BAD_CHECKSUMS="${WORK}/checksums-bad.txt"
# GoReleaser / shasum format: "<sha256>  <filename>" (two spaces)
echo "${GOOD_SUM}  ${ASSET_NAME}" >"${GOOD_CHECKSUMS}"
echo "${BAD_SUM}  ${ASSET_NAME}" >"${BAD_CHECKSUMS}"

log "== match: verify_checksum should succeed =="
if verify_checksum "${ASSET_PATH}" "${GOOD_CHECKSUMS}" "${ASSET_NAME}" >>"${LOG}" 2>&1; then
  log "PASS: match verified"
else
  log "FAIL: expected match to succeed"
  exit 1
fi

log "== mismatch: verify_checksum should fail (exit non-zero) =="
set +e
verify_checksum "${ASSET_PATH}" "${BAD_CHECKSUMS}" "${ASSET_NAME}" >>"${LOG}" 2>&1
rc=$?
set -e
if [ "${rc}" -ne 0 ]; then
  log "PASS: mismatch exited ${rc}"
else
  log "FAIL: expected mismatch to fail"
  exit 1
fi

log "== end-to-end local install with good checksums =="
INSTALL_DIR="${WORK}/bin"
mkdir -p "${INSTALL_DIR}"
GPD_INSTALL_ASSET="${ASSET_PATH}" \
GPD_INSTALL_CHECKSUMS="${GOOD_CHECKSUMS}" \
GPD_INSTALL_SKIP_EXTRACT=1 \
GPD_INSTALL_SKIP_VERSION=1 \
INSTALL_DIR="${INSTALL_DIR}" \
  bash "${ROOT}/install.sh" >>"${LOG}" 2>&1
if [ -x "${INSTALL_DIR}/gpd" ]; then
  log "PASS: install placed binary at ${INSTALL_DIR}/gpd"
else
  log "FAIL: binary missing after install"
  exit 1
fi

log "== end-to-end local install with bad checksums must fail =="
INSTALL_DIR2="${WORK}/bin2"
mkdir -p "${INSTALL_DIR2}"
set +e
GPD_INSTALL_ASSET="${ASSET_PATH}" \
GPD_INSTALL_CHECKSUMS="${BAD_CHECKSUMS}" \
GPD_INSTALL_SKIP_EXTRACT=1 \
GPD_INSTALL_SKIP_VERSION=1 \
INSTALL_DIR="${INSTALL_DIR2}" \
  bash "${ROOT}/install.sh" >>"${LOG}" 2>&1
rc=$?
set -e
if [ "${rc}" -ne 0 ]; then
  log "PASS: bad checksum install exited ${rc}"
else
  log "FAIL: bad checksum install should not succeed"
  exit 1
fi
# Ensure we did not leave a "verified" install of a bad asset as success path.
if [ -f "${INSTALL_DIR2}/gpd" ]; then
  # install may have partially run before verify — verify happens before mv when
  # using GPD_INSTALL_ASSET path; binary should not be installed on mismatch.
  log "NOTE: binary present after failed install (checking verify runs before install)"
  # In our script verify runs before install; if file exists, fail.
  log "FAIL: binary should not be installed after checksum mismatch"
  exit 1
else
  log "PASS: no binary installed after mismatch"
fi

# ---------------------------------------------------------------------------
# Release asset naming + checksum file format (matches .goreleaser.yml / install.sh)
# ---------------------------------------------------------------------------
log "== goreleaser config: SHA-256 + expected checksum asset name =="
GORELEASER_YML="${ROOT}/.goreleaser.yml"
if [ ! -f "${GORELEASER_YML}" ]; then
  log "FAIL: missing ${GORELEASER_YML}"
  exit 1
fi

if ! grep -qE 'algorithm:[[:space:]]*sha256' "${GORELEASER_YML}"; then
  log "FAIL: .goreleaser.yml must set checksum.algorithm to sha256 (install.sh uses SHA-256)"
  exit 1
fi
if ! grep -qE 'name_template:[[:space:]]*"gpd_\{\{ \.Version \}\}_checksums\.txt"' "${GORELEASER_YML}"; then
  log "FAIL: .goreleaser.yml checksum name_template must be gpd_{{ .Version }}_checksums.txt"
  exit 1
fi
if ! grep -q 'x86_64' "${GORELEASER_YML}"; then
  log "FAIL: .goreleaser.yml archive name must map amd64 -> x86_64 for install.sh"
  exit 1
fi
# Archive template must include ProjectName, Version, Os, Arch (with x86_64 branch)
if ! grep -qE 'ProjectName.*Version.*Os' "${GORELEASER_YML}"; then
  log "FAIL: .goreleaser.yml archive name_template must include ProjectName/Version/Os"
  exit 1
fi
log "PASS: goreleaser checksum algorithm and naming aligned with install.sh"

log "== simulated release checksums file (shasum format + archive names) =="
VERSION="9.9.9"
OS_LIST="darwin linux"
ARCH_LIST="x86_64 arm64"
CHECKSUMS_RELEASE="${WORK}/gpd_${VERSION}_checksums.txt"
: >"${CHECKSUMS_RELEASE}"

for os in ${OS_LIST}; do
  for arch in ${ARCH_LIST}; do
    fname="gpd_${VERSION}_${os}_${arch}.tar.gz"
    fpath="${WORK}/${fname}"
    # Fake archive payload (not a real tar; we only verify naming + checksum lines)
    printf 'fake-archive-%s-%s\n' "${os}" "${arch}" >"${fpath}"
    if command -v shasum >/dev/null 2>&1; then
      sum="$(shasum -a 256 "${fpath}" | awk '{print $1}')"
    else
      sum="$(sha256sum "${fpath}" | awk '{print $1}')"
    fi
    # Standard GoReleaser/shasum line: hash, two spaces, filename
    printf '%s  %s\n' "${sum}" "${fname}" >>"${CHECKSUMS_RELEASE}"
  done
done

# Validate file shape: 64-hex hash, two spaces, gpd_VERSION_OS_ARCH.tar.gz
line_count=0
while IFS= read -r line || [ -n "${line}" ]; do
  line_count=$((line_count + 1))
  if ! printf '%s\n' "${line}" | grep -qE '^[0-9a-f]{64}  gpd_[0-9]+\.[0-9]+\.[0-9]+_(darwin|linux)_(x86_64|arm64)\.tar\.gz$'; then
    log "FAIL: bad checksum line format: ${line}"
    exit 1
  fi
done <"${CHECKSUMS_RELEASE}"

if [ "${line_count}" -ne 4 ]; then
  log "FAIL: expected 4 darwin/linux archive lines, got ${line_count}"
  exit 1
fi
log "PASS: checksum file has ${line_count} valid shasum lines"

# install.sh FILENAME pattern must match one of those assets for current host mapping
SAMPLE_ASSET="gpd_${VERSION}_darwin_arm64.tar.gz"
if ! verify_checksum "${WORK}/${SAMPLE_ASSET}" "${CHECKSUMS_RELEASE}" "${SAMPLE_ASSET}" >>"${LOG}" 2>&1; then
  log "FAIL: verify_checksum failed for release-style asset ${SAMPLE_ASSET}"
  exit 1
fi
log "PASS: verify_checksum accepts gpd_\${VERSION}_\${OS}_\${ARCH}.tar.gz entry"

# Also accept fallback name checksums.txt with same body
cp "${CHECKSUMS_RELEASE}" "${WORK}/checksums.txt"
if ! verify_checksum "${WORK}/${SAMPLE_ASSET}" "${WORK}/checksums.txt" "${SAMPLE_ASSET}" >>"${LOG}" 2>&1; then
  log "FAIL: verify_checksum failed against fallback checksums.txt"
  exit 1
fi
log "PASS: fallback checksums.txt format works"

# Document expected URL names for operators (matches install.sh)
log "Expected release assets (install.sh):"
log "  archive:   gpd_\${VERSION}_\${OS}_\${ARCH}.tar.gz  (OS=darwin|linux, ARCH=x86_64|arm64)"
log "  checksums: gpd_\${VERSION}_checksums.txt  (preferred) or checksums.txt"

log "ALL INSTALL VERIFY CHECKS PASSED"
exit 0
