#! /bin/bash
#
# Copyright 2021 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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
# CURRENT_FUNCTION is the target kpt function. e.g. kubeval.
# TAG can be any valid docker tags. If the TAG is semver e.g. v1.2.3, shorter
# versions of this semver will be tagged too. e.g. v1.2 and v1.
# DEFAULT_CR is the desired container registry e.g. ghcr.io/kptdev/krm-functions-catalog. This is
# optional. If not set, the default value ghcr.io/kptdev/krm-functions-catalog/contrib will be used.
# If CR_REGISTRY is set, it will override DEFAULT_CR.
# example 1:
# Invocation: DEFAULT_CR=ghcr.io/kptdev/krm-functions-catalog CURRENT_FUNCTION=kubeval TAG=v1.2.3 ts-function-release.sh build
# It builds ghcr.io/kptdev/krm-functions-catalog/kubeval:v1.2.3, ghcr.io/kptdev/krm-functions-catalog/kubeval:v1.2 and
# ghcr.io/kptdev/krm-functions-catalog/kubeval:v1.
# Invocation: DEFAULT_CR=ghcr.io/kptdev/krm-functions-catalog CURRENT_FUNCTION=kubeval TAG=v1.2.3 ts-function-release.sh push
# It pushes the above 3 images.
# example 2:
# Invocation: CURRENT_FUNCTION=kubeval TAG=latest ts-function-release.sh build
# It builds ghcr.io/kptdev/krm-functions-catalog/kubeval:latest.
# Invocation: CURRENT_FUNCTION=kubeval TAG=latest ts-function-release.sh push
# It pushes ghcr.io/kptdev/krm-functions-catalog/kubeval:latest.

# This script currently is used in functions/ts/Makefile.

set -euo pipefail

scripts_dir="$(dirname "$0")"
# git-tag-parser.sh has been shell-checked separately.
# shellcheck source=/dev/null
source "${scripts_dir}"/git-tag-parser.sh
# shellcheck source=/dev/null
source "${scripts_dir}"/docker.sh

versions=$(get_versions "${TAG}")
FUNCTION_TYPE="$2"

# https://github.com/kptdev/kpt/issues/1394
# This make it work for npm 7.0.0+
export npm_package_kpt_docker_repo_base="${CR_REGISTRY}"

case "$1" in
  build)
    for version in ${versions}; do
      docker_build "load" "${FUNCTION_TYPE}" "ts" "${CURRENT_FUNCTION}" "${version}"
    done
    ;;
  push)
    for version in ${versions}; do
      docker_build "push" "${FUNCTION_TYPE}" "ts" "${CURRENT_FUNCTION}" "${version}"
    done
    ;;
  *)
    echo "Usage: $0 {build|push}"
    exit 1
esac
