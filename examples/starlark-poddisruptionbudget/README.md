# starlark: PodDisruptionBudget

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with `StarlarkRun` as the `functionConfig`. And we will
show how to access a parameter with complex data structure in
the `functionConfig` and use it in the script.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/starlark-poddisruptionbudget
```

We are going to use the following `Kptfile` and `fn-config.yaml` to configure
the function:

```yaml
# Kptfile
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/starlark:latest
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: pdb-for-depl
params:
  pdb:
    apiVersion: policy/v1beta1
    kind: PodDisruptionBudget
    metadata:
      name: nginx-pdb
    spec:
      selector:
        matchLabels:
          app: nginx
      minAvailable: 1
source: |
  def is_pdb(r):
    return r["apiVersion"] == "policy/v1beta1" and r["kind"] == "PodDisruptionBudget" and r["metadata"]["name"] == "nginx-pdb"
  def ensure_pdb(resources, pdb):
    for resource in resources:
      if is_pdb(resource):
        return
    resources.append(pdb)
  pdb = ctx.resource_list["functionConfig"]["params"]["pdb"]
  ensure_pdb(ctx.resource_list["items"], pdb)
```

The Starlark script lives in the `source` field. This script reads the input
resources from `ctx.resource_list` and the `PodDisruptionBudget` resource
from `ctx.resource_list["functionConfig"]["params"]["pdb"]`. It will ensure
there is a `PodDisruptionBudget` resource for the nginx `Deployment`. If not, it
will create it with the `PodDisruptionBudget` resource provided in `params`. 

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-poddisruptionbudget
```

### Expected result

A new file should have been created, and it should contain a
`PodDisruptionBudget` object.

[`starlark`]: https://catalog.kpt.dev/starlark/v0.1/
