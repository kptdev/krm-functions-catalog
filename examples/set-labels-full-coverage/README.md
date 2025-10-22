---
parent_function: "set-labels"
---
# set-labels: Full Coverage Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert all common labels to different built-in resources and CustomResourceDefinition (CRD) resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-labels-full-coverage
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-labels:latest
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
  annotations:
    config.kubernetes.io/local-config: "true"
labels:
  color: orange
  fruit: apple
  app: new
```

The desired labels is provided using `labels` field. We have a CRD with group
`dev.example.com`, version `v1` and kind `MyResource`. 

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-full-coverage
```

### Expected result

Check all resources, the following labels should be upserted to the `labels`, `matchingLabels` or `seletors(labelSelector)` fields: 
`color: orange`, `fruit: apple` and `app: new`.

[`set-labels`]: {{< relref "function-catalog/set-labels/v0.2/" >}}