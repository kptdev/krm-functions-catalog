---
parent_function: "search-replace"
---
# search-replace: Simple Example

### Overview

The `search-replace` function performs search and optionally replace fields
across all resources.

This is a simple example depicting search and replace operation on KRM resource config.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/search-replace-simple
```

Let's start with the input resource

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
  namespace: my-space
```

Search matchers are provided with `by-` prefix. When multiple matchers are
provided they are AND’ed together. `put-` matchers are mutually exclusive.

We use the following `ConfigMap` to provide input matchers to the function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
...
data:
  by-path: metadata.name
  by-value: the-deployment
  put-value: my-deployment
```

Invoking `search-replace` function would apply the changes to resource configs

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: my-space
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render search-replace-simple
```

### Expected result

Check the value of deployment `the-deployment` is changed to `my-deployment`.
