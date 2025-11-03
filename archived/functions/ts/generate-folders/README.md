# generate-folders

## Overview

This function transforms the `ResourceHierarchy` custom resource into `Folder`
custom resources constituting the hierarchy. Post-translation, it's necessary to
use the `kpt-folder-parent` function from this repo to translate the results
into Cork configs.

## Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

The input `ResourceHierarchy` custom resources must be passed in
using `input items` instead of using `input functionConfig`.

This function can be used both declaratively and imperatively.

The function is compliant with the [KRM function spec]. It means the input must
be wrapped in a `ResourceList` before passing to the function.

```yaml
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
  - apiVersion: blueprints.cloud.google.com/v1alpha3
    kind: ResourceHierarchy
    metadata:
      name: test-hierarchy
    ...
```

### ResourceHierarchy

The function supports the following versions of the `ResourceHierarchy`
resource:

- blueprints.cloud.google.com/v1alpha3
- cft.dev/v1alpha2
- cft.dev/v1alpha1

`blueprints.cloud.google.com/v1alpha3` is latest and recommended version. New
users should use the latest version.

The `config` array within `spec` represents the desired folder hierarchy. Each
item represents a top level folder. Nested folders can be created within the
config.

For example, if you have the following `ResourceHierarchy`:

```yaml
spec:
  layers:
    - layer_one
    - layer_two
  config:
    - vegetables:
        - carrot
        - tomato
    - fruits:
        - apple
        - banana
```

This will produce the following folders:

- `vegetables`, `vegetables.carrot`, `vegetables.tomato`
- `fruits`, `fruits.apple`, `fruits.banana`

A `folder-ref` annotation will automatically be created for all but the root
folders. For example, `fruits.apple` points to `fruits`:

```yaml
- apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
  kind: Folder
  metadata:
    name: fruits.apple
    annotations:
      cnrm.cloud.google.com/folder-ref: fruits
    namespace: hierarchy
  spec:
    displayName: apple
```

For more information about _why_ `folder-ref` is needed, see the usage doc of
KRM function `functions/kpt-folder-parent`.

[KRM function spec]: https://kpt.dev/book/05-developing-functions/#functions-specification
