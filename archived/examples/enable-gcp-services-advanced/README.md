# enable-gcp-services: Advanced Example

### Overview

In this example, we will see how to generate [project services](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) for multiple projects and how service pruning is done.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/archived/examples/enable-gcp-services-advanced
```

Explore the package by running:
```shell
$ kpt pkg tree enable-gcp-services-advanced
Package "enable-gcp-services-advanced"
├── [Kptfile]  Kptfile enable-gcp-services-advanced
├── [proj1-services.yaml]  ProjectServiceSet proj1-service
├── [proj2-services.yaml]  ProjectServiceSet proj2-service
├── [resources.yaml]  ComputeNetwork computenetwork-sample
├── [resources.yaml]  RedisInstance redisinstance-sample
├── [service_proj1-service-bigquery.yaml]  Service proj1-service-bigquery
├── [service_proj1-service-compute.yaml]  Service proj1-service-compute
├── [service_proj1-service-redis.yaml]  Service proj1-service-redis
└── [service_proj2-service-redis.yaml]  Service proj2-service-redis
```

We can see two `ProjectServiceSet` resources `proj1-service` and `proj2-service` for managing service enablement in two projects `proj1` and `proj2`. We can also see the services `compute.googleapis.com`, `redis.googleapis.com` managed within `ProjectServiceSet` `proj1-service` resource and `redis.googleapis.com` within `ProjectServiceSet` `proj2-service` resource. Notice that `bigquery.googleapis.com` is no longer present in `proj1-service` `ProjectServiceSet`. We can re render this package to prune that service resource.

### Function invocation

Render the configuration, `enable-gcp-services` has been defined declaratively in the Kptfile.

```shell
$ kpt fn render enable-gcp-services-advanced
```

### Expected result

```shell
Package "enable-gcp-services-advanced": 
[RUNNING] "ghcr.io/kptdev/krm-functions-catalog/archived/enable-gcp-services:latest"
[PASS] "ghcr.io/kptdev/krm-functions-catalog/archived/enable-gcp-services:latest"
  Results:
    [INFO] pruned service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-bigquery" in file "service_proj1-service-bigquery.yaml"
    [INFO] recreated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-compute" in file "service_proj1-service-compute.yaml"
    [INFO] recreated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-redis" in file "service_proj1-service-redis.yaml"
    [INFO] recreated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj2-service-redis" in file "service_proj2-service-redis.yaml"

Successfully executed 1 function(s) in 1 package(s).
```

We can see that the `bigquery` service resource was pruned. Run the following command to see current service resources:

```shell
$ kpt pkg tree enable-gcp-services-advanced
```

We can see `proj1-service-bigquery` no longer exists.

```diff
Package "enable-gcp-services-advanced"
 ├── [Kptfile]  Kptfile enable-gcp-services-advanced
 ├── [proj1-services.yaml]  ProjectServiceSet proj1-service
 ├── [proj2-services.yaml]  ProjectServiceSet proj2-service
 ├── [resources.yaml]  ComputeNetwork computenetwork-sample
 ├── [resources.yaml]  RedisInstance redisinstance-sample
-├── [service_proj1-service-bigquery.yaml]  Service proj1-service-bigquery
 ├── [service_proj1-service-compute.yaml]  Service proj1-service-compute
 ├── [service_proj1-service-redis.yaml]  Service proj1-service-redis
 └── [service_proj2-service-redis.yaml]  Service proj2-service-redis
```
