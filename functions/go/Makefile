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
GOPATH := $(shell go env GOPATH)
TAG ?= unstable
DEFAULT_CR ?= ghcr.io/kptdev/krm-functions-catalog
GOBIN := $(shell go env GOPATH)/bin

# Edit this list to contain all go functions
FUNCTIONS := \
	apply-replacements \
	apply-setters \
	create-setters \
	enable-gcp-services \
	ensure-name-substring \
	export-terraform \
	fix \
	format \
	gatekeeper \
	generate-kpt-pkg-docs \
	remove-local-config-resources \
	render-helm-chart \
	list-setters \
	search-replace \
	set-annotations \
	set-enforcement-action \
	set-image \
	set-labels \
	set-namespace \
	set-project-id \
	source-gcloud-config \
	starlark \
	upsert-resource \
	set-gcp-resource-ids \
	set-name-prefix

# Targets for running all function tests
FUNCTION_TESTS := $(patsubst %,%-TEST,$(FUNCTIONS))
# Targets for generating all functions docs
FUNCTION_GENERATE_DOCS := $(patsubst %,%-GENERATE,$(FUNCTIONS))

FUNCTION_CHECKLICENSES := $(patsubst %,%-CHECKLICENSES,$(FUNCTIONS))
# Targets to build functions
FUNCTION_BUILDS := $(patsubst %,%-BUILD,$(FUNCTIONS))
# Targets to push images
FUNCTION_PUSH := $(patsubst %,%-PUSH,$(FUNCTIONS))
# Current function name used by individual function targets
CURRENT_FUNCTION ?= unknown

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: verify
verify: generate $(FUNCTIONS) ## Verify all functions

.PHONY: $(FUNCTIONS)
$(FUNCTIONS):
	$(MAKE) CURRENT_FUNCTION=$@ func-verify

.PHONY: test
test: generate $(FUNCTION_TESTS) ## Run unit tests for all functions

.PHONY: $(FUNCTION_TESTS)
$(FUNCTION_TESTS):
	$(MAKE) CURRENT_FUNCTION=$(subst -TEST,,$@) func-test

generate: install-mdtogo generate-docs

.PHONY: generate-docs
generate-docs: $(FUNCTION_GENERATE_DOCS) ## Generate docs for all functions

.PHONY: $(FUNCTION_GENERATE_DOCS)
$(FUNCTION_GENERATE_DOCS):
	$(MAKE) CURRENT_FUNCTION=$(subst -GENERATE,,$@) docs-generate

.PHONY: build
build: generate $(FUNCTION_BUILDS) ## Build all function images. 'TAG' env is used to specify tag. 'dev' will be used if not set.

.PHONY: $(FUNCTION_BUILDS)
$(FUNCTION_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -BUILD,,$@) TAG=$(TAG) DEFAULT_CR=$(DEFAULT_CR) func-build

.PHONY: push
push: $(FUNCTION_PUSH) ## Push images to registry. WARN: This operation should only be done in CI environment.

.PHONY: $(FUNCTION_PUSH)
$(FUNCTION_PUSH):
	$(MAKE) CURRENT_FUNCTION=$(subst -PUSH,,$@) TAG=$(TAG) DEFAULT_CR=$(DEFAULT_CR) func-push

.PHONY: check-licenses
check-licenses: install-go-licenses $(FUNCTION_CHECKLICENSES)

add-licenses:
	@echo Update license in path: functions/go
	GOBIN=$(GOBIN) ../../hack/update-license.sh

.PHONY: $(FUNCTION_CHECKLICENSES)
$(FUNCTION_CHECKLICENSES):
	$(MAKE) CURRENT_FUNCTION=$(subst -CHECKLICENSES,,$@) func-check-licenses

install-go-licenses:
ifeq (, $(shell which go-licenses))
	../../scripts/install-go-licenses.sh
GO_LICENSES=$(GOPATH)/bin/go-licenses
else
GO_LICENSES=$(shell which go-licenses)
endif

install-mdtogo:
	go install github.com/kptdev/kpt/mdtogo@main

# Recipes for individual function
.PHONY: func-fix func-vet func-fmt func-test func-lint \
	func-build func-push func-verify func-check-licenses docs-generate add-licenses
func-verify: docs-generate func-fix func-vet func-fmt func-test func-lint func-check-licenses

func-fix:
	cd $(CURRENT_FUNCTION) && go fix ./...

func-fmt:
	cd $(CURRENT_FUNCTION) && go fmt ./...

func-lint:
	(which $(GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.64.8/install.sh | sh -s -- -b $(GOPATH)/bin v1.64.8)
	cd $(CURRENT_FUNCTION) && time $(GOPATH)/bin/golangci-lint run --timeout=10m ./...

func-test:
	cd $(CURRENT_FUNCTION) && go test -cover ./...

func-vet:
	cd $(CURRENT_FUNCTION) && go vet ./...

func-check-licenses: install-go-licenses
	@echo Checking licenses for $(CURRENT_FUNCTION)...
	cd $(CURRENT_FUNCTION) && $(GO_LICENSES) check .

func-build: func-verify
	@echo Building image for $(CURRENT_FUNCTION)...
	../../scripts/go-function-release.sh build curated

func-push:
	@echo "Checking if $(CURRENT_FUNCTION) is a valid listed function..."
	@if echo "$(FUNCTIONS)" | grep -qw "$(CURRENT_FUNCTION)"; then \
		echo "Pushing image $(CURRENT_FUNCTION)..."; \
		../../scripts/go-function-release.sh push curated; \
	else \
		echo "Error: $(CURRENT_FUNCTION) is not a valid listed function!"; \
		exit 1; \
	fi

docs-generate:
	$(GOBIN)/mdtogo $(CURRENT_FUNCTION) $(CURRENT_FUNCTION)/generated --license=none --strategy=cmdDocs

