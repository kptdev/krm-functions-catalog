---
title: "set-namespace"
linkTitle: "set-namespace"
tags: "mutator"
weight: 4
description: |
  Upsert `namespace` for KRM resources 
menu:
  main:
    parent: "Function Catalog"
---

# set-namespace

{{< listversions >}}

{{< listexamples >}}

### Overview

<!--mdtogo:Short-->

The `set-namespace` function update or add namespace to all namespaced
resources. Kubernetes supports multiple virtual clusters backed by the same
physical cluster through namespaces.

Namespaces are often used in the following scenarios:

- Separate resources between environments (prod, staging and test).
- Separate resources between different team or users to divide resource quota.

<!--mdtogo-->

You can learn more about namespace [here][namespace].

<!--mdtogo:Long-->

### Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

For all namespaced resurces, the `set-namespace` function adds the namespace
if `metadata.namespace` doesn't exist. Otherwise, it updates the existing value.
It will skip the resources that are known to be cluster-scoped (e.g. `Node`
, `CustomResourceDefinitions`, `ClusterRole`). Whether a resource is namespaced
is determined by the OpenAPI schema. If the API path
contains `namespaces/{namespace}` then the resource is considered namespaced.
Otherwise, it's not. Currently, this function is using API version 1.20.4.

In addition to updating the `metadata.namespace` field for applicable resources,
by default the function will also update the [fields][commonnamespace] that
target the namespace. There are a few cases that worth pointing out:

- If there is a `Namespace` resource, its `metadata.name` field will be updated.
- If it's a `RoleBinding` or `ClusterRoleBinding` resource, `subject` can
  reference either a namespaced resource or a cluster-scoped resource. When
  there are references to namespace in the `subject` fields, they will be
  updated by the function.

This function can be used both declaratively and imperatively.

### FunctionConfig

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `SetNamespace`

To use a `ConfigMap` as the `functionConfig`, the desired namespace must be
specified in the `data.namespace` field.

To add a namespace `staging` to all resources, we use the
following `functionConfig`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  namespace: staging
```

To use a `SetNamespace` custom resource as the `functionConfig`, the desired
namespace must be specified in the `namespace` field. Sometimes you have
resources (especially custom resources) that have namespace fields in fields
other than the [defaults][commonnamespace], you can specify such label fields
using `additionalNamespaceFields`. It will be used jointly with the
[defaults][commonnamespace].

`additionalNamespaceFields` has following fields:

- `group`: Select the resources by API version group. Will select all groups if
  omitted.
- `version`: Select the resources by API version. Will select all versions if
  omitted.
- `kind`: Select the resources by resource kind. Will select all kinds if
  omitted.
- `path`: Specify the path to the field that the value needs to be updated. This
  field is required.
- `create`: If it's set to true, the field specified will be created if it
  doesn't exist. Otherwise, the function will only update the existing field.

To add namespace `staging` to all built-in resources and the
path `spec/selector/namespace` in in `MyKind` resource, we use the
following `functionConfig`:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetNamespace
metadata:
name: my-config
additionalNamespaceFields:
  - path: spec/selector/namespace
    kind: MyKind
    version: v1
    group: example.com
    create: true
```

<!--mdtogo-->

[namespace]: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/

