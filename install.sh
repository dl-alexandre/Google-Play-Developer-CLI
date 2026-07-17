#!/usr/bin/env bash
# install.sh — install gpd with mandatory release checksum verification.
#
# Environment:
#   INSTALL_DIR              install destination (default: ~/.local/bin or /usr/local/bin)
#   GPD_INSTALL_INSECURE=1   allow install when checksum verification cannot run
#   GPD_INSTALL_ASSET        (test) path to a local release asset instead of downloading
#   GPD_INSTALL_CHECKSUMS    (test) path to a local checksums file
#   GPD_INSTALL_SKIP_EXTRACT=1 (test) treat asset as the binary itself (no tar)
set -euo pipefail

REPO="${GPD_INSTALL_REPO:-dl-alexandre/Google-Play-Developer-CLI}"
BIN_NAME="gpd"
DEFAULT_INSTALL_DIR="/usr/local/bin"
if [ -n "${HOME:-}" ]; then
  DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
fi
INSTALL_DIR="${INSTALL_DIR:-${DEFAULT_INSTALL_DIR}}"
DOWNLOAD_MAX_ATTEMPTS=3
DOWNLOAD_RETRY_DELAY_SECONDS=1

curl_with_retry() {
  local attempt=1
  while true; do
    if curl -fsSL "$@"; then
      return 0
    fi
    if [ "${attempt}" -ge "${DOWNLOAD_MAX_ATTEMPTS}" ]; then
      return 1
    fi
    attempt=$((attempt + 1))
    echo "Download failed; retrying (${attempt}/${DOWNLOAD_MAX_ATTEMPTS})..." >&2
    sleep "${DOWNLOAD_RETRY_DELAY_SECONDS}"
  done
}

# verification_unavailable aborts unless GPD_INSTALL_INSECURE=1.
verification_unavailable() {
  local reason="$1"
  if [ "${GPD_INSTALL_INSECURE:-}" = "1" ]; then
    echo "!!! WARNING: ${reason}" >&2
    echo "!!! GPD_INSTALL_INSECURE=1 is set; installing WITHOUT checksum verification." >&2
    return 0
  fi
  echo "Error: ${reason}" >&2
  echo "Refusing to install without SHA-256 checksum verification." >&2
  echo "If you understand the risk and must install anyway, re-run with GPD_INSTALL_INSECURE=1." >&2
  exit 1
}

# verify_checksum ASSET_PATH CHECKSUMS_PATH ASSET_NAME
# Returns 0 on match, 1 on mismatch / missing entry. Exits 1 on hard failure
# when not insecure.
verify_checksum() {
  local asset_path="$1"
  local checksums_path="$2"
  local asset_name="$3"

  if [ ! -f "${checksums_path}" ]; then
    verification_unavailable "Checksums file not found: ${checksums_path}"
    return 0
  fi

  local expected
  expected="$(awk -v asset="${asset_name}" '$2 == asset || $2 == "*" asset { print $1; exit }' "${checksums_path}")"
  if [ -z "${expected}" ]; then
    verification_unavailable "Asset ${asset_name} not found in checksums file."
    return 0
  fi

  local actual=""
  if command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${asset_path}" | awk '{print $1}')"
  elif command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${asset_path}" | awk '{print $1}')"
  else
    verification_unavailable "No checksum tool (shasum/sha256sum) available."
    return 0
  fi

  if [ "${expected}" != "${actual}" ]; then
    echo "Error: Checksum verification failed for ${asset_name}." >&2
    echo "  expected: ${expected}" >&2
    echo "  actual:   ${actual}" >&2
    return 1
  fi

  echo "Checksum verified for ${asset_name}"
  return 0
}

# Allow sourcing for unit tests without running install.
if [ "${GPD_INSTALL_LIB_ONLY:-}" = "1" ]; then
  return 0 2>/dev/null || exit 0
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64|amd64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

case "${OS}" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

if [ -n "${GPD_INSTALL_ASSET:-}" ]; then
  # Test / offline path: local fixture asset.
  ASSET_PATH="${GPD_INSTALL_ASSET}"
  ASSET_NAME="$(basename "${ASSET_PATH}")"
  if [ ! -f "${ASSET_PATH}" ]; then
    echo "Error: GPD_INSTALL_ASSET not found: ${ASSET_PATH}" >&2
    exit 1
  fi
  cp "${ASSET_PATH}" "${TMP_DIR}/${ASSET_NAME}"
  ASSET_PATH="${TMP_DIR}/${ASSET_NAME}"

  if [ -n "${GPD_INSTALL_CHECKSUMS:-}" ]; then
    CHECKSUMS_PATH="${GPD_INSTALL_CHECKSUMS}"
  else
    verification_unavailable "GPD_INSTALL_ASSET set without GPD_INSTALL_CHECKSUMS."
    CHECKSUMS_PATH=""
  fi

  if [ -n "${CHECKSUMS_PATH}" ]; then
    if ! verify_checksum "${ASSET_PATH}" "${CHECKSUMS_PATH}" "${ASSET_NAME}"; then
      exit 1
    fi
  fi

  if [ "${GPD_INSTALL_SKIP_EXTRACT:-}" = "1" ]; then
    BIN_SRC="${ASSET_PATH}"
  else
    tar -xzf "${ASSET_PATH}" -C "${TMP_DIR}"
    BIN_SRC="${TMP_DIR}/${BIN_NAME}"
  fi
else
  # Production path: download latest release.
  LATEST="$(curl_with_retry "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')"
  VERSION="${LATEST#v}"
  if [ -z "${VERSION}" ]; then
    echo "Failed to get latest version" >&2
    exit 1
  fi

  echo "Installing gpd ${LATEST}..."

  FILENAME="gpd_${VERSION}_${OS}_${ARCH}.tar.gz"
  BASE_URL="https://github.com/${REPO}/releases/download/${LATEST}"
  BIN_URL="${BASE_URL}/${FILENAME}"
  CHECKSUMS_ASSET="checksums.txt"
  # Prefer gpd_*_checksums.txt if present; fall back to checksums.txt
  CHECKSUMS_URL="${BASE_URL}/gpd_${VERSION}_checksums.txt"
  CHECKSUMS_URL_ALT="${BASE_URL}/checksums.txt"

  curl_with_retry "${BIN_URL}" -o "${TMP_DIR}/${FILENAME}"

  CHECKSUMS_PATH="${TMP_DIR}/checksums.txt"
  if curl_with_retry "${CHECKSUMS_URL}" -o "${CHECKSUMS_PATH}"; then
    :
  elif curl_with_retry "${CHECKSUMS_URL_ALT}" -o "${CHECKSUMS_PATH}"; then
    :
  else
    verification_unavailable "Could not download checksums from release ${LATEST}."
    CHECKSUMS_PATH=""
  fi

  if [ -n "${CHECKSUMS_PATH}" ] && [ -f "${CHECKSUMS_PATH}" ]; then
    if ! verify_checksum "${TMP_DIR}/${FILENAME}" "${CHECKSUMS_PATH}" "${FILENAME}"; then
      exit 1
    fi
  fi

  tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"
  BIN_SRC="${TMP_DIR}/${BIN_NAME}"
fi

if [ ! -f "${BIN_SRC}" ]; then
  echo "Error: binary not found after extract: ${BIN_SRC}" >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}"
if [ -w "${INSTALL_DIR}" ]; then
  mv "${BIN_SRC}" "${INSTALL_DIR}/${BIN_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${BIN_SRC}" "${INSTALL_DIR}/${BIN_NAME}"
fi

chmod +x "${INSTALL_DIR}/${BIN_NAME}"
echo "gpd installed to ${INSTALL_DIR}/${BIN_NAME}"
if [ "${GPD_INSTALL_SKIP_VERSION:-}" != "1" ]; then
  "${INSTALL_DIR}/${BIN_NAME}" version || true
fi
