---
parent_function: "ensure-name-substring"
---
# ensure-name-substring: Depends on example

### Overview

This example demonstrates how the [`ensure-name-substring`] function interacts
with resources that use the [`depends-on`] annotation. The [`depends-on`]
annotation is a special annotation which is used to specify one or more resource
dependencies. If the resource being referenced in a depends-on annotation is
included in the input to the [`ensure-name-substring`] function, then the
function will also apply the ensure-name-substring logic to the name portion of
the annotation.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/ensure-name-substring-depends-on
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
  - image: ghcr.io/kptdev/krm-functions-catalog/ensure-name-substring:latest
    configMap:
      prepend: prod-
```

The function configuration is provided using a `ConfigMap`. We set only one
key-value pair:
- `prepend: prod-`: The desired name substring.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render ensure-name-substring-depends-on
```

### Expected result

Check that:
- all resources have `metadata.name` prepended with `prod-`.
- the `Deployment` with name `wordpress` had its `depends-on`
annotation updated, since the corresponding `StatefulSet` is also included in
the package.
- the `Deployment` with name `bar` did not have its `depends-on`
annotation updated, since it references a resource which was not included in the
package.
- the `ClusterRoleBinding` with name `read-secrets-global` had its `depends-on`
annotation updated, since the corresponding `ClusterRole` is also included in
the package.

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/

[`depends-on`]: https://kpt.dev/reference/annotations/depends-on/
