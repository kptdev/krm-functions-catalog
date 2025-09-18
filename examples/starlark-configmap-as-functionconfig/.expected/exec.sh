#! /bin/bash

kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/starlark:latest --image-pull-policy never -- source="$(cat set-replicas.star)" replicas=5
