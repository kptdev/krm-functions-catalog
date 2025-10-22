---
parent_function: "set-image"
---
# set-image: Imperative Example

### Overview

This example shows how to set annotations in the `.spec.containers[].image`
field on all resources by running [`set-image`] function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-image-imperative
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval set-annotations-imperative --image ghcr.io/kptdev/krm-functions-catalog/set-image:latest -- name=nginx newName=bitnami/nginx newTag=1.21.4
```

The labels provided in key-value pairs after `--` will be converted to a
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Check the image `nginx` has been replaced with `bitnami/nginx:1.21.4` for all resources.

[`set-image`]: {{< relref "function-catalog/set-image/v0.1/" >}}
