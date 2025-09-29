#!/bin/bash
#
# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERNUM='0|[1-9][0-9]*'
SEMVER_REGEX="^[vV]?($VERNUM)\\.($VERNUM)\\.($VERNUM)$"

# get_versions return versions: vX.Y.Z, vX.Y-latest and vX-latest if the input matches
# SEMVER_REGEX. Otherwise, it returns the input.
# example 1:
# Invocation: get_versions v1.2.3
# Return: v1.2.3 v1.2-latest v1-latest
# example 2:
# Invocation: get_versions latest
# Return: latest
function get_versions {
  local version=$1
  if [[ "${version}" =~ $SEMVER_REGEX ]]; then
    local major=${BASH_REMATCH[1]}
    local minor=${BASH_REMATCH[2]}
    local patch=${BASH_REMATCH[3]}

    echo v"${major}"."${minor}"."${patch}"
    echo v"${major}"."${minor}"-latest
    echo v"${major}"-latest
  else
    echo "${version}"
  fi
}

# parse_git_tag expects the GIT_TAG environment variable to be set. Otherwise,
# it stops and exit with 0.
# It GIT_TAG is set, it splits the GIT_TAG by "/" and returns the desired part
# depending on the input argument.
# The format of GIT_TAG can be either functions/<fn-lang>/<fn-name>/<version>
# (recommended) or <fn-lang>/<fn-name>/<version>.
# example 1:
# Invocation: GIT_TAG=functions/ts/kubeval/v1.2.3 parse_git_tag lang
# Return: ts
# example 2:
# Invocation: GIT_TAG=functions/ts/kubeval/v1.2.3 parse_git_tag fn_name
# Return: kubeval
# example 3:
# Invocation: GIT_TAG=functions/ts/kubeval/v1.2.3 parse_git_tag fn_ver
# Return: v1.2.3
function parse_git_tag {
  # If GIT_TAG is not set, just skip processing.
  if [ -z "${GIT_TAG}" ];
  then
    exit 0
  fi

  # Split GIT_TAG by '/'. e.g. if GIT_TAG is "go/set-namespace/v1.2.3", we will
  # get "go", "set-namespace" and "v1.2.3".
  IFS='/' read -ra GIT_TAG_ARRAY <<< "${GIT_TAG}"
  if [ "${GIT_TAG_ARRAY[0]}" == "functions" ]; then
    fn_lang=${GIT_TAG_ARRAY[1]}
    fn_name=${GIT_TAG_ARRAY[2]}
    fn_ver=${GIT_TAG_ARRAY[3]}
  elif [ "${GIT_TAG_ARRAY[0]}" == "contrib" ]; then
    fn_lang=${GIT_TAG_ARRAY[2]}
    fn_name=${GIT_TAG_ARRAY[3]}
    fn_ver=${GIT_TAG_ARRAY[4]}
  else
    fn_lang=${GIT_TAG_ARRAY[0]}
    fn_name=${GIT_TAG_ARRAY[1]}
    fn_ver=${GIT_TAG_ARRAY[2]}
  fi

  case "$1" in
    lang)
      echo "${fn_lang}"
      ;;
    name)
      echo "${fn_name}"
      ;;
    version)
      echo "${fn_ver}"
      ;;
    *)
      echo "Usage: $0 {lang|name|version}"
      exit 1
  esac
}
