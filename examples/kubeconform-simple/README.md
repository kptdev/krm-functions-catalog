# kubeconform: Simple Example

### Overview

This example demonstrates how to declaratively run [`kubeconform`] function to
validate KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/contrib/examples/kubeconform-simple
```

The following is the `Kptfile` in this example: 

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  validators:
    - image: ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib/kubeconform:latest
      configMap:
        strict: 'true'
        skip_kinds: MyCustom
```

The function configuration is provided using a `ConfigMap`. We set 2 key-value
pairs:
- `strict: 'true'`: We disallow unknown fields.
- `skip_kinds: MyCustom`: We skip resources of kind `MyCustom`.

### Function invocation

Try it out by running the following commands:

```shell
$ kpt fn render kubeconform-simple --results-dir /tmp
```

### Expected Results

Let's take a look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib/kubeconform:latest
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
          path: resources.yaml
      - message: additional properties 'templates' not allowed
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec
        file:
          path: resources.yaml
```

There are validation error in the `resources.yaml` file, to fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed now.

[`kubeconform`]:https://github.com/yannh/kubeconform
