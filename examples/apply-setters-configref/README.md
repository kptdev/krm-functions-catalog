---
parent_function: "apply-setters"
---
# apply-setters: ConfigRef Example

### Overview

This example demonstrates how to declaratively run [`apply-setters`] function
using `configRef` to reference a `ConfigMap` resource in the package.
Unlike `configPath` which references a file path, `configRef` identifies the
configuration resource by its API identity (apiVersion, kind, name).

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/main/examples/apply-setters-configref
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:latest
      configRef:
        apiVersion: v1
        kind: ConfigMap
        name: setters
```

```yaml
# fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  nginx-replicas: "3"
  tag: 1.16.2
```

The `configRef` field references the `ConfigMap` resource by its `apiVersion`,
`kind`, and `name`. The function resolves the matching resource from the package
and uses it as its configuration. The setter values in the ConfigMap's `data`
field are applied to resources containing `# kpt-set: ${setter-name}` comments.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render apply-setters-configref
```

### Expected result

Check the `replicas` field is set to `3` and the `image` field is set to
`nginx:1.16.2` in the Deployment resource.

[`apply-setters`]: {{< relref "apply-setters/v0.2/" >}}
