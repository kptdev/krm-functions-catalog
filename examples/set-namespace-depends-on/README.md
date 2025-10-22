---
parent_function: "set-namespace"
---
# set-namespace: Depends on example

### Overview

This example demonstrates how the [`set-namespace`] function interacts with
resources that use the [`depends-on`] annotation. The [`depends-on`] annotation
is a special annotation which is used to specify one or more resource
dependencies. These resources can be namespaced, having the form
`<group>/namespaces/<namespace>/<kind>/<name>`. If the resource being referenced
in a depends-on annotation is namespaced and included in the input to the
[`set-namespace`] function, then the function will also update the namespace
portion of the annotation.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-namespace-depends-on
```

We use the following `set-namespace-depends-on/Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-namespace:latest
      configMap:
        namespace: example-ns
```

The function configuration is provided using a `ConfigMap`. We set only one
key-value pair:
- `namespace: example-ns`: The desired namespace.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-depends-on
```

### Expected result

Check that:
- all resources have `metadata.namespace` set to `example-ns`.
- the `Deployment` with name `wordpress` had its `depends-on`
annotation updated, since the corresponding `StatefulSet` is also included in
the package.
- the `Deployment` with name `bar` did not have its `depends-on`
annotation updated, since it references a namespaced resource which was not
included in the package.

[`set-namespace`]: {{< relref "function-catalog/set-namespace/v0.4/" >}}
[`depends-on`]: https://kpt.dev/reference/annotations/depends-on/
