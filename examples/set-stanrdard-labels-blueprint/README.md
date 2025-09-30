# set-standard-labels: Blueprint Example

### Overview

This example demonstrates how to add kubernetes recommended labels to kpt Blueprint packages
by running the [`set-starndard-labels`] function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-stanrdard-labels-blueprint
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: frontend
  annotations:
    config.kubernetes.io/local-config: "true"
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-stanrdard-labels-blueprint
```

### Expected result

Check the deployment has label `app.kubernetes.io/name: frontend` added.

[`set-labels`]: https://catalog.kpt.dev/set-standard-labels/v0.1/
