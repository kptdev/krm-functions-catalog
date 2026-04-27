#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:latest --image-pull-policy=never -- \
by-value=project-id by-file-path='**/setters.yaml' put-value=new-project
