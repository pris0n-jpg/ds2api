#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

LINT_BIN="${GOLANGCI_LINT_BIN:-golangci-lint}"

# v2 separates formatters from linters; enforce both in one entrypoint.
if [[ "$LINT_BIN" == *" "* ]]; then
  eval "$LINT_BIN fmt --diff -c .golangci.yml"
  eval "$LINT_BIN run -c .golangci.yml"
else
  "$LINT_BIN" fmt --diff -c .golangci.yml
  "$LINT_BIN" run -c .golangci.yml
fi
