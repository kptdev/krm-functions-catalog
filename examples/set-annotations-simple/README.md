---
parent_function: "set-annotations"
---
# set-annotations: Simple Example

### Overview

This example demonstrates how to declaratively run [`set-annotations`] function
to upsert annotations to the `.metadata.annotations` field on all resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-annotations-simple
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-annotations:latest
      configMap:
        color: orange
        fruit: apple
```

The desired annotations are provided as key-value pairs through `ConfigMap`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-annotations-simple
```

### Expected result

Check the 2 annotations have been added.

[`set-annotations`]: {{< relref "function-catalog/set-annotations/v0.1/" >}}
