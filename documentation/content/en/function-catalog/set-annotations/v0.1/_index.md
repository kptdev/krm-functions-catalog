---
title: "set-annotations"
linkTitle: "set-annotations"
tags: "mutator"
weight: 4
description: |
   Add a list of annotations to all resources
menu:
  main:
    parent: "Function Catalog"
---

{{< listversions >}}

{{< listexamples >}}

### Overview

<!--mdtogo:Short-->

The `set-annotations` function adds a list of annotations to all resources.
Annotations are commonly used in Kubernetes for attaching arbitrary metadata to
KRM resources.

For example, annotations can be used in the following scenarios:

- Provide information for controllers. (e.g. gce ingress controller will only
  take actions on `Ingress` resources with
  annotation `kubernetes.io/ingress.class: gce`)
- Tools store information for later use. (e.g. `kubectl apply` stores what a
  user applied previously in
  annotation `kubectl.kubernetes.io/last-applied-configuration`)

<!--mdtogo-->

You can learn more about annotations [here][annotations].

### Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

For each annotation, the function adds it if it doesn't exist. Otherwise, it
replaces the existing annotation value with the same key.

In addition to updating the `metadata.annotations` field for each resource, the
function will also update any fields that contain `ObjectMeta` (
e.g. `PodTemplate`) by default. e.g. field `spec.template.metadata.annotations`
in `Deployment` will be updated to include the desired annotations.

This function can be used both declaratively and imperatively.

#### FunctionConfig

<!--mdtogo:Long-->

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `SetAnnotations`

To use a `ConfigMap` as the `functionConfig`, the desired annotations must be
specified in the `data` field.

To add 2 annotations `color: orange` and `fruit: apple` to all resources:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange
  fruit: apple
```

To use a `SetAnnotations` custom resource as the `functionConfig`, the desired
annotations must be specified in the `annotations` field. Sometimes you have
resources (especially custom resources) that have annotations fields in fields
other than the [defaults][commonannotations], you can specify such annotations
fields using `additionalAnnotationFields`. It will be used jointly with the
[defaults][commonannotations].

`additionalAnnotationFields` has following fields:

- `group`: Select the resources by API version group. Will select all groups if
  omitted.
- `version`: Select the resources by API version. Will select all versions if
  omitted.
- `kind`: Select the resources by resource kind. Will select all kinds if
  omitted.
- `path`: Specify the path to the field that the value will be updated. This
  field is required.
- `create`: If it's set to true, the field specified will be created if it
  doesn't exist. Otherwise, the function will only update the existing field.

To add 2 annotations `color: orange` and `fruit: apple` to all built-in
resources and the path `data.selector.annotations` in `MyOwnKind` resource, we
use the following `functionConfig`:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetAnnotations
metadata:
  name: my-config
annotations:
  color: orange
  fruit: apple
additionalAnnotationFields:
  - path: data/selector/annotations
    kind: MyOwnKind
    create: true
```

<!--mdtogo-->

[annotations]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/

[commonannotations]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonannotations.go#L6
