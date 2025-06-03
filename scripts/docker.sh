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

repo_base=$(cd "$(dirname "$(dirname "$0")")" || exit ; pwd)

DEFAULT_GCR=${DEFAULT_GCR:-gcr.io/kpt-fn-contrib}
GCR_REGISTRY=${GCR_REGISTRY:-${DEFAULT_GCR}}

function err {
  echo "$1"
  exit 1
}

function docker_build {
  action=$1 # docker buildx operation, it should be either load or push.
  type=$2 # function type, e.g. contrib, curated
  lang=$3 # function language, e.g. go, ts
  name=$4 # function name, e.g. apply-setters
  tag=$5 # function tag, e.g. v1.2.3

  build_args=()

  case "${type}" in
    contrib) function_dir="${repo_base}/contrib/functions/${lang}/${name}" ;;
    curated) function_dir="${repo_base}/functions/${lang}/${name}" ;;
    *) err "unknown function type: ${type}" ;;
  esac

  case "${lang}" in
    ts)
      translated_name=$(echo "${name}" | tr - _)
      build_args+=(--build-arg "FILENAME=${translated_name}_run.js")
      override_dockerfile="${function_dir}/build/${translated_name}.Dockerfile"
      ;;
    *) override_dockerfile="${function_dir}"/Dockerfile ;;
  esac

  dockerfile="${repo_base}/build/docker/${lang}/Dockerfile"
  [[ -f "${override_dockerfile}" ]] && dockerfile="${override_dockerfile}"
  [[ -f "${dockerfile}" ]] || err "Dockerfile does not exist: ${dockerfile}"

  defaults="${repo_base}/build/docker/${lang}/defaults.env"
  [[ -f "${defaults}" ]] || err "defaults file does not exist: ${defaults}"
  # shellcheck source=/dev/null
  source "${defaults}"
  build_args+=(--build-arg "BUILDER_IMAGE=${BUILDER_IMAGE}")
  build_args+=(--build-arg "BASE_IMAGE=${BASE_IMAGE}")

  echo "building ${GCR_REGISTRY}/${name}:${tag}"

  case "${action}" in
    load)
      IFS=' ' read -r -a extra_args_array <<< "${EXTRA_BUILD_ARGS:-}"
      # Use + conditional parameter expansion to protect from unbound array variable
      docker buildx build --load \
        -t "${GCR_REGISTRY}/${name}:${tag}" \
        -f "${dockerfile}" \
        "${build_args[@]+"${build_args[@]}"}" \
        "${extra_args_array[@]}" \
        "${function_dir}"    
      ;;
    push)
      # build and push multi-arch image.
      docker buildx build --push \
        -t "${GCR_REGISTRY}/${name}:${tag}" \
        -f "${dockerfile}" \
        --platform "linux/amd64,linux/arm64" \
        "${build_args[@]+"${build_args[@]}"}" \
        "${function_dir}"    
      ;;
    *)
      echo "action must be load or push"
      exit 1
  esac
}
