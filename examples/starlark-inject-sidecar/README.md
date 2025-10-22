---
parent_function: "starlark"
---
# starlark: Inject Sidecar

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with an inline starlark script as function configuration
to inject sidecar container to `Deployment`.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/starlark-inject-sidecar
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
  name: inject-sidecar-to-depl
source: |
  def ensure_inject_sidecar_to_depl(r):
    if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
      containers = r["spec"]["template"]["spec"]["containers"]
      for container in containers:
        if container["name"] == "logging-agent":
          return
      sidecar = {
        "name": "logging-agent",
        "image": "k8s.gcr.io/fluentd-gcp:1.30",
      }
      containers.append(sidecar)
  def ensure_inject_sidecar(resources):
    for resource in resources:
      ensure_inject_sidecar_to_depl(resource)
  ensure_inject_sidecar(ctx.resource_list["items"])
```

The Starlark script is embedded in the `source` field. This script reads the
input KRM resources from `ctx.resource_list` and inject a logging agent sidecar
container in the `Deployment`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-inject-sidecar
```

### Expected result

The logging agent container should have been injected. 

[`starlark`]: {{< relref "function-catalog/starlark/v0.5/" >}}
