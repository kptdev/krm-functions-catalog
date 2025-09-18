# ensure-name-substring: Prefix Example

### Overview

This example demonstrates how to declaratively run the [`ensure-name-substring`]
function to prepend prefix in the resource names.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/ensure-name-substring-prefix
```

We use the following `Kptfile` to run the function.

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

We are going to prepend prefix `prod-` to resource names.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render ensure-name-substring-prefix
```

### Expected result

Check all resources have `prod-` in their names:

We have a `Service` object whose name is `with-prod-service` which already
contains substring `prod-`. This resource will be skipped.

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/
