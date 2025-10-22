---
parent_function: "render-helm-chart"
---
# render-helm-chart: Remote Chart with CRDs

### Overview

This example demonstrates how to imperatively invoke the `render-helm-chart`
function to render a helm chart that contains CRDs in the templated output.

### Function invocation

First, let's render a terraform chart without CRDs:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest --network -- \
name=terraform \
repo=https://helm.releases.hashicorp.com \
version=1.0.0 \
releaseName=terraforming-mars 
```

You should see rendered chart resources in your local filesystem. You can run the
following command to see what you have:

```shell
$ kpt pkg tree
├── [configmap_terraforming-mars-terraform-test.yaml]  ConfigMap terraforming-mars-terraform-test
├── [pod_terraforming-mars-terraform-test.yaml]  Pod terraforming-mars-terraform-test
├── [role_terraforming-mars-terraform-sync-workspace.yaml]  Role terraforming-mars-terraform-sync-workspace
├── [rolebinding_terraforming-mars-terraform-sync-workspace.yaml]  RoleBinding terraforming-mars-terraform-sync-workspace
├── [serviceaccount_terraforming-mars-terraform-sync-workspace.yaml]  ServiceAccount terraforming-mars-terraform-sync-workspace
├── [workspace_terraforming-mars-terraform-test.yaml]  Workspace terraforming-mars-terraform-test
└── default
    └── [deployment_terraforming-mars-terraform-sync-workspace.yaml]  Deployment default/terraforming-mars-terraform-sync-workspace
```

Now, let's run the command again, this time including CRDs:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest --network -- \
name=terraform \
repo=https://helm.releases.hashicorp.com \
version=1.0.0 \
releaseName=terraforming-mars \
includeCRDs=true
```

### Expected result

Run the following command to see what you have:

```shell
$ kpt pkg tree
├── [configmap_terraforming-mars-terraform-test.yaml]  ConfigMap terraforming-mars-terraform-test
├── [customresourcedefinition_workspaces.app.terraform.io.yaml]  CustomResourceDefinition workspaces.app.terraform.io
├── [pod_terraforming-mars-terraform-test.yaml]  Pod terraforming-mars-terraform-test
├── [role_terraforming-mars-terraform-sync-workspace.yaml]  Role terraforming-mars-terraform-sync-workspace
├── [rolebinding_terraforming-mars-terraform-sync-workspace.yaml]  RoleBinding terraforming-mars-terraform-sync-workspace
├── [serviceaccount_terraforming-mars-terraform-sync-workspace.yaml]  ServiceAccount terraforming-mars-terraform-sync-workspace
├── [workspace_terraforming-mars-terraform-test.yaml]  Workspace terraforming-mars-terraform-test
└── default
    └── [deployment_terraforming-mars-terraform-sync-workspace.yaml]  Deployment default/terraforming-mars-terraform-sync-workspace
```

Notice that you now have a new file `customresourcedefinition_workspaces.app.terraform.io.yaml` that wasn't there before,
containing the CRDs you included.
