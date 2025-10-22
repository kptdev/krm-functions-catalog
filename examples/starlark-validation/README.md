---
parent_function: "starlark"
---
# starlark: Validation

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with an inline starlark script as function configuration
to validate a `ConfigMap`.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/starlark-validation
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
  name: no-private-key
source: |
  def contains_private_key(r):
    return r["apiVersion"] == "v1" and r["kind"] == "ConfigMap" and r["data"]["private-key"]
  def ensure_no_private_key(resource_list):
    for resource in resource_list["items"]:
      if contains_private_key(resource):
        fail("it is prohibited to have private key in a configmap")
  ensure_no_private_key(ctx.resource_list)
```

The Starlark script is embedded in the `source` field. This script reads the
input KRM resources from `ctx.resource_list` and validate there are no private
keys in the `ConfigMap`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-validation --results-dir /tmp
```

### Expected result

Let's take a look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: ghcr.io/kptdev/krm-functions-catalog/starlark:latest
    stderr: 'fail: it is prohibited to have private key in a configmap'
    exitCode: 1
    results:
      - message: 'fail: it is prohibited to have private key in a configmap'
        severity: error
```

To pass validation, let's replace the key `private-key` in the `ConfigMap` with
something else e.g. `public_key`.
Rerun the command. It will succeed.

[`starlark`]: {{< relref "function-catalog/starlark/v0.5/" >}}
