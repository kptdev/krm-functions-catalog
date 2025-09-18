# set-annotations: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-annotations`] function
to upsert annotations to the `.metadata.annotations` field on all resources.

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-annotations:latest
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetAnnotations
metadata:
  name: my-func-config
annotations:
  fruit: apple
  color: orange
additionalAnnotationFields:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/annotations
```

The desired annotations are provided using the `annotations` field. We have a
CRD with group `dev.example.com`, version `v1` and kind `MyResource`. We want
the annotations to be added to field `.spec.selector.annotations` as well. We
specify it in field `additionalAnnotationFields`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/set-annotations-advanced
$ kpt fn render set-annotations-advanced
```

### Expected result

Check all resources have 2 annotations: `color: orange` and `fruit: apple`. And
the resource of kind `MyResource` also has these 2 annotations in
`spec.selector.annotations`.

[`set-annotations`]: https://catalog.kpt.dev/set-annotations/v0.1/
