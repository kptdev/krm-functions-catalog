---
parent_function: "starlark"
---
# starlark: ConfigMap as functionConfig

### Overview

In this example, we are going to demonstrate how to run the [`starlark`]
function imperatively with a `ConfigMap` as the `functionConfig` to set
namespaces to KRM resources. And the `ConfigMap` is generated from the command
line arguments.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/starlark-configmap-as-functionconfig
```

We are going to use the following starlark script:

```python
# set-replicas.star
def setReplicas(resources, replicas):
    for r in resources:
        if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
            r["spec"]["replicas"] = replicas

replicas = ctx.resource_list["functionConfig"]["data"]["replicas"]
setReplicas(ctx.resource_list["items"], replicas)
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/starlark:latest -- source="$(cat set-replicas.star)" replicas=5
```

### Expected result

Check the `spec.replicas` field has been set to `5` in the `Deployment`.

[`starlark`]: {{< relref "function-catalog/starlark/v0.5/" >}}
