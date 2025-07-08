# apply-setters: Simple Example

### Overview

In this example, we will see how to apply desired setter values to the 
resource fields parameterized by `kpt-set` comments.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/apply-setters-simple
```

We use `ConfigMap` to configure the `apply-setters` function. The desired
setter values are provided as key-value pairs using `data` field where key is
the name of the setter(as seen in the reference comments) and value is the new
desired value for the tagged field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  env: |-
    - prod
    - dev
  nested-env: |-
    - key: some-other-key
      value: some-other-value
  nginx-replicas: "3"
  tag: 1.16.2
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render apply-setters-simple
```

### Expected result

1. Check the value of field `replicas` is set to `3` in `Deployment` resource.
2. Check the value of field `image` is set to value `nginx:1.16.2` in `Deployment` resource.
3. Check the value of field `environments` is set to value `[dev, prod]` in `MyKind` resource with the name of `foo`.
4. Check the value of field `environments` is set to value:
```
- key: some-other-key
  value: some-other-value
```
in `MyKind` resource with the name of `bar`.

#### Note:

Refer to the `create-setters` function documentation for information about creating setters.
