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

# Generate GitHub-style release notes scoped to a single folder.
#
# Calls the GitHub Release Notes API, then keeps only bullets whose linked PR
# also appears in git history for the folder (e.g. functions/go/set-namespace/).
#
# Prerequisites: gh (authenticated), git.
#
# Example:
#   scripts/generate-folder-release-notes.sh \
#     --function upsert-resource \
#     --previous-tag v0.2.3 \
#     --new-tag v0.2.4 \
#     --ref main

set -euo pipefail

GITHUB_REPOSITORY=${GITHUB_REPOSITORY:-"kptdev/krm-functions-catalog"}

usage() {
  cat <<'EOF'
Usage: generate-folder-release-notes.sh [options]

Generate release notes for changes under functions/go/NAME/ by filtering the
output of the GitHub Release Notes API.

Options:
  --function NAME         Function name under functions/go/ (required)
  --previous-tag VERSION  Previous release version (required, e.g. v0.2.3)
  --new-tag VERSION       New release version (required, e.g. v0.2.4)
  --ref REF               Commitish for the new release (default: HEAD)
  -o, --output FILE       Write notes to FILE instead of stdout
  -h, --help              Show this help

EOF
}

fail() {
  echo "generate-folder-release-notes.sh: $*" >&2
  exit 1
}

normalize_version() {
  local version="$1"
  if [[ "$version" =~ ^v ]]; then
    printf '%s' "$version"
  else
    printf 'v%s' "$version"
  fi
}

function_tag() {
  local version
  version="$(normalize_version "$1")"
  printf 'functions/go/%s/%s' "$function_name" "$version"
}

function_name=""
previous_version=""
new_version=""
ref="HEAD"
output_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --function)
      [[ $# -ge 2 ]] || fail "missing value for --function"
      function_name="$2"
      shift 2
      ;;
    --previous-tag)
      [[ $# -ge 2 ]] || fail "missing value for --previous-tag"
      previous_version="$2"
      shift 2
      ;;
    --new-tag)
      [[ $# -ge 2 ]] || fail "missing value for --new-tag"
      new_version="$2"
      shift 2
      ;;
    --ref)
      [[ $# -ge 2 ]] || fail "missing value for --ref"
      ref="$2"
      shift 2
      ;;
    -o | --output)
      [[ $# -ge 2 ]] || fail "missing value for $1"
      output_file="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1 (try --help)"
      ;;
  esac
done

if [[ -z "$function_name" ]]; then
  fail "--function is required (try --help)"
fi
if [[ -z "$previous_version" ]]; then
  fail "--previous-tag is required"
fi
if [[ -z "$new_version" ]]; then
  fail "--new-tag is required"
fi

previous_tag="$(function_tag "$previous_version")"
new_tag="$(function_tag "$new_version")"
folder_path="functions/go/${function_name}"

declare -A folder_pr_numbers=()

extract_pr_number() {
  local line="$1"
  if [[ "$line" =~ pull/([0-9]+) ]]; then
    printf '%s' "${BASH_REMATCH[1]}"
  fi
}

extract_pr_number_from_subject() {
  local subject="$1"
  if [[ "$subject" =~ \(#([0-9]+)\)$ ]]; then
    printf '%s' "${BASH_REMATCH[1]}"
  fi
}

# Discover PRs that touched the folder via git history.
collect_folder_pr_numbers() {
  local subject pr_number
  while IFS= read -r subject; do
    [[ -z "$subject" ]] && continue
    pr_number="$(extract_pr_number_from_subject "$subject")"
    [[ -n "$pr_number" ]] || continue
    folder_pr_numbers["$pr_number"]=1
  done < <(git log "${previous_tag}..${ref}" --format='%s' -- "${folder_path}/")
}

pr_in_folder() {
  [[ -n "${folder_pr_numbers[$1]:-}" ]]
}

filter_bullet_line() {
  local line="$1"
  local pr_number

  pr_number="$(extract_pr_number "$line")"
  [[ -n "$pr_number" ]] || return 1
  pr_in_folder "$pr_number"
}

collect_folder_pr_numbers

parse_and_filter_notes() {
  local body="$1"
  local section=""
  local whats_changed=()
  local new_contributors=()
  local full_changelog=""

  while IFS= read -r line || [[ -n "$line" ]]; do
    case "$line" in
      "## What's Changed")
        section="whats_changed"
        continue
        ;;
      "## New Contributors")
        section="new_contributors"
        continue
        ;;
      "**Full Changelog"*)
        full_changelog="$line"
        section=""
        continue
        ;;
    esac

    if [[ "$line" == "* "* ]]; then
      if filter_bullet_line "$line"; then
        case "$section" in
          whats_changed) whats_changed+=("$line") ;;
          new_contributors) new_contributors+=("$line") ;;
        esac
      fi
    fi
  done <<< "$body"

  printf '%s\n' "## What's Changed"
  if [[ ${#whats_changed[@]} -eq 0 ]]; then
    printf "* No changes under \`%s/\` in this range.\n" "$folder_path"
  else
    printf '%s\n' "${whats_changed[@]}"
  fi

  if [[ ${#new_contributors[@]} -gt 0 ]]; then
    printf '\n%s\n' "## New Contributors"
    printf '%s\n' "${new_contributors[@]}"
  fi

  if [[ -n "$full_changelog" ]]; then
    printf '\n%s\n' "$full_changelog"
  fi
}

notes_body="$(
  gh api "repos/${GITHUB_REPOSITORY}/releases/generate-notes" \
    -f tag_name="$new_tag" \
    -f previous_tag_name="$previous_tag" \
    -f target_commitish="$ref" \
    --jq .body
)"

filtered_notes="$(parse_and_filter_notes "$notes_body")"

if [[ -n "$output_file" ]]; then
  printf '%s\n' "$filtered_notes" > "$output_file"
else
  printf '%s\n' "$filtered_notes"
fi
