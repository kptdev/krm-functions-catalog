---
parent_function: "kubeconform"
---
# kubeconform: Imperative Example

### Overview

This example demonstrates how to imperatively invoke [`kubeconform`] function to
validate KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/kubeconform-imperative
```

We have a `ReplicationController` in `app.yaml` that has 2 schema violations:
- `.spec.templates` is unknown, since it should be `.spec.template`.
- `spec.replicas` must not be a string.

### Function invocation

Try it out by running the following command:

```shell
# We set `strict=true` to disallow unknown field and `skip_kinds=MyCustom,MyOtherCustom` to skip 2 kinds that we don't have schemas.
$ kpt fn eval kubeconform-imperative --image ghcr.io/kptdev/krm-functions-catalog/kubeconform:latest --results-dir /tmp -- strict=true skip_kinds=MyCustom,MyOtherCustom
```

The key-value pair(s) provided after `--` will be converted to `ConfigMap` by
kpt and used as the function configuration.

### Expected Results

Let's look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: ghcr.io/kptdev/krm-functions-catalog/kubeconform:latest
    stderr: 'failed to evaluate function: error: function failure'
    exitCode: 1
    results:
      - message: got string, want null or integer
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec.replicas
        file:
          path: app.yaml
      - message: additional properties 'templates' not allowed
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec
        file:
          path: app.yaml
```

To fix them:

- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed.

[`kubeconform`]:https://github.com/yannh/kubeconform
