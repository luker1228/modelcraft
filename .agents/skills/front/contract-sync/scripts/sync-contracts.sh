#!/usr/bin/env bash
# sync-contracts.sh — Sync Go backend API contracts to frontend contract directory
#
# Source:      ./modelcraft-go/api/
# Destination: ./modelcraft-front/contract/
#
# This script copies the backend's API contract definitions (GraphQL schemas
# and OpenAPI specs) so the frontend can reference them without reaching
# across the monorepo boundary.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Walk up to find modelcraft-front project root (contains package.json)
_dir="$SCRIPT_DIR"
while [ "$_dir" != "/" ]; do
  if [ -f "$_dir/package.json" ]; then
    PROJECT_ROOT="$_dir"
    break
  fi
  _dir="$(cd "$_dir/.." && pwd)"
done
if [ -z "${PROJECT_ROOT:-}" ]; then
  echo "Error: Cannot find project root (no package.json found)" >&2
  exit 1
fi

MONO_ROOT="$(cd "$PROJECT_ROOT/.." && pwd)"

SRC="$MONO_ROOT/modelcraft-go/api"
DST="$PROJECT_ROOT/contract"

# Ensure source exists
if [ ! -d "$SRC" ]; then
  echo "Error: Source directory not found: $SRC" >&2
  exit 1
fi

# Clean and recreate destination
rm -rf "$DST"
mkdir -p "$DST"

# Copy GraphQL schemas
cp -r "$SRC/graph" "$DST/graph"

# Copy OpenAPI specs
cp -r "$SRC/openapi" "$DST/openapi"

# Remove non-contract files (generated, maintenance docs, examples)
rm -f "$DST/openapi/openapi.yaml"       # generated bundle
rm -f "$DST/openapi/oapi-codegen.yaml"  # codegen config
rm -rf "$DST/openapi/examples"          # example payloads, not contracts
rm -f "$DST/openapi/README.md"          # backend maintenance guide

echo "Contract sync complete: $SRC -> $DST"
echo ""
echo "Synced files:"
find "$DST" -type f | sort | while read -r f; do
  echo "  ${f#$DST/}"
done
