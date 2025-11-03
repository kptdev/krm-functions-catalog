# Copyright 2020 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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
SHELL=/bin/bash
export TAG ?= latest
export DEFAULT_CR ?= ghcr.io/kptdev/krm-functions-catalog
export DEFAULT_CONTRIB_CR ?= ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib


.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test unit-test e2e-test build push

unit-test: ## Run unit tests for Go functions
	cd functions/go && $(MAKE) test
	cd contrib/functions/go && $(MAKE) test

e2e-test: ## Run all e2e tests
	cd tests/e2etest && go test -v -run TestE2E ./...

test: unit-test e2e-test ## Run all unit tests and e2e tests

# find all subdirectories with a go.mod file in them
GO_MOD_DIRS = $(shell find . -name 'go.mod' -not -path './documentation/*' -exec sh -c 'echo "$$(dirname "{}")"' \; )
# NOTE: the above line is complicated for Mac and busybox compatibilty reasons.
# It is meant to be equivalent with this:  find . -name 'go.mod' -printf "'%h' " 

.PHONY: tidy
tidy:
	@for f in $(GO_MOD_DIRS); do (cd $$f; echo "Tidying $$f"; go mod tidy) || exit 1; done

build: ## Build all function images.
	cd functions/go && $(MAKE) build
	cd contrib/functions/go && $(MAKE) build

push-curated: ## Push images to registry. WARN: This operation should only be done in CI environment.
	cd functions/go && $(MAKE) push

push-contrib: ## Push images to registry. WARN: This operation should only be done in CI environment.
	cd contrib/functions/go && $(MAKE) push

update-function-docs: ## Update documentation for a function release branch
	(cd scripts/update_function_docs/ && go build -o update_function_docs .)
	RELEASE_BRANCH=$(RELEASE_BRANCH) ./scripts/update_function_docs/update_function_docs
