#!/usr/bin/env bash
# Copyright 2026 The kpt Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# generate_docs.sh - Generate/sync Hugo doc pages from function README.md + metadata.yaml
#
# Usage:
#   ./scripts/generate_docs/generate_docs.sh [--dry-run] [function-name]
#
# Examples:
#   ./scripts/generate_docs/generate_docs.sh                        # sync all functions
#   ./scripts/generate_docs/generate_docs.sh set-namespace          # sync latest minor
#   ./scripts/generate_docs/generate_docs.sh --dry-run              # preview all
#
# If function-name is provided, only that function is processed.
# Otherwise all functions in functions/go/ are processed.
#
# The script always targets the latest minor version from git tags.
# Older minor version docs are historical snapshots and should not be regenerated.
#
# Logic per function:
#   1. Read metadata.yaml — skip if hidden: true
#   2. Check git tags — skip if no release tags
#   3. Determine latest minor version from tags
#   4. Generate documentation/content/en/<fn>/<minor>/_index.md
#      from front matter (metadata.yaml) + body (README.md)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
FUNCTIONS_DIR="${REPO_ROOT}/functions/go"
DOCS_DIR="${REPO_ROOT}/documentation/content/en"

DRY_RUN=false
TARGET_FN=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --help|-h)
      echo "Usage: $(basename "$0") [--dry-run] [function-name]"
      echo ""
      echo "Examples:"
      echo "  $(basename "$0")                        # sync all functions"
      echo "  $(basename "$0") set-namespace          # sync latest minor"
      echo "  $(basename "$0") --dry-run              # preview all"
      exit 0
      ;;
    *) TARGET_FN="$1"; shift ;;
  esac
done

# Parse tags from metadata.yaml (YAML list under tags:)
parse_tags() {
  local metadata_file="$1"
  sed -n '/^tags:/,/^[^ ]/p' "$metadata_file" \
    | grep '^ *- ' \
    | sed 's/^ *- //' \
    | paste -sd ', ' -
}

# Parse description from metadata.yaml
parse_description() {
  local metadata_file="$1"
  grep '^description:' "$metadata_file" | sed 's/^description: *//'
}

# Check if hidden: true in metadata.yaml
is_hidden() {
  local metadata_file="$1"
  grep -q '^hidden: *true' "$metadata_file" 2>/dev/null
}

# Get all git tags for a function
get_fn_tags() {
  local fn_name="$1"
  git tag | grep "^functions/go/${fn_name}/v" || true
}

# Get latest minor version from git tags for a function
# e.g. functions/go/set-namespace/v0.4.5 -> v0.4
get_latest_minor() {
  local fn_name="$1"
  local tags
  tags=$(get_fn_tags "$fn_name")
  if [ -z "$tags" ]; then
    return
  fi
  echo "$tags" \
    | sed "s|functions/go/${fn_name}/||" \
    | sort -V \
    | tail -1 \
    | sed 's/\.[0-9]*$//'
}


# Generate _index.md content from metadata + README
generate_index() {
  local fn_name="$1"
  local tags="$2"
  local description="$3"
  local readme_file="$4"

  # Front matter
  cat <<EOF
---
title: "${fn_name}"
linkTitle: "${fn_name}"
tags: "${tags}"
weight: 4
description: |
  ${description}
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}
EOF

  # README body (skip the "# title" first line)
  tail -n +2 "$readme_file"
}

process_function() {
  local fn_name="$1"
  local fn_dir="${FUNCTIONS_DIR}/${fn_name}"
  local metadata_file="${fn_dir}/metadata.yaml"
  local readme_file="${fn_dir}/README.md"

  if [ ! -f "$metadata_file" ]; then
    echo "SKIP ${fn_name}: no metadata.yaml"
    return
  fi

  if [ ! -f "$readme_file" ]; then
    echo "SKIP ${fn_name}: no README.md"
    return
  fi

  local first_line
  first_line=$(head -1 "$readme_file")
  if [[ ! "$first_line" =~ ^#\ .+ ]]; then
    echo "WARN ${fn_name}: README.md first line is not '# function-name', got: ${first_line}"
    return
  fi

  if is_hidden "$metadata_file"; then
    echo "SKIP ${fn_name}: hidden"
    return
  fi

  local latest_minor
  latest_minor=$(get_latest_minor "$fn_name")
  if [ -z "$latest_minor" ]; then
    echo "SKIP ${fn_name}: no release tags"
    return
  fi

  local tags description
  tags=$(parse_tags "$metadata_file")
  description=$(parse_description "$metadata_file")

  local doc_dir="${DOCS_DIR}/${fn_name}/${latest_minor}"
  local doc_file="${doc_dir}/_index.md"

  if [ "$DRY_RUN" = true ]; then
    echo "WOULD generate ${doc_file}"
    return
  fi

  mkdir -p "$doc_dir"
  generate_index "$fn_name" "$tags" "$description" "$readme_file" > "$doc_file"
  echo "GENERATED ${doc_file}"
}

# Main
echo "Validating metadata..."
if ! bash "${SCRIPT_DIR}/validate_metadata.sh"; then
  echo "ERROR: metadata validation failed, aborting doc generation"
  exit 1
fi

echo ""
echo "Fetching tags..."
git fetch --tags --quiet 2>/dev/null || true

if [ -n "$TARGET_FN" ]; then
  if [ ! -d "${FUNCTIONS_DIR}/${TARGET_FN}" ]; then
    echo "ERROR: function '${TARGET_FN}' not found in ${FUNCTIONS_DIR}"
    exit 1
  fi
  process_function "$TARGET_FN"
else
  for fn_dir in "${FUNCTIONS_DIR}"/*/; do
    fn_name=$(basename "$fn_dir")
    [ "$fn_name" = "_template" ] && continue
    process_function "$fn_name"
  done
fi
