# set-labels

## Overview

<!--mdtogo:Short-->

The `set-labels` function adds a list of labels to all resources. It's a common
practice to add a set of labels for all the resources in a package. Kubernetes
has some [recommended labels].

For example, labels can be used in the following scenarios:

- Identify the KRM resources by querying their labels.
- Set labels for all resources within a package (e.g. environment=staging).

<!--mdtogo-->

You can learn more about labels [here][labels].

## Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

For each label, the function adds it if it doesn't exist. Otherwise, it replaces
the existing label with the same name.

In addition to updating the `metadata.labels` field for each resource, the
function will also update the [selectors][commonlabels] that target the labels
by default. e.g. the selectors for `Service` will be updated to include the
desired labels.

This function can be used both declaratively and imperatively.

### FunctionConfig

<!--mdtogo:Long-->

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `SetLabels`

To use a `ConfigMap` as the `functionConfig`, the desired labels must be
specified in the `data` field.

To add 2 labels `color: orange` and `fruit: apple` to all resources, we use the
following `functionConfig`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange
  fruit: apple
```

To use a `SetLabels` custom resource as the `functionConfig`, the desired labels
must be specified in the `labels` field. 

To add 2 labels `color: orange` and `fruit: apple` to all built-in resources, we use the
following `functionConfig`:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
```

The `set-labels` function adds the desired labels to more fields than just the one
found in `.metadata`. For example, the `Deployment` resource has a `spec.selector`,
which gets updated by the `set-labels` function by default. This is not always desired.

To disable updating of `.spec.selectors`, define a `SetLabels` custom resource, and set
`.options.setSelectorLabels` to `false` like so:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
options:
  setSelectorLabels: false # default is true
```

<!--mdtogo-->

[labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/

[recommended labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

[commonlabels]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonlabels.go#L6
