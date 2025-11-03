# export-terraform: Advanced Example

### Overview

In this example, we will see how to export Terraform configuration from a complex blueprint with many KCC resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/archived/examples/export-terraform-advanced
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval export-terraform-advanced --image ghcr.io/kptdev/krm-functions-catalog/archived/export-terraform:latest
```

### Expected result
The function should export successfully
```shell
[RUNNING] "ghcr.io/kptdev/krm-functions-catalog/archived/export-terraform:latest"
[PASS] "ghcr.io/kptdev/krm-functions-catalog/archived/export-terraform:latest" in 1.5s
```

A `ConfigMap` will be placed in `terraform.yaml` which contains the converted Terraform code.
