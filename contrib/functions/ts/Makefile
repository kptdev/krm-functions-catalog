# Copyright 2020 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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
SHELL=/bin/bash
TAG ?= unstable
CR = ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib
GOBIN := $(shell go env GOPATH)/bin

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test build npm-ci push

# Edit this list to contain all ts functions
FUNCTIONS := \
	analyze-istio \
	sops

# Targets for running all function tests
FUNCTION_TESTS := $(patsubst %,%-TEST,$(FUNCTIONS))

# Targets to build functions
FUNCTION_BUILDS := $(patsubst %,%-BUILD,$(FUNCTIONS))
# Targets to push function images
FUNCTION_PUSH := $(patsubst %,%-PUSH,$(FUNCTIONS))
# Targets to run npm ci
FUNCTION_NPM_CI := $(patsubst %,%-NPM-CI,$(FUNCTIONS))
# Current function name used by individual function targets
CURRENT_FUNCTION ?= unknown

npm-ci: $(FUNCTION_NPM_CI) ## Run npm ci command

.PHONY: $(FUNCTION_NPM_CI)
$(FUNCTION_NPM_CI):
	$(MAKE) CURRENT_FUNCTION=$(subst -NPM-CI,,$@) func-npm-ci

test: $(FUNCTION_TESTS) ## Run unit tests for all functions

.PHONY: $(FUNCTION_TESTS)
$(FUNCTION_TESTS):
	$(MAKE) CURRENT_FUNCTION=$(subst -TEST,,$@) func-test

build: $(FUNCTION_BUILDS) ## Build all function images. Variable 'TAG' is used to specify tag. 'dev' will be used if not set.

.PHONY: $(FUNCTION_BUILDS)
$(FUNCTION_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -BUILD,,$@) TAG=$(TAG) DEFAULT_CR=$(CR) func-build

push: $(FUNCTION_PUSH) ## Push images to registry. WARN: This operation should only be done in CI environment.

.PHONY: $(FUNCTION_PUSH)
$(FUNCTION_PUSH):
	$(MAKE) CURRENT_FUNCTION=$(subst -PUSH,,$@) TAG=$(TAG) DEFAULT_CR=$(CR) func-push

# Recipes for individual function
.PHONY: func-test func-build func-push func-npm-ci
func-npm-ci:
	cd $(CURRENT_FUNCTION) && npm ci --ignore-scripts

func-test: func-npm-ci
	cd $(CURRENT_FUNCTION) && npm test

func-build: func-test
	../../../scripts/ts-function-release.sh build contrib

func-push:
	../../../scripts/ts-function-release.sh push contrib
