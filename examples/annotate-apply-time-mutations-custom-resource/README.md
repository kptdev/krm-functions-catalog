# annotate-apply-time-mutations: Custom Resource Object Example

### Overview

This example shows how to use the `annotate-apply-time-mutations` function with
an `ApplyTimeMutation` custom resource object.

The scenario: `pod-b` needs the IP address and port of `pod-a` in a
`SERVICE_HOST` environment variable, but `pod-a`'s IP is only assigned after it
is scheduled. An `ApplyTimeMutation` resource specifies two substitutions on the
same target field, and the function converts them into the annotation that
`kpt live apply` reads at apply time.

Running `annotate-apply-time-mutations` function on the example package will:

1. Scan resources for `ApplyTimeMutation` objects.
2. Generate the `config.kubernetes.io/apply-time-mutation` annotation on the
   target resources, which `kpt live apply` reads to perform the substitution.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/annotate-apply-time-mutations-custom-resource
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval annotate-apply-time-mutations-custom-resource --image ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

### Expected result

1. File resources.yaml will include a `config.kubernetes.io/apply-time-mutation`
   annotation on `pod-b` with two substitutions: one for `pod-a`'s IP and one
   for its port.

When `kpt live apply` is subsequently run, it will apply `pod-a` first, wait
for it to be reconciled, then substitute `status.podIP` and the container port
into `pod-b`'s `SERVICE_HOST` environment variable before applying it.

