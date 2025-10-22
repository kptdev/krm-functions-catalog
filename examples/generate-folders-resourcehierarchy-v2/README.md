---
parent_function: "generate-folders"
---
# generate-folders: ResourceHierarchy V2

### Overview

This example demonstrates how to declaratively run the [`generate-folders`]
function to transform the `ResourceHierarchy` custom resource with
apiVersion `cft.dev/v1alpha2` into `Folder` custom resources.

**Note**: New users should use the latest `ResourceHierarchy` with
apiVersion `blueprints.cloud.google.com/v1alpha3`

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/generate-folders-resourcehierarchy-v2
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render generate-folders-resourcehierarchy-v2
```

### Expected result

You will find 8 additional files whose names start with `folder_` after running
the command above. You can run the following command to see what you have in the
package:

```shell
$ kpt pkg tree generate-folders-resourcehierarchy-v2
Package "generate-folders-resourcehierarchy-v2"
├── [Kptfile]  Kptfile example
├── [folder_dev.team-2.yaml]  Folder dev.team-2
├── [folder_dev.team-one.yaml]  Folder dev.team-one
├── [folder_dev.yaml]  Folder dev
├── [folder_foo.bar.yaml]  Folder foo.bar
├── [folder_foo.yaml]  Folder foo
├── [folder_prod.team-2.yaml]  Folder prod.team-2
├── [folder_prod.team-one.yaml]  Folder prod.team-one
├── [folder_prod.yaml]  Folder prod
└── [resource-hierarchy.yaml]  ResourceHierarchy test-simple
```
