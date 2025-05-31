# set-standard-labels: Abstract Example

### Overview

This example demonstrates how to add kubernetes recommended labels to kpt Abstract (or Catalog) packages
by running the [`set-starndard-labels`] function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-stanrdard-labels-abstract
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: base-app
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-standard-labels:unstable
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-stanrdard-labels-abstract
```

### Expected result

Check the deployment has label `app.kubernetes.io/name: base-app` added.

[`set-labels`]: https://catalog.kpt.dev/set-standard-labels/v0.1/
