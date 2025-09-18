# set-project-id

## Overview

<!--mdtogo:Short-->

The `set-project-id` function sets 'project-id'
[setter](https://catalog.kpt.dev/apply-setters/v0.2/?id=definitions) and
`cnrm.cloud.google.com/project-id` annotation to the provided project ID, only
if they are not already set.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

set-project-id function is expected to be executed imperatively like:

```shell
kpt fn eval --include-meta-resources --image ghcr.io/kptdev/krm-functions-catalog/set-project-id:latest -- 'project-id=foo'
```

The `set-project-id` function does the following:

1.  Sets the the 'project-id' setter to the provided project ID.
    *   If an 'apply-setters' function is present in the Kptfile which does not
        have the 'project-id' setter or sets it to an empty value, update that
        config to set the “project-id” setter to the provided project ID value.
    *   If no 'apply-setters' function is present in the Kptfile, add one with
        just the 'project-id' setter set to the provided project ID value.
    *   If no pipeline is declared in the Kptfile, declare it and add
        'apply-setters' function as mutator with the 'project-id' setter set to
        the provided project ID value.
2.  For all
    [Config Connector resources](https://cloud.google.com/config-connector/docs/reference/overview),
    check if they have the `cnrm.cloud.google.com/project-id` annotation set. If
    the annotation is not present or is set to an empty value, set it to the
    provided project ID value.

### FunctionConfig

This function supports `ConfigMap` `functionConfig` and expects the 'project-id'
value to be present in the map.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Setting the `project-id` setter on the package without setters.

Let's start with the Kptfile of an example package.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example-package
```

Invoke the function:

```shell
kpt fn eval --include-meta-resources --image ghcr.io/kptdev/krm-functions-catalog/set-project-id:latest -- 'project-id=foo'
```

Kptfile will be updated to the following:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example-package
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        project-id: foo
```

<!--mdtogo-->
