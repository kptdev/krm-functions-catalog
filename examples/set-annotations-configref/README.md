---
parent_function: "set-annotations"
---
# set-annotations: ConfigRef Example

### Overview

This example demonstrates how to declaratively run [`set-annotations`] function
using `configRef` to reference a `SetAnnotations` fn-config resource in the package.
Unlike `configPath` which references a file path, `configRef` identifies the
configuration resource by its API identity (apiVersion, kind, name).

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/main/examples/set-annotations-configref
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-annotations:latest
      configRef:
        apiVersion: fn.kpt.dev/v1alpha1
        kind: SetAnnotations
        name: my-func-config
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetAnnotations
metadata:
  name: my-func-config
  annotations:
    config.kubernetes.io/local-config: "true"
annotations:
  color: orange
  fruit: apple
```

The `configRef` field references the `SetAnnotations` resource by its `apiVersion`,
`kind`, and `name`. The function resolves the matching resource from the package
and uses it as its configuration.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-annotations-configref
```

### Expected result

Check all resources have 2 annotations `color: orange` and `fruit: apple`.
Note that when using `configRef`, the `set-annotations` function also annotates
the Kptfile and fn-config resource metadata, since the referenced config resource
remains part of the package input.

[`set-annotations`]: {{< relref "set-annotations/v0.1/" >}}
