---
parent_function: "set-image"
---
# set-image: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-image`] function
to set the image for the `.spec.containers[].image` field on all resources.

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-image:latest
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetImage
metadata:
  name: my-func-config
image:
  name: nginx
  newName: bitnami/nginx
  newTag: 1.21.4
additionalImageFields:
- kind: MyKind
  create: false
  group: dev.example.com
  path: spec/manifest/images[]/image
  version: v1

```

The desired image is provided using the `image` field. We have a CRD with group
`dev.example.com`, version `v1` and kind `MyKind`. We want the image to be
set for the field `.spec.manifest.images[].image` as well. We specify it in
field `additionalImageFields`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-image-advanced
$ kpt fn render set-image-advanced
```

### Expected result

Check that the image `nginx` has been set to `bitnami/nginx:1.21.4` in the
standard `.spec.containers[].image` field of the resource of kind `Pod`. And the
image `nginx` has been set to `bitnami/nginx:1.21.4` in the custom
`.spec.manifest.images[].image` location of the resource of kind `MyKind`

[`set-image`]: {{< relref "function-catalog/set-image/v0.1/" >}}
