#! /bin/bash
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

# This script requires TAG and CURRENT_FUNCTION to be set.
# CURRENT_FUNCTION is the target kpt function. e.g. set-namespace.
# TAG can be any valid docker tags. If the TAG is semver e.g. v1.2.3, shorter
# version of this semver will be tagged too. e.g. v1.2 and v1.
# DEFAULT_GCR is the desired container registry e.g. gcr.io/kpt-fn. This is
# optional. If not set, the default value gcr.io/kpt-fn-contrib will be used.
# If GCR_REGISTRY is set, it will override DEFAULT_GCR.
# example 1:
# Invocation: DEFAULT_GCR=gcr.io/kpt-fn CURRENT_FUNCTION=set-namespace TAG=v1.2.3 go-function-release.sh build
# It builds gcr.io/kpt-fn/set-namespace:v1.2.3, gcr.io/kpt-fn/set-namespace:v1.2
# and gcr.io/kpt-fn/set-namespace:v1.
# Invocation: DEFAULT_GCR=gcr.io/kpt-fn CURRENT_FUNCTION=set-namespace TAG=v1.2.3 go-function-release.sh push
# It pushes the above 3 images.
# example 2:
# Invocation: CURRENT_FUNCTION=set-namespace TAG=unstable go-function-release.sh build
# It builds gcr.io/kpt-fn/set-namespace:unstable.
# Invocation: CURRENT_FUNCTION=set-namespace TAG=unstable go-function-release.sh push
# It pushes gcr.io/kpt-fn/set-namespace:unstable.

# This script currently is used in functions/go/Makefile.

set -euo pipefail

scripts_dir="$(dirname "$0")"
# git-tag-parser.sh has been shell-checked separately.
# shellcheck source=/dev/null
source "${scripts_dir}"/git-tag-parser.sh
# shellcheck source=/dev/null
source "${scripts_dir}"/docker.sh

versions=$(get_versions "${TAG}")

FUNCTION_TYPE="$2"
EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS:-}"

case "$1" in
  build)
    for version in ${versions}; do
      docker_build "load" "${FUNCTION_TYPE}" "go" "${CURRENT_FUNCTION}" "${version}" "${EXTRA_BUILD_ARGS}"
    done
    ;;
  push)
    for version in ${versions}; do
      docker_build "push" "${FUNCTION_TYPE}" "go" "${CURRENT_FUNCTION}" "${version}"
    done
    ;;
  *)
    echo "Usage: $0 {build|push}"
    exit 1
esac
