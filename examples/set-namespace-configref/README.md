---
parent_function: "set-namespace"
---
# set-namespace: ConfigRef Example

### Overview

This example demonstrates how to declaratively run [`set-namespace`] function
using `configRef` to reference a `ConfigMap` resource in the package.
Unlike `configMap` which inlines the key-value pairs, `configRef` points to
an existing resource in the package by its API identity (apiVersion, kind, name).

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/main/examples/set-namespace-configref
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-namespace:latest
      configRef:
        apiVersion: v1
        kind: ConfigMap
        name: ns-config
```

```yaml
# fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ns-config
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  namespace: example-ns
```

The `configRef` field references the `ConfigMap` resource by its `apiVersion`,
`kind`, and `name`. The function resolves the matching resource from the package
and uses it as its configuration.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-configref
```

### Expected result

Check the namespace has been updated from `example` to `example-ns` on all resources.

[`set-namespace`]: {{< relref "set-namespace/v0.4/" >}}
