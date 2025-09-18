# GCP project to use for development
GOBIN := $(shell go env GOPATH)/bin

export GITHUB_PROJECT_ID ?= $(shell gcloud config get-value project)
export IMAGE_REPO ?= ghcr.io/$(GITHUB_PROJECT_ID)
export IMAGE_TAG ?= latest

build:
	KO_DOCKER_REPO=ko.local go run github.com/google/ko@v0.18.0 build -B --tags=${IMAGE_TAG} .

push:
	KO_DOCKER_REPO=${IMAGE_REPO} go run github.com/google/ko@v0.18.0 build -B --tags=${IMAGE_TAG} .
