#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:latest --image-pull-policy never -- \
by-path='data.**' by-value-regex='(.*)nginx.com(.*)' put-comment='kpt-set: ${1}${host}${2}'
