# export-terraform

## Overview

<!--mdtogo:Short-->

Generate idiomatic Terraform configuration equivalents for Config Connector resources.

<!--mdtogo-->

This function generates idiomatic Terraform configuration by parsing [Config Connector (KCC)](https://cloud.google.com/config-connector/docs) resources and pushing the equivalent Terraform configuration into a `ConfigMap`.

Where appropriate, the generated Terraform references [Cloud Foundation Toolkit modules](https://g.co/dev/terraformfoundation).
The goal is to make the generated output as close to possible as what a human would have written.

The following KCC resources are supported:
- Folder
- Project
- ComputeSharedVPCHostProject
- IAMPartialPolicy
- IAMPolicy
- IAMPolicyMember
- LoggingLogSink
- BigQueryDataset
- PubSubTopic
- StorageBucket
- IAMAuditConfig
- ComputeNetwork
- ComputeSubnetwork
- ComputeFirewall
- ComputeRoute
- ComputeRouter
- ComputeRouterNAT
- ComputeAddress
- ServiceNetworkingConnection
- LoggingLogBucket

The output Terraform will be saved to a `ConfigMap` in `terraform.yaml` at the root of the package.
Each key in the `ConfigMap` corresponds to a different file which is part of the Terraform module.

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: terraform
  annotations:
    config.kubernetes.io/local-config: "true"
    blueprints.cloud.google.com/syntax: "hcl"
    blueprints.cloud.google.com/flavor: "terraform"
data:
  folders.tf: |+
    resource "google_folder" "test" {
      display_name = "Test Display"
      parent       = "organizations/123456789012"
    }
```

### Skipping Resources
Any resource annotated with `cnrm.cloud.google.com/ignore-clusterless: "true"` will be excluded from the export.

### Attribution
The exported Terraform configuration will include a `provider_meta` block for attributing it back to this function.
If you want to prevent attributing the configuration to this function, you should delete this block.

```
terraform {
  provider_meta "google" {
    module_name = "blueprints/terraform/exported-krm/v0.1.0"
  }
}
```

<!--mdtogo:Long-->

## Usage

The function executes as follows:

1. Searches for supported KCC resources in the package
2. For all KCC resources found, generate the equivalent Terraform
3. Output the Terraform into a `ConfigMap`.

`export-terraform` function can be executed imperatively as follows:

```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/export-terraform:latest
```

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Consider the following package:

```
sample
└─ folder.yaml
```

```yaml
# folder.yaml
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Folder
metadata:
  name: test
  namespace: hierarchy
spec:
  displayName: Test Display
  organizationRef:
    external: '123456789012'
```

Invoke the function in the package directory:

```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/export-terraform:latest
```

The resulting package structure would look like this:

```
sample
|─ folder.yaml
└─ terraform.yaml
```
