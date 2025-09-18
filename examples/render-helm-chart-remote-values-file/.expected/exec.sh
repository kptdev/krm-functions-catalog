#!/usr/bin/env bash

kpt fn eval --image-pull-policy never --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--network \
--mount type=bind,src="$(pwd)",dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=https://raw.githubusercontent.com/kptdev/krm-functions-catalog/42021718ecffe068c44e774746d75ee4870c96c6/examples/inflate-helm-chart-local/helloworld-values/values.yaml
