---
parent_function: "set-image"
---
# set-image: Digest Example

### Overview

This example demonstrates how to declaratively run [`set-image`] function
to set the `.spec.containers[].image` field to a specified digest on certain
resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-image-digest
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
  - image: ghcr.io/kptdev/krm-functions-catalog/set-image:latest
    configMap:
      name: nginx
      newName: bitnami/nginx
      digest: sha256:e1d6f00f191515945233f562777fc9ab3b7637ea75169b3dd628d46c9b24400f
```

The desired image specification is provided through `ConfigMap` keys `name`,
`newName`, and `digest`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-image-digest
```

### Expected result

Check the 2 images have been set to the specified digest.

[`set-image`]: {{< relref "function-catalog/set-image/v0.1/" >}}
