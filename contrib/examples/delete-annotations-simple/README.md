# delete-annotations: Simple Example

### Overview

In this example, we will see how to delete annotations on a set of resources in a package/folder

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/contrib/examples/delete-annotations-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render delete-annotations-simple
```

### Expected result

One of the two resources i.e. `ConfigMap` in `resources.yaml` should have been mutated with the annotation `annotation-to-delete` removed from `metadata.annotations` where as there shouldn't be any changes to the second resource i.e. `Namespace` as it didn't have the supplied annotation.
