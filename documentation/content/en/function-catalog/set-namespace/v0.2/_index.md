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

{{< listversions >}}

{{< listexamples >}}

## Overview

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

## Usage

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
- If there's a `RoleBinding` or `ClusterRoleBinding` resource, the function will
  update the namespace in the `ServiceAccount` if one of the following are true:
  1) the subject element `name` is `default`.
  2) the subject element `name` matches the name of a `ServiceAccount` resource declared in the package.
  
In the following example, the `set-namespace` function will update:
- `subjects[0].namespace` since `subjects[0].name` is `default`.
- `subjects[1].namespace` since `subjects[1].name` matches a `ServiceAccount`
  name declared in the package.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
  namespace: original-namespace
