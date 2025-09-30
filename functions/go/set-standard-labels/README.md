# set-standard-labels

## Overview

<!--mdtogo:Short-->

The `set-standard-labels` function adds the [recommended labels] to Kpt package resources.

<!--mdtogo-->

## Usage

For Blueprint Package

- Add `app.kubernetes.io/name` label using the Kpt package name.

For Deployment  Package

- Preserve `app.kubernetes.io/name` label from the upstream package.
- Add `app.kubernetes.io/instance` label using the Kpt package name.

### FunctionConfig

<!--mdtogo:Long-->

In most cases, you don't need to provide the function config. The `set-standard-labels` will
use the `Kptfile` and `package-context.yaml` info to find out the right values to update.

You can also force a package to be treated as "Deployment" or "Blueprint" via the function
config:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetStandardLabels
metadata:
  name: test
  annotations:
    config.kubernetes.io/local-config: "true"
forDeployment: true # add instance label.
```

You cannot force the `set-standard-labels` to use a different label values. If you need to 
customize the label values, please use [`set-labels`] function.

<!--mdtogo-->

[labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/

[recommended labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

[commonlabels]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonlabels.go#L6

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.2/