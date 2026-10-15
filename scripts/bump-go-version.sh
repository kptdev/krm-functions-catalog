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

# Bump the Go version declared across the repository.
#
# Updates the `go` directive in every module's go.mod (excluding vendored
# dependencies and the archived/ tree) and the GO_VERSION used by the Netlify
# documentation build. CI reads the `go` directive via setup-go to select the
# toolchain, so this is the single change needed when a dependency raises its
# minimum Go version (e.g. kpt/mdtogo@main requiring go >= 1.26.8).
#
# The script matches the `go` directive line exactly, so dependency version
# strings that happen to contain the old version are left untouched. It is
# idempotent: re-running after a partial change only updates what still differs.
#
# Target version: pass it explicitly, or let the script derive it from the kpt
# toolchain this repo must stay compatible with. The functions/go Makefile
# installs github.com/kptdev/kpt/mdtogo@main, so its `go` directive is the
# version to align to. Passing "kpt" (or omitting the version) resolves that
# directive from the Go module proxy -- a single HTTPS call, no clone. An
# explicit version is preferred for reproducible, reviewable bumps; "kpt" is a
# convenience for "whatever kpt needs right now".
#
# Usage:
#   scripts/bump-go-version.sh <new-version> [old-version]
#   scripts/bump-go-version.sh kpt [old-version]
#   scripts/bump-go-version.sh                 # same as: kpt
#
# Arguments:
#   new-version   Target Go version, e.g. 1.26.8. The literal "kpt" (or no
#                 argument) derives it from kpt/mdtogo@main instead.
#   old-version   Version to replace. Optional; if omitted, every `go` directive
#                 and GO_VERSION found is set to the target regardless of its
#                 current value.
#
# Environment:
#   KPT_MDTOGO_MODULE   Override the module queried in "kpt" mode
#                       (default: github.com/kptdev/kpt/mdtogo@main).
#   GOPROXY             Honoured when resolving the kpt version
#                       (default: https://proxy.golang.org).
#
# Examples:
#   scripts/bump-go-version.sh 1.26.8
#   scripts/bump-go-version.sh 1.26.8 1.26.5
#   scripts/bump-go-version.sh kpt 1.26.5
#
# This script only rewrites version strings. It does not run `go mod tidy`,
# which must run under the matching Go toolchain and is already provided as a
# repo-wide target. After bumping, reconcile and verify with Go >= <new-version>
# installed:
#   git diff --stat
#   make tidy                        # repo-wide `go mod tidy` over every module
#   (cd functions/go && make test)   # build + unit tests

set -euo pipefail

if [ "$#" -gt 2 ]; then
  echo "Usage: $0 [<new-version>|kpt] [old-version]" >&2
  echo "  e.g. $0 1.26.8        # set every go directive to 1.26.8" >&2
  echo "       $0 1.26.8 1.26.5 # only replace 1.26.5 -> 1.26.8" >&2
  echo "       $0 kpt           # derive target from kpt/mdtogo@main" >&2
  exit 1
fi

# Default to kpt-derived target when no version is given.
NEW="${1:-kpt}"
OLD="${2:-}"

version_re='^[0-9]+\.[0-9]+\.[0-9]+$'

# Resolve the target from kpt's mdtogo module via the Go module proxy.
# This is the module the functions/go Makefile installs, so its `go`
# directive is exactly what this repo must stay compatible with.
resolve_kpt_go_version() {
  local module="${KPT_MDTOGO_MODULE:-github.com/kptdev/kpt/mdtogo@main}"
  local proxy="${GOPROXY:-https://proxy.golang.org}"
  # GOPROXY may be a comma/pipe-separated list; use the first entry.
  proxy="${proxy%%[,|]*}"
  if [ "$proxy" = "off" ] || [ "$proxy" = "direct" ]; then
    proxy="https://proxy.golang.org"
  fi

  local path="${module%@*}"
  local ref="${module#*@}"

  # Resolve the ref (e.g. "main") to a concrete version, then read its go.mod.
  local info version gomod
  info="$(curl -fsSL "${proxy}/${path}/@v/${ref}.info" 2>/dev/null)" || {
    echo "Error: could not resolve ${module} via ${proxy}" >&2
    return 1
  }
  version="$(printf '%s' "$info" | grep -oE '"Version":"[^"]+"' | head -n1 | sed -E 's/.*":"(.*)"/\1/')"
  if [ -z "$version" ]; then
    echo "Error: could not parse version from proxy response for ${module}" >&2
    return 1
  fi
  gomod="$(curl -fsSL "${proxy}/${path}/@v/${version}.mod" 2>/dev/null)" || {
    echo "Error: could not fetch go.mod for ${path}@${version}" >&2
    return 1
  }
  printf '%s' "$gomod" | grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1 | awk '{print $2}'
}

if [ "$NEW" = "kpt" ]; then
  echo "Resolving target Go version from ${KPT_MDTOGO_MODULE:-github.com/kptdev/kpt/mdtogo@main} ..."
  NEW="$(resolve_kpt_go_version)" || exit 1
  # The proxy go directive may be MAJOR.MINOR; normalise to MAJOR.MINOR.0.
  if [[ "$NEW" =~ ^[0-9]+\.[0-9]+$ ]]; then
    NEW="${NEW}.0"
  fi
  echo "kpt targets Go ${NEW}"
fi

# Validate version format (MAJOR.MINOR.PATCH).
if [[ ! "$NEW" =~ $version_re ]]; then
  echo "Error: target version '$NEW' is not of the form MAJOR.MINOR.PATCH" >&2
  exit 1
fi
if [ -n "$OLD" ] && [[ ! "$OLD" =~ $version_re ]]; then
  echo "Error: old-version '$OLD' is not of the form MAJOR.MINOR.PATCH" >&2
  exit 1
fi

# Always operate from the repository root.
cd "$(git rev-parse --show-toplevel)"

if [ -n "$OLD" ]; then
  echo "Bumping Go ${OLD} -> ${NEW}"
else
  echo "Setting Go version to ${NEW} (all go directives)"
fi

changed=0

# 1) Module files: update the `go` directive only.
#    Skip vendored dependencies and the archived/ tree.
while IFS= read -r -d '' f; do
  if [ -n "$OLD" ]; then
    pattern="^go ${OLD}\$"
    replacement="go ${NEW}"
  else
    # Match any patch of the same series we are moving to is not assumed;
    # replace whatever single go directive exists.
    pattern='^go [0-9]+\.[0-9]+\.[0-9]+$'
    replacement="go ${NEW}"
  fi

  if grep -qE "$pattern" "$f"; then
    current="$(grep -E '^go [0-9]+\.[0-9]+\.[0-9]+' "$f" | head -n1 | awk '{print $2}')"
    if [ "$current" != "$NEW" ]; then
      sed -i -E "s/${pattern}/${replacement}/" "$f"
      echo "  updated: $f (go ${current} -> ${NEW})"
      changed=$((changed + 1))
    fi
  fi
done < <(find . -name go.mod -not -path '*/vendor/*' -not -path './archived/*' -print0)

# 2) Netlify documentation build environment.
NETLIFY="documentation/netlify.toml"
if [ -f "$NETLIFY" ]; then
  current="$(grep -E 'GO_VERSION = "[0-9]+\.[0-9]+\.[0-9]+"' "$NETLIFY" | head -n1 | sed -E 's/.*"([0-9.]+)".*/\1/' || true)"
  if [ -n "$current" ] && [ "$current" != "$NEW" ]; then
    if [ -z "$OLD" ] || [ "$current" = "$OLD" ]; then
      sed -i -E "s/GO_VERSION = \"[0-9]+\.[0-9]+\.[0-9]+\"/GO_VERSION = \"${NEW}\"/" "$NETLIFY"
      echo "  updated: $NETLIFY (GO_VERSION ${current} -> ${NEW})"
      changed=$((changed + 1))
    fi
  fi
fi

echo "Done. Files changed: ${changed}"

# 3) Report any lingering references to the old version (sanity check).
if [ -n "$OLD" ]; then
  echo "Remaining '${OLD}' references in go.mod / netlify.toml (excluding archived/):"
  if grep -rn --include=go.mod --include=netlify.toml -- "$OLD" . | grep -v './archived/'; then
    echo "  ^ review the above; some may be unrelated dependency versions"
  else
    echo "  none"
  fi
fi
