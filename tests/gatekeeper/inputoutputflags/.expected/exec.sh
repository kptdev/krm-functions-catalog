#!/usr/bin/env bash

docker run --rm -v "$(pwd)"/resources:/resources ghcr.io/kptdev/krm-functions-catalog/gatekeeper:latest --input resources/resources.yaml --output resources/resources.yaml
