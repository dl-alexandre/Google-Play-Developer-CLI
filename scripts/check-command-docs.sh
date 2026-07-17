#!/usr/bin/env bash
# Fail if docs/COMMANDS.md is stale relative to live gpd help.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${GPD_BIN:-${ROOT}/bin/gpd}"
DOC="${ROOT}/docs/COMMANDS.md"

if [ ! -x "${BIN}" ]; then
  (cd "${ROOT}" && make build)
  BIN="${ROOT}/bin/gpd"
fi

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT

GPD_BIN="${BIN}" bash "${ROOT}/scripts/generate-command-docs.sh" "${TMP}"

# Compare taxonomy-relevant sections: ensure live top-level families appear in committed doc.
HELP="$("${BIN}" --help 2>&1 || true)"
for family in auth publish workflow validate; do
  if ! echo "${HELP}" | grep -E "^  ${family}( |$)" >/dev/null 2>&1; then
    # help may wrap differently — also accept anywhere as command token
    if ! echo "${HELP}" | grep -w "${family}" >/dev/null 2>&1; then
      echo "error: live help missing expected family: ${family}" >&2
      exit 1
    fi
  fi
  if ! grep -E "\`${family}\`| ${family} " "${DOC}" >/dev/null 2>&1 && ! grep -w "${family}" "${DOC}" >/dev/null 2>&1; then
    echo "error: docs/COMMANDS.md missing family ${family}; run make generate-command-docs" >&2
    exit 1
  fi
done

# Auth subcommands present in live help must appear in doc.
AUTH_HELP="$("${BIN}" auth --help 2>&1 || true)"
for sub in status login logout list switch check doctor diagnose init; do
  if echo "${AUTH_HELP}" | grep -E "auth ${sub}" >/dev/null 2>&1; then
    if ! grep -E "auth ${sub}| ${sub}" "${DOC}" >/dev/null 2>&1; then
      echo "error: docs/COMMANDS.md missing auth subcommand ${sub}" >&2
      exit 1
    fi
  fi
done

# Must not invent absent auth subcommands that we know we do not have.
# (Placeholder: keep list empty unless we intentionally remove commands.)

# Drift: regenerated content should match committed file for the help bodies.
# Compare ignoring the Generated: timestamp line.
filter() { grep -v '^Generated:' "$1" | grep -v '^```$' ; }
if ! diff -u <(filter "${DOC}") <(filter "${TMP}") >/dev/null 2>&1; then
  echo "error: docs/COMMANDS.md is stale relative to live help." >&2
  echo "Run: make generate-command-docs && git add docs/COMMANDS.md" >&2
  diff -u <(filter "${DOC}") <(filter "${TMP}") | head -80 >&2 || true
  exit 1
fi

echo "docs/COMMANDS.md is up to date with live help"
