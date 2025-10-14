#!/usr/bin/env bash

set -eo pipefail

kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/kubeconform:latest --image-pull-policy never \
  --results-dir="$(pwd)/../results" \
  --mount type=bind,src="$(pwd)/jsonschema",dst=/schema-dir/master-standalone \
  -- schema_location=file:///schema-dir
