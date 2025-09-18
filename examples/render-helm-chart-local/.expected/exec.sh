#!/usr/bin/env bash

kpt fn eval --image-pull-policy never --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--mount type=bind,src="$(pwd)",dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=/tmp/charts/helloworld-values/values.yaml \
skipTests=true