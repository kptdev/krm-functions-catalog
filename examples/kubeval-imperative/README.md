# kubeval: Imperative Example

### Overview

This example demonstrates how to imperatively invoke [`kubeval`] function to
validate KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/kubeval-imperative
```

We have a `ReplicationController` in `app.yaml` that has 2 schema violations:
- `.spec.templates` is unknown, since it should be `.spec.template`.
- `spec.replicas` must not be a string.

### Function invocation

Try it out by running the following command:

```shell
# We set `strict=true` to disallow unknown field and `skip_kinds=MyCustom,MyOtherCustom` to skip 2 kinds that we don't have schemas.
$ kpt fn eval kubeval-imperative --image ghcr.io/kptdev/krm-functions-catalog/kubeval:latest --results-dir /tmp -- strict=true skip_kinds=MyCustom,MyOtherCustom
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
  - image: ghcr.io/kptdev/krm-functions-catalog/kubeval:latest
    exitCode: 1
    results:
      - message: Additional property templates is not allowed
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: templates
        file:
          path: app.yaml
      - message: 'Invalid type. Expected: [integer,null], given: string'
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec.replicas
        file:
          path: app.yaml
```

To fix them:

- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed.

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/
