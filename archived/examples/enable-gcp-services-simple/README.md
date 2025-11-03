# enable-gcp-services: Simple Example

### Overview

In this example, we will see how to generate [project services](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) for `compute.googleapis.com` and `redis.googleapis.com`.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/archived/examples/enable-gcp-services-simple
```

### Function invocation

Render the configuration, `enable-gcp-services` has been defined declaratively in the Kptfile.

```shell
$ kpt fn render enable-gcp-services-simple
```

### Expected result

```shell
[RUNNING] "ghcr.io/kptdev/krm-functions-catalog/archived/enable-gcp-services:latest"
[PASS] "ghcr.io/kptdev/krm-functions-catalog/archived/enable-gcp-services:latest"
  Results:
    [INFO] generated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-compute"
    [INFO] generated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-redis"

Successfully executed 1 function(s) in 1 package(s).
```

Run the following command to see generated service resources:

```shell
$ kpt pkg tree enable-gcp-services-simple
```

You will see two new generated resources `proj1-service-compute` and `proj1-service-redis`.

```shell
Package "enable-gcp-services-simple"
├── [Kptfile]  Kptfile enable-gcp-services-simple
├── [service_proj1-service-compute.yaml]  Service proj1-service-compute
├── [service_proj1-service-redis.yaml]  Service proj1-service-redis
└── [services.yaml]  ProjectServiceList proj1-service
```
