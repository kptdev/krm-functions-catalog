#! /usr/bin/env bash

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

# Manual patch release for functions/go KRM functions (used by GitHub Actions and runnable locally).
#
# Environment (required unless noted):
#   MANUAL_PATCH_FUNCTIONS  Comma-separated function names (must match FUNCTIONS in functions/go/Makefile),
#                           or the single word "all" (case-insensitive) to release every Makefile-listed
#                           function except keys in MANUAL_PATCH_EXCLUDE_FROM_ALL below. With "all",
#                           functions with no prior functions/go/<name>/vMAJOR.MINOR.PATCH tag are skipped
#                           with a notice.
#   GITHUB_REPOSITORY       owner/repo (e.g. kptdev/krm-functions-catalog).
#   GITHUB_SHA              Commit SHA to tag and release (images built from this tree).
#   GH_TOKEN                Token for gh api / gh release create (optional in dry-run).
#
# Optional:
#   MANUAL_PATCH_DRY_RUN    If "true", only print planned versions (no push, tags, or releases).
#
# Prerequisites when not dry-run: docker logged in to GHCR; gh, jq, git, make, Go toolchains as in CI.

set -euo pipefail

scripts_dir="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "${scripts_dir}/.." && pwd)"
cd "${repo_root}"

functions_input="${MANUAL_PATCH_FUNCTIONS:-}"
repo="${GITHUB_REPOSITORY:-}"
sha="${GITHUB_SHA:-}"
dry_run="${MANUAL_PATCH_DRY_RUN:-false}"

if [[ -z "$functions_input" ]]; then
  echo "::error::MANUAL_PATCH_FUNCTIONS is not set" >&2
  exit 1
fi
if [[ -z "$repo" ]]; then
  echo "::error::GITHUB_REPOSITORY is not set" >&2
  exit 1
fi
if [[ -z "$sha" ]]; then
  echo "::error::GITHUB_SHA is not set" >&2
  exit 1
fi

# Tags/releases appear as the GitHub Actions bot in the UI.
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git config user.name "github-actions[bot]"

# Same list make func-push uses (see FUNCTIONS in functions/go/Makefile).
mapfile -t allowed < <(awk '
  /^FUNCTIONS :=/ { p = 1; next }
  p && /^# Targets for running/ { exit }
  p && /^[[:space:]]+[a-z0-9-]+/ {
    line = $0
    sub(/^[[:space:]]+/, "", line)
    sub(/[[:space:]]*\\$/, "", line)
    if (line != "") print line
  }
' functions/go/Makefile)

if [[ ${#allowed[@]} -eq 0 ]]; then
  echo "::error::Could not parse FUNCTIONS from functions/go/Makefile" >&2
  exit 1
fi

# Keys = function names omitted when MANUAL_PATCH_FUNCTIONS is "all" (no SemVer line / not shipped).
declare -A MANUAL_PATCH_EXCLUDE_FROM_ALL=(
  [sleep]=1
)

_trim_space() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

is_excluded_from_all() {
  [[ -v MANUAL_PATCH_EXCLUDE_FROM_ALL["$1"] ]]
}

is_allowed() {
  local n="$1"
  for a in "${allowed[@]}"; do
    [[ "$n" == "$a" ]] && return 0
  done
  return 1
}

outer_trim="$(_trim_space "${functions_input}")"
declare -a functions=()
expand_all=0

# Bulk mode: every Makefile function except MANUAL_PATCH_EXCLUDE_FROM_ALL.
if [[ "${outer_trim,,}" == "all" ]]; then
  expand_all=1
  for a in "${allowed[@]}"; do
    if is_excluded_from_all "$a"; then
      continue
    fi
    if [[ ! -f "functions/go/${a}/metadata.yaml" ]]; then
      echo "::error::Missing functions/go/${a}/metadata.yaml" >&2
      exit 1
    fi
    functions+=("$a")
  done
  if [[ ${#functions[@]} -eq 0 ]]; then
    echo "::error::No functions left after applying MANUAL_PATCH_EXCLUDE_FROM_ALL" >&2
    exit 1
  fi
  mapfile -t functions < <(printf '%s\n' "${functions[@]}" | sort -u)
else
  # Explicit list: comma-separated names (strict if missing tags or excluded utility).
  IFS=',' read -ra raw_parts <<< "${functions_input}"
  for part in "${raw_parts[@]}"; do
    fn="$(_trim_space "$part")"
    [[ -z "$fn" ]] && continue
    if is_excluded_from_all "$fn"; then
      echo "::error::Function '${fn}' is not SemVer-released (excluded from catalog patch releases)." >&2
      exit 1
    fi
    if ! is_allowed "$fn"; then
      echo "::error::Function '$fn' is not in FUNCTIONS in functions/go/Makefile" >&2
      exit 1
    fi
    if [[ ! -f "functions/go/${fn}/metadata.yaml" ]]; then
      echo "::error::Missing functions/go/${fn}/metadata.yaml" >&2
      exit 1
    fi
    functions+=("$fn")
  done
  if [[ ${#functions[@]} -eq 0 ]]; then
    echo "::error::No function names after parsing MANUAL_PATCH_FUNCTIONS" >&2
    exit 1
  fi
fi

for fn in "${functions[@]}"; do
  echo "===== ${fn} ====="

  # Latest strict SemVer long tag for this function (sort -V orders versions correctly).
  prev_long="$(
    git tag -l "functions/go/${fn}/v*" |
      grep -E -- '/v[0-9]+\.[0-9]+\.[0-9]+$' |
      sort -V |
      tail -n 1
  )"

  if [[ -z "${prev_long}" ]]; then
    if [[ "${expand_all}" -eq 1 ]]; then
      echo "::notice::Skipping ${fn}: no prior semver tag functions/go/${fn}/vMAJOR.MINOR.PATCH"
      continue
    fi
    echo "::error::No prior semver tag functions/go/${fn}/vMAJOR.MINOR.PATCH; cannot infer next patch." >&2
    exit 1
  fi

  ver="${prev_long##*/}"
  v="${ver#v}"
  IFS=. read -r major minor patch <<< "${v}"
  next_ver="v${major}.${minor}.$((patch + 1))"
  long_tag="functions/go/${fn}/${next_ver}"

  echo "Previous: ${prev_long} -> Next: ${long_tag}"

  if [[ "$dry_run" == "true" ]]; then
    echo "(dry_run) would func-push, tag, and gh release create for ${long_tag}"
    continue
  fi

  # Multi-arch push (see go-function-release.sh / docker.sh).
  (cd functions/go && make func-push TAG="${next_ver}" CURRENT_FUNCTION="${fn}" DEFAULT_CR="ghcr.io/${repo}")

  if git rev-parse "$long_tag" >/dev/null 2>&1; then
    echo "::error::Tag ${long_tag} already exists locally" >&2
    exit 1
  fi

  # Long tag (full path) then short tag (<fn>/v…) to match release.yaml behavior.
  git tag "$long_tag" "$sha"
  git push origin "refs/tags/${long_tag}"

  git fetch origin "refs/tags/${long_tag}"
  oid="$(git rev-parse FETCH_HEAD^{})"
  short_tag="${fn}/${next_ver}"
  git tag -f "$short_tag" "$oid"
  git push -f origin "refs/tags/${short_tag}"

  image_base="$(grep '^image:' "functions/go/${fn}/metadata.yaml" | awk '{print $2}' | tr -d '"')"
  if [[ -z "$image_base" ]]; then
    echo "::error::No image: field in functions/go/${fn}/metadata.yaml" >&2
    exit 1
  fi
  image_ref="${image_base}:${next_ver}"

  # GitHub-generated notes between previous_tag_name and target; subshell + trap cleans mktemp on failure.
  notes_json="$(gh api "repos/${repo}/releases/generate-notes" \
    -f tag_name="${long_tag}" \
    -f target_commitish="${sha}" \
    -f previous_tag_name="${prev_long}")"

  gen_body="$(printf '%s' "$notes_json" | jq -r .body)"
  gen_name="$(printf '%s' "$notes_json" | jq -r .name)"

  (
    notes_file="$(mktemp)"
    trap 'rm -f "$notes_file"' EXIT
    {
      echo "## Container image"
      echo
      echo '```'
      echo "${image_ref}"
      echo '```'
      echo
      printf '%s\n' "${gen_body}"
    } >"${notes_file}"

    gh release create "${long_tag}" \
      --repo "${repo}" \
      --title "${gen_name}" \
      --notes-file "${notes_file}"
  )
done
