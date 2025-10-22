---
parent_function: "create-setters"
---
# create-setters: Simple Example

### Overview

In this example, we will see how to add [setter] comments to
resource fields using `create-setters` function.

### Fetch the example package

Get the example package by running the following command:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/create-setters-simple
```

We use `ConfigMap` to configure the `create-setters` function.
The desired setter values are provided as key-value pairs using `data` field.
Here, key is the name of the setter and value is the field value to be parameterized.

```yaml
# create-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  nginx-replicas: "4"
  env: |
    - dev
    - stage
  tag: 1.16.1
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render create-setters-simple
```

### Expected result

1. Check the setter comment `kpt-set: ${nginx-replicas}` is added to `replicas` field value `4` in `Deployment` resource.
2. Check the setter comment `kpt-set: nginx:${tag}` is added to `image` field value `nginx:1.16.1` in `Deployment` resource.
3. Check the setter comment `kpt-set: ${env}` is added to `environment` field in `MyKind` resource.

[setter]: {{< relref "function-catalog/apply-setters/v0.2/#definitions" >}}