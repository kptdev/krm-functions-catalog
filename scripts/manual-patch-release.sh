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
#   MANUAL_PATCH_FUNCTIONS  Comma-separated function names (must match output of: make list-functions),
#                           or the single word "all" (case-insensitive) to release every name from
#                           `make list-functions`. With "all", functions with no prior
#                           functions/go/<name>/vMAJOR.MINOR.PATCH tag are skipped with a notice.
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

# Same names as FUNCTIONS in functions/go/Makefile (see target list-functions).
mapfile -t allowed < <(make -s list-functions)

if [[ ${#allowed[@]} -eq 0 ]]; then
  echo "::error::make list-functions produced no output" >&2
  exit 1
fi

_trim_space() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
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

# Bulk mode: every function from `make list-functions`.
if [[ "${outer_trim,,}" == "all" ]]; then
  expand_all=1
  mapfile -t functions < <(printf '%s\n' "${allowed[@]}" | sort -u)
else
  # Explicit list: comma-separated names (strict if missing tags).
  IFS=',' read -ra raw_parts <<< "${functions_input}"
  for part in "${raw_parts[@]}"; do
    fn="$(_trim_space "$part")"
    [[ -z "$fn" ]] && continue
    if ! is_allowed "$fn"; then
      echo "::error::Function '$fn' is not in output of: make list-functions" >&2
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
      tail -n 1 || true
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

  # Fail fast before pushing images, to avoid overwriting an existing release.
  short_tag="${fn}/${next_ver}"

  if git rev-parse "$long_tag" >/dev/null 2>&1; then
    echo "::error::Tag ${long_tag} already exists locally" >&2
    exit 1
  fi
  if git rev-parse "$short_tag" >/dev/null 2>&1; then
    echo "::error::Tag ${short_tag} already exists locally" >&2
    exit 1
  fi

  # Multi-arch push (see go-function-release.sh / docker.sh).
  (cd functions/go && make func-push TAG="${next_ver}" CURRENT_FUNCTION="${fn}" DEFAULT_CR="ghcr.io/${repo}")

  # Long tag (full path) then short tag (<fn>/v…) to match release.yaml behavior.
  git tag "$long_tag" "$sha"
  git push origin "refs/tags/${long_tag}"

  git fetch origin "refs/tags/${long_tag}"
  oid="$(git rev-parse FETCH_HEAD^{})"
  short_tag="${fn}/${next_ver}"
  git tag "$short_tag" "$oid"
  git push origin "refs/tags/${short_tag}"

  # Same registry path as make func-push (DEFAULT_CR + function name).
  image_ref="ghcr.io/${repo}/${fn}:${next_ver}"

  # GitHub-generated notes between previous_tag_name and target; subshell + trap cleans mktemp on failure.
  notes_json="$(gh api "repos/${repo}/releases/generate-notes" \
    -f tag_name="${long_tag}" \
    -f target_commitish="${sha}" \
    -f previous_tag_name="${prev_long}")"

  gen_body="$(printf '%s' "$notes_json" | jq -r .body)"
  # Match existing catalog releases (see scripts/release-krm-functions.sh): "<fn> vX.Y.Z".
  release_title="${fn} ${next_ver}"

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
      --title "${release_title}" \
      --notes-file "${notes_file}"
  )
done
