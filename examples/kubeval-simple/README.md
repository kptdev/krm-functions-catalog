# kubeval: Simple Example

### Overview

This example demonstrates how to declaratively run [`kubeval`] function to
validate KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/kubeval-simple
```

The following is the `Kptfile` in this example: 

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  validators:
    - image: ghcr.io/kptdev/krm-functions-catalog/kubeval:latest
      configMap:
        strict: 'true'
```

The function configuration is provided using a `ConfigMap`. We set 2 key-value
pairs:
- `strict: 'true'`: We disallow unknown fields.
- `skip_kinds: MyCustom`: We skip resources of kind `MyCustom`.

### Function invocation

Try it out by running the following commands:

```shell
$ kpt fn render kubeval-simple --results-dir /tmp
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
          path: resources.yaml
      - message: 'Invalid type. Expected: [integer,null], given: string'
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec.replicas
        file:
          path: resources.yaml
```

There are validation error in the `resources.yaml` file, to fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed now.

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/
