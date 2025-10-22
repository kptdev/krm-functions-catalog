---
title: "set-enforcement-action"
linkTitle: "set-enforcement-action"
tags: "config sync, mutator"
weight: 4
description: |
   Set the enforcement action for policy constraints to deny or dryrun
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

## Overview

<!--mdtogo:Short-->

Applies the supplied enforcement action on [policy constraints](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller#constraints) within a package.

<!--mdtogo-->

Policy Controller allows users to define [constraints](https://cloud.google.com/anthos-config-management/docs/how-to/creating-constraints) to enforce policies for your Kubernetes clusters.
This function provides a quick way for users to set the [`spec.enforcementAction`](https://cloud.google.com/anthos-config-management/docs/how-to/auditing-constraints#types_of_enforcement_actions)
attribute to:
1. `dryrun` for auditing the cluster for policy violations or
2. `warn` for letting non-compliant resources be applied to the cluster with warnings or 
3. `deny` for enforcing the constraints and denying the resource application altogether

<!--mdtogo:Long-->

## Usage

The function will execute as follows:

1. Searches for resources with `apiVersion: constraints.gatekeeper.sh/v1beta1`
2. Applies the enforement action value provided in KptFile to following element:
   `spec.enforcementAction`

`set-enforcement-action` function can be executed imperatively as follows:

```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/set-enforcement-action:v0.1 -- enforcementAction=deny
```

To execute `set-enforcement-action` declaratively include the function in kpt package pipeline as follows:
```yaml
...
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-enforcement-action:v0.1
      configMap:
        enforcementAction: deny
...
```

<!--mdtogo-->
