# export-terraform: Imperative Example

### Overview

In this example, we will see how to export Terraform configuration from KCC resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/export-terraform-imperative
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval export-terraform-imperative --image ghcr.io/kptdev/krm-functions-catalog/export-terraform:latest
```

### Expected result
The function should export successfully
```shell
[RUNNING] "ghcr.io/kptdev/krm-functions-catalog/export-terraform:latest"
[PASS] "ghcr.io/kptdev/krm-functions-catalog/export-terraform:latest" in 1.5s
```

A `ConfigMap` will be placed in `terraform.yaml` which contains the converted Terraform code.
