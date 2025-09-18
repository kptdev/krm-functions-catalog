# ensure-name-substring: Advanced Example

### Overview

This example demonstrates how to declaratively run the [`ensure-name-substring`]
function to prepend prefix in the resource names.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/ensure-name-substring-advanced
```

We use the following `Kptfile` and `fn-config.yaml` to run the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/ensure-name-substring:latest
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: EnsureNameSubstring
metadata:
  name: my-config
substring: prod-
editMode: prepend
additionalNameFields:
  - group: dev.example.com
    version: v1
    kind: MyResource
    path: spec/name
```

We are going to prepend prefix `prod-` to resource names.
The function will not only update field `.metadata.name` but also field
`.spec.name` in `MyResource`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render ensure-name-substring-advanced
```

### Expected result

Check all resources have `prod-` in their names and the field `.spec.name` in
`MyResource` also got updated.

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/
