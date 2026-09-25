# set-default-labels: Nested Package Example

### Overview

This example shows how to run [`set-default-labels`] in a nested KPT package. 

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-default-labels-nested-package root-app
```

Take a look at the `Kptfile`, it has the `set-default-labels` configured. Note: you do not need to specify the function config via `configPath` or `configMap`. 

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: root-app
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-default-labels:unstable
```

### Function invocation

Invoke the function:

```shell
$ kpt fn render root-app
```

### Expected result
A [recommended label] `app.kubernetes.io/name: root-app` using the root package name is added to all the resources `labels` and `matchLabels` fields.

[`set-default-labels`]: https://catalog.kpt.dev/set-default-labels/v0.1/
[recommended label]: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/