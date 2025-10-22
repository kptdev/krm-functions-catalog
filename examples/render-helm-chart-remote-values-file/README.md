---
parent_function: "render-helm-chart"
---
# render-helm-chart: Remote Values File

### Overview

This example demonstrates how to imperatively invoke the `render-helm-chart`
function with a remote values file.

### Function invocation

Run the following command to fetch the example package:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/render-helm-chart-remote-values-file
```

Run the following commands to render the helm chart in your local
filesystem with the remote values file.

```shell
$ kpt fn eval render-helm-chart-remote-values-file --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--network \
--mount type=bind,src="$(pwd)"/render-helm-chart-remote-values-file,dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=https://raw.githubusercontent.com/kptdev/krm-functions-catalog/42021718ecffe068c44e774746d75ee4870c96c6/examples/inflate-helm-chart-local/helloworld-values/values.yaml
```

### Expected result

You can run the following command to see the new files you have:

```shell
$ kpt pkg tree render-helm-chart-remote-values-file
render-helm-chart-remote-values-file
├── [deployment_test-helloworld-chart.yaml]  Deployment test-helloworld-chart
├── [pod_test-helloworld-chart-test-connection.yaml]  Pod test-helloworld-chart-test-connection
├── [service_test-helloworld-chart.yaml]  Service test-helloworld-chart
└── [serviceaccount_test-helloworld-chart.yaml]  ServiceAccount test-helloworld-chart
```

You should be able to find `replicas: 5` in
file `deployment_test-helloworld-chart.yaml`, which demonstrates that
the correct values file provided by --valuesFile was used.
