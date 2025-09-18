# set-project-id: Simple Example

### Overview

This example shows how [`set-project-id`] function works.

Running `set-project-id` function on the example packed will:

1.  Set `project-id` [setter](https://catalog.kpt.dev/apply-setters/v0.1/?id=definitions).
2.  Add `cnrm.cloud.google.com/project-id` annotation on
    [Config Connector resources](https://cloud.google.com/config-connector/docs/reference/overview)
    that don't have it.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/set-project-id-simple
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval set-project-id-simple --image ghcr.io/kptdev/krm-functions-catalog/set-project-id:latest -- 'project-id=foo'
```

### Expected result

1.  File setters.yaml will include `project-id: foo` setter.
2.  In resources.yaml `my-test-project-second-bucket` StorageBucket resource
    will include `cnrm.cloud.google.com/project-id: foo` annotation.

[`set-project-id`]: https://catalog.kpt.dev/set-project-id/v0.1/
