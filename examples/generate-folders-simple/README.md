---
parent_function: "generate-folders"
---
# generate-folders: Simple Example

### Overview

This example shows how to use the `generate-folders` function to generate
Config Connector `Folder` resources from a `ResourceHierarchy`.

### Fetch the example package

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/generate-folders-simple
```

### Function invocation

```shell
kpt fn render generate-folders-simple
```

### Expected result

Rendering the package appends generated `Folder` resources to `resources.yaml`.
