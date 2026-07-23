---
parent_function: "delete-annotations"
---
# delete-annotations: Imperative Example

### Overview

This example shows how to delete annotations from the `.metadata.annotations` field
on all resources by running `delete-annotations` function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/main/examples/delete-annotations-imperative
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval delete-annotations-imperative --image ghcr.io/kptdev/krm-functions-catalog/delete-annotations:latest -- annotationKeys=annotation-to-delete
```

The annotation keys provided after `--` will be converted to a
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Check that the annotation `annotation-to-delete` has been removed from
all resources where it was present.

