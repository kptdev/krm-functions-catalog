---
parent_function: "set-namespace"
---
# set-namespace: Simple Example

### Overview

This example demonstrates how to run [`set-namespace`] function
to replace the  `namespace` resource type in a variety of KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-namespace-simple
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-namespace:latest
      configMap:
        namespace: example-ns
```

The function configuration is provided using a `ConfigMap`. We set only one
key-value pair:
- `namespace: example-ns`: The desired namespace.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-simple
```

### Expected result

Check all resources have `metadata.namespace` set to `example-ns`

[`set-namespace`]: {{< relref "function-catalog/set-namespace/v0.4/" >}}
