# GCP project to use for development
GOBIN := $(shell go env GOPATH)/bin

export GITHUB_PROJECT_ID ?= \
	$(if $(shell which gcloud),$(shell gcloud config get-value project), \
	ghcr.io/kptdev/krm-functions-catalog)
export IMAGE_REPO ?= ghcr.io/$(GITHUB_PROJECT_ID)
export IMAGE_TAG ?= latest

build: install-ko
	GOBIN=$(GOBIN) KO_DOCKER_REPO=ko.local ko build -B --tags=${IMAGE_TAG} .

push: install-ko
	GOBIN=$(GOBIN) KO_DOCKER_REPO=${IMAGE_REPO} ko build -B --tags=${IMAGE_TAG} .


install-ko:
	(which ko || go install github.com/google/ko@latest)
