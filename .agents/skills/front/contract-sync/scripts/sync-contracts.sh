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

# Walk up to find the monorepo root (contains modelcraft-front/ and modelcraft-backend/)
_dir="$SCRIPT_DIR"
while [ "$_dir" != "/" ]; do
  if [ -d "$_dir/modelcraft-front" ] && [ -d "$_dir/modelcraft-backend" ]; then
    MONO_ROOT="$_dir"
    break
  fi
  _dir="$(cd "$_dir/.." && pwd)"
done
if [ -z "${MONO_ROOT:-}" ]; then
  echo "Error: Cannot find monorepo root (directory with both modelcraft-front/ and modelcraft-backend/)" >&2
  exit 1
fi

PROJECT_ROOT="$MONO_ROOT/modelcraft-front"

SRC="$MONO_ROOT/modelcraft-backend/api"
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
