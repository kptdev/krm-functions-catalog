# render-helm-chart: Remote Chart

### Overview

This example demonstrates how to imperatively invoke the `render-helm-chart`
function to render a helm chart that lives in a remote repo.

### Function invocation

Run the following command to render a minecraft chart.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest --network -- \
name=minecraft \
repo=https://itzg.github.io/minecraft-server-charts \
version=3.1.3 \
releaseName=test
```

### Expected result

You should have several files in your local filesystem. You can run the
following command to see what you have:

```shell
$ kpt pkg tree
├── [secret_test-minecraft.yaml]  Secret test-minecraft
└── [service_test-minecraft.yaml]  Service test-minecraft
```
