---
parent_function: "set-labels"
---
# set-labels: Simple Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert labels to the `.metadata.labels` field on all resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-labels-simple
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-labels:latest
      configMap:
        color: orange
        fruit: apple
```

The desired labels are provided as key-value pairs through `ConfigMap`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-simple
```

### Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

[`set-labels`]: {{< relref "function-catalog/set-labels/v0.2/" >}}
