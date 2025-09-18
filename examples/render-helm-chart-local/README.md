# render-helm-chart: Local Chart

### Overview

This example demonstrates how to imperatively invoke the `render-helm-chart`
function to render a helm chart that lives in your local filesystem.

### Function invocation

Run the following command to fetch the example package:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/render-helm-chart-local
```

```shell
$ cd render-helm-chart-local
```

Run the following commands to render the helm chart in your local
filesystem.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--mount type=bind,src=$(pwd),dst=/tmp/charts \
-- name=helloworld-chart \
releaseName=test
```

You can optionally provide your own values files using `valuesFile`.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--mount type=bind,src=$(pwd),dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=/tmp/charts/helloworld-values/values.yaml
```

You can optionally skip tests in the templated output with `skipTests`.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest \
--mount type=bind,src=$(pwd),dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=/tmp/charts/helloworld-values/values.yaml \
skipTests=true
```

### Expected result

You can run the following command to see the new files you have:

```shell
$ kpt pkg tree
├── [deployment_test-helloworld-chart.yaml]  Deployment test-helloworld-chart
├── [pod_test-helloworld-chart-test-connection.yaml]  Pod test-helloworld-chart-test-connection
├── [service_test-helloworld-chart.yaml]  Service test-helloworld-chart
└── [serviceaccount_test-helloworld-chart.yaml]  ServiceAccount test-helloworld-chart
```

You should be able to find `replicas: 5` in
file `deployment_test-helloworld-chart.yaml`, which demonstrates that
the correct values file provided by --valuesFile was used.

If you provided the `skipTests` option, `pod_test-helloworld-chart-test-connection.yaml`
will not appear in your files. 
