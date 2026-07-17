#!/usr/bin/env bash
# Generate docs/COMMANDS.md from live gpd help output.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${1:-${ROOT}/docs/COMMANDS.md}"
BIN="${GPD_BIN:-${ROOT}/bin/gpd}"

if [ ! -x "${BIN}" ]; then
  echo "Building gpd..."
  (cd "${ROOT}" && make build)
  BIN="${ROOT}/bin/gpd"
fi

if [ ! -x "${BIN}" ]; then
  echo "error: gpd binary not found at ${BIN}" >&2
  exit 1
fi

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT

{
  echo "# Command Reference Guide"
  echo
  echo "This file is generated from live CLI help output."
  echo "For authoritative command behavior, also use:"
  echo
  echo '```bash'
  echo "gpd --help"
  echo "gpd <command> --help"
  echo "gpd <command> <subcommand> --help"
  echo '```'
  echo
  echo "To regenerate:"
  echo
  echo '```bash'
  echo "make generate-command-docs"
  echo '```'
  echo
  echo "Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo
  echo "## Global help"
  echo
  echo '```'
  "${BIN}" --help 2>&1 || true
  echo '```'
  echo
  echo "## Auth"
  echo
  echo '```'
  "${BIN}" auth --help 2>&1 || true
  echo '```'
  echo
  echo "## Publish"
  echo
  echo '```'
  "${BIN}" publish --help 2>&1 || true
  echo '```'
  echo
  echo "## Validate"
  echo
  echo '```'
  "${BIN}" validate --help 2>&1 || true
  echo '```'
  echo
  echo "## Workflow"
  echo
  echo '```'
  "${BIN}" workflow --help 2>&1 || true
  echo '```'
  echo
  echo "## Command families (top-level)"
  echo
  # Extract top-level command names from help (lines like "  auth ...")
  "${BIN}" --help 2>&1 | awk '
    /^Commands:/ { in_cmds=1; next }
    /^Flags:/ { in_cmds=0 }
    /^Extension/ { in_cmds=0 }
    in_cmds && /^  [a-z]/ {
      # first token after indent
      line=$0
      sub(/^  /, "", line)
      split(line, a, /[ \t]/)
      name=a[1]
      if (name != "" && name != "gpd") {
        print "- `" name "`"
      }
    }
  '
  echo
} >"${TMP}"

mkdir -p "$(dirname "${OUT}")"
mv "${TMP}" "${OUT}"
echo "Wrote ${OUT}"
