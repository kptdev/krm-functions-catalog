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

# This Dockerfile is used for building the base image for testing, building and
# releasing functions in the CD pipeline. The base image has docker toolchain,
# golang toolchain and nodejs toolchain

FROM docker

# Install gcc, bash, make and curl
RUN apk add --no-cache build-base bash make curl

# Install buildx
COPY --from=docker/buildx-bin:latest /buildx /usr/libexec/docker/cli-plugins/docker-buildx

# Install golang
COPY --from=golang:1.24.10-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# Install node and npm
RUN apk add --no-cache --update nodejs npm
