#! /bin/bash
#
# Copyright 2021 Google LLC
# Modifications Copyright (C) 2025-2026 OpenInfra Foundation Europe.
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

repo_base=$(cd "$(dirname "$(dirname "$0")")" || exit ; pwd)

CR_REGISTRY=${DEFAULT_CR:-ghcr.io/kptdev/krm-functions-catalog}

function err {
  echo "$1"
  exit 1
}

function docker_build {
  action=$1 # docker buildx operation, it should be either load or push.
  name=$2 # function name, e.g. apply-setters
  shift 2
  local tags=("$@") # one or more tags, e.g. v1.2.3 v1.2 v1

  build_args=()

  function_dir="${repo_base}/functions/go/${name}"

  override_dockerfile="${function_dir}"/Dockerfile

  dockerfile="${repo_base}/build/docker/go/Dockerfile"
  [[ -f "${override_dockerfile}" ]] && dockerfile="${override_dockerfile}"
  [[ -f "${dockerfile}" ]] || err "Dockerfile does not exist: ${dockerfile}"

  echo "Using Dockerfile ${dockerfile}" 

  defaults="${repo_base}/build/docker/go/defaults.env"
  [[ -f "${defaults}" ]] || err "defaults file does not exist: ${defaults}"
  # shellcheck source=/dev/null
  source "${defaults}"
  build_args+=(--build-arg "BUILDER_IMAGE=${BUILDER_IMAGE}")
  build_args+=(--build-arg "BASE_IMAGE=${BASE_IMAGE}")

  if [[ ! -f "${function_dir}/go.mod" ]]; then
    function_dir="${repo_base}/functions/go/"
    echo "Setting build context to ${function_dir}"
  fi

  # Build tag arguments
  local tag_args=()
  for tag in "${tags[@]}"; do
    tag_args+=(-t "${CR_REGISTRY}/${name}:${tag}")
  done

  echo "building ${CR_REGISTRY}/${name} with tags: ${tags[*]}"

  case "${action}" in
    load)
      IFS=' ' read -r -a extra_args_array <<< "${EXTRA_BUILD_ARGS:-}"

      # Use + conditional parameter expansion to protect from unbound array variable
      docker buildx build --load \
        "${tag_args[@]}" \
        -f "${dockerfile}" \
        "${build_args[@]+"${build_args[@]}"}" \
        "${extra_args_array[@]+"${extra_args_array[@]}"}" \
        "${function_dir}"    
      ;;
    push)
      IFS=' ' read -r -a extra_args_array <<< "${EXTRA_BUILD_ARGS:-}"
      # build and push multi-arch image.
      docker buildx build --push \
        "${tag_args[@]}" \
        -f "${dockerfile}" \
        --platform "linux/amd64,linux/arm64" \
        "${build_args[@]+"${build_args[@]}"}" \
        "${extra_args_array[@]+"${extra_args_array[@]}"}" \
        "${function_dir}"    
      ;;
    *)
      echo "action must be load or push"
      exit 1
  esac
}
