# Copyright (C) 2025 OpenInfra Foundation Europe
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

ARG BUILDER_IMAGE
ARG BASE_IMAGE

# -------- Build Stage --------
FROM --platform=$BUILDPLATFORM $BUILDER_IMAGE AS build
ENV CGO_ENABLED=0
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /usr/local/bin/function ./

ARG KUBECONFORM_VERSION="v0.7.0"
RUN go install github.com/yannh/kubeconform/cmd/kubeconform@${KUBECONFORM_VERSION}

# -------- Final Stage --------
FROM $BASE_IMAGE

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=build /usr/local/bin/function /usr/local/bin/function
COPY --from=build /go/bin/kubeconform /usr/local/bin/kubeconform

COPY jsonschema/jsonschema-k8s.tar.gz /tmp/jsonschema.tar.gz

RUN mkdir /jsonschema && \
    tar -xzf /tmp/jsonschema.tar.gz -C /jsonschema && \
    rm /tmp/jsonschema.tar.gz && \
    chmod +x /usr/local/bin/function /usr/local/bin/kubeconform && \
    chmod -R a+r /jsonschema && \
    chown -R appuser:appgroup /usr/local/bin /jsonschema && \
    rm -rf /var/cache/apk/* /tmp/* /root/.cache

USER appuser

ENTRYPOINT ["function"]