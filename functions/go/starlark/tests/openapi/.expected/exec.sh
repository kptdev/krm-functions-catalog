#! /bin/bash

kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/starlark:latest --fn-config fn-config.yaml --image-pull-policy never
