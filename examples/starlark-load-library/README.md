# starlark: Load Library

### Overview

In this example, we are going to demonstrate how to load a library in the 
[`starlark`] function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/starlark-load-library
```

We are going to use the following `Kptfile` and `fn-config.yaml` to configure
the function:

```yaml
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
  name: set-namespace-to-prod
source: |
  load('encoding/json.star', 'json')

  def updateReplicas(resources):
    for resource in resources:
      if resource["kind"] == "Deployment":
        obj = json.decode(resource["metadata"]["annotations"]["last-applied"])
        resource["spec"]["replicas"] = obj["spec"]["replicas"]+1
  updateReplicas(ctx.resource_list["items"])
```

We load the json library by `load('encoding/json.star', 'json')`. Then we invoke
the `json.decode` method to deserialize the content from an annotation.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-load-library
```

### Expected result

Check the `.spec.replicas` field should have been updated to 4.

[`starlark`]: https://catalog.kpt.dev/starlark/v0.1/
