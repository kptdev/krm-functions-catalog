# set-standard-labels: Manual Force Example

### Overview

This example demonstrates how to add kubernetes recommended labels to a kpt package via manually 
forcing the package (with package-context.yaml) to be treated as a blueprint package. 
by running the [`set-starndard-labels`] function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-stanrdard-labels-manual-force
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: custom-frontend
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-standard-labels:unstable
      configMap:
        isDeployment: false
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-stanrdard-labels-manual-force
```

### Expected result

Check the label is changed from `app.kubernetes.io/name: frontend` to 
`app.kubernetes.io/name: custom-frontend`.

[`set-labels`]: https://catalog.kpt.dev/set-standard-labels/v0.1/
