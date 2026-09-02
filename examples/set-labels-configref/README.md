---
parent_function: "set-labels"
---
# set-labels: ConfigRef Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
using `configRef` to reference a `SetLabels` fn-config resource in the package.
Unlike `configPath` which references a file path, `configRef` identifies the
configuration resource by its API identity (apiVersion, kind, name).

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/main/examples/set-labels-configref
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-labels:latest
      configRef:
        apiVersion: fn.kpt.dev/v1alpha1
        kind: SetLabels
        name: my-func-config
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-func-config
  annotations:
    config.kubernetes.io/local-config: "true"
labels:
  color: orange
  fruit: apple
```

The `configRef` field references the `SetLabels` resource by its `apiVersion`,
`kind`, and `name`. The function resolves the matching resource from the package
and uses it as its configuration.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-configref
```

### Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

[`set-labels`]: {{< relref "set-labels/v0.2/" >}}
