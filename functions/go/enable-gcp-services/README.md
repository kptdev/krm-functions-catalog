# enable-gcp-services

## Overview

<!--mdtogo:Short-->

The `enable-gcp-services` function generates [GCP project service](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service)
resources from a list of services to enable [GCP APIs](https://cloud.google.com/apis) within a specified project. This allows users to succinctly define all
the services necessary in a single resource and have tighter control over which services are enabled in a specific project.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

`enable-gcp-services` function can be used both declaratively and imperatively.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/enable-gcp-services:latest
```

The `enable-gcp-services` function does the following:

1. Discovers all `ProjectServiceSet` custom resources in a given package and nested packages if any.

1. For each `ProjectServiceSet` CR, it generates [GCP project service](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) resources as specified in the `spec.services` list.
    * Adds all annotations defined for `ProjectServiceSet` CR to each generated resource. This can be used for enabling features like
[disable-on-destroy](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service#custom_resource_definition_properties) for generated services.
    * Sets namespace if any defined for `ProjectServiceSet` CR to each generated resource.
    * Sets projectID if any defined for `ProjectServiceSet` CR to each generated resource.
1. Each generated [GCP project service](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) resource
has a `blueprints.cloud.google.com/ownerReference` annotation. This annotation allows `enable-gcp-services` function to
track generated resources for the declarative management of the generated resources. Any changes made to the generate resources will be overwritten and should be made to the `ProjectServiceSet` CR instead.

### `ProjectServiceSet`

This function only supports local-config custom resources of kind `ProjectServiceSet` and can be provided using input items along with other KRM resources. Multiple `ProjectServiceSet` CRs can be declared in a package.

`ProjectServiceSet` has the following supported parameters:

```yaml
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: my-project-services
  annotations:
    cnrm.cloud.google.com/deletion-policy: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services: # list of services to generate
    - compute.googleapis.com
  projectID: foo
```

| Field        |  Description | Example | Required
| -----------: |  ----------- | ----------- | -----------
`spec.services[]`    | A list of GCP services to enable | [compute.googleapis.com,bigquery.googleapis.com] | yes
`spec.projectID`   | Project ID where the services should be enabled. | my-project-id | no

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Let's start with a `ProjectServiceSet` CR for enabling two services `compute.googleapis.com` and `redis.googleapis.com` in a GCP Project `proj1`.

```yaml
# services-config.yaml
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: proj1-service
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  services:
    - compute.googleapis.com
    - redis.googleapis.com
  projectID: proj1
```

Invoke the function:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/enable-gcp-services:latest
```

Generated resources looks like the following:

```yaml
# service_proj1-service-compute.yaml
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: proj1-service-compute
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/proj1-service'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: proj1
```

```yaml
# service_proj1-service-redis.yaml
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: proj1-service-redis
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/proj1-service'
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: proj1
```

<!--mdtogo-->
