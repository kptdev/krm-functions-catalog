---
parent_function: "ensure-name-substring"
---
# ensure-name-substring: Suffix Example

### Overview

This example demonstrates how to declaratively run the [`ensure-name-substring`]
function to append suffix in the resource names.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/ensure-name-substring-suffix
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
        append: -prod
```

We are going to append suffix `-prod` to resource names.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render ensure-name-substring-suffix
```

### Expected result

Check all resources have `-prod` in their names:

We have a `Service` object whose name is `the-service-prod` which already
contains substring`-prod`. This resource will be skipped.

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/
