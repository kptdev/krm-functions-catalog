# remove-local-config-resources: Imperative Example

### Overview

In this example, we will see how locally annotated resources are pruned from the
supplied resource list.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/remove-local-config-resources-imperative
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn eval remove-local-config-resources-imperative -i ghcr.io/kptdev/krm-functions-catalog/remove-local-config-resources:latest
```

### Expected result

The following resource should have been pruned from resouces.yaml since it had the
annotation: `config.kubernetes.io/local-config: "true"`

```yaml
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  name: sample-hierarchy
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  config:
    - simple
  parentRef:
    external: "123456789012"
```

The resources.yaml file shoud look like this

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-applied
data:
  foo: bar
```
