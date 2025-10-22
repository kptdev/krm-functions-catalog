---
title: "remove-local-config-resources"
linkTitle: "remove-local-config-resources"
tags: "config sync, mutator"
weight: 4
description: |
   Removes `local-config` annotated resources
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

## Overview

<!--mdtogo:Short-->

Removes resources with the annotation `config.kubernetes.io/local-config: true` from the resource list.

<!--mdtogo-->

This function provides a quick way for users to prune resources that are only intended to work client-side.
For example, it can be useful for Config Sync to avoid picking up client-side resources for validation and hydration.

<!--mdtogo:Long-->

## Usage

The function will execute as follows:

1. Searched for defined resources in a package
2. Deletes the resources with the following annotation:
   `config.kubernetes.io/local-config: true`

`remove-local-config-resources` function can be executed imperatively as follows:

```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/remove-local-config-resources:v0.1
```

To execute `remove-local-config-resources` declaratively include the function in kpt package pipeline as follows:
```yaml
...
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/remove-local-config-resources:v0.1
...
```

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Consider the following package:

```
sample
├─ applied.yaml
└─ local.yaml
```

```yaml
# applied.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-applied
data:
  foo: bar
```

```yaml
# local.yaml
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  name: sample-hierarchy
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  config:
    - simple
  parentRef:
    external: "123456789012"
```

Invoke the function in the package directory:

```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/remove-local-config-resources:v0.1
```

The resulting package structure would look like this:

```
sample
└─ applied.yaml
```
