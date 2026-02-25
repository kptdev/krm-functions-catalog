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

# This script generates the 'gh' commands to create the releases for KRM functions.
# The single parameter filters the old releases list and bumps the patch version for
# each KRM function on the list that matches the entered parameter.

set -euo pipefail

bump_patch() {
  old_ver="$1"

  # shellcheck disable=SC2001
  ver_major_minor=$(echo "$old_ver" | sed 's/\.[0-9]*$//')
  # shellcheck disable=SC2001
  ver_patch_old=$(echo "$old_ver" | sed 's/^v[0-9]*\.[0-9]*\.//')
  ver_patch_new=$((ver_patch_old+1))
  echo "$ver_major_minor.$ver_patch_new"
}

upgrade_from() {
  name="$1"
  old_ver="$2"
  tag="$3"

  tag_prefix=$(echo "$tag" | sed 's/\/v[^/]*$//')
  new_ver="$(bump_patch "$old_ver")"

  release_tag="$tag_prefix/$new_ver"
  safe_release_tag=$(printf '%q' "$release_tag")
  title="$name $new_ver"
  safe_title=$(printf '%q' "$title")

  echo "gh release create $safe_release_tag --title $safe_title --generate-notes"
}

if [ -z "$1" ]
then
 echo "Usage: $0 pattern";
 echo "  pattern: A pattern that filters the \"old\" revisions that we want to bump from"
 exit 1
fi

# NOTE: This parsing assumes 'gh release list' outputs at least four space-separated columns.
gh release list --limit 10000 | grep "$1" | sed 's/Latest//' | awk 'NF>=4 {printf("%s,%s,%s,%s\n", $1,$2,$3,$4)}' | \
while read -r line
do
  IFS=',' read -r -a old_release <<< "$line"
  if [ "${#old_release[@]}" -lt 4 ]; then
    echo "Warning: unexpected format from 'gh release list', skipping line: $line" >&2
    continue
  fi
  upgrade_from "${old_release[0]}" "${old_release[1]}" "${old_release[2]}"
done
