# annotate-apply-time-mutations: Custom Resource Object Example

### Overview

This example shows how the [annotate-apply-time-mutations] function works.

Running `annotate-apply-time-mutations` function on the example package will:

1.  Set `config.k8s.io/apply-time-mutation` annotation on resources generated from an `ApplyTimeMutation` resource object.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/contrib/examples/annotate-apply-time-mutations-custom-resource
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval annotate-apply-time-mutations-custom-resource --image ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib/annotate-apply-time-mutations:latest
```

### Expected result

1.  File resources.yaml will include `config.k8s.io/apply-time-mutation` annotation matching the mutation spec.

[annotate-apply-time-mutations] https://catalog.kpt.dev/annotate-apply-time-mutations/v0.1/?id=definitions
