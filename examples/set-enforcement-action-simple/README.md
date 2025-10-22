---
parent_function: "set-enforcement-action"
---
# set-enforcement-action: Simple Example

### Overview

In this example, we will see how to set enforcement action for policy constraints to
dryrun for auditing purposes.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/set-enforcement-action-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render set-enforcement-action-simple
```

### Expected result

The two policy constraints should have been mutated with their `spec.enforcementAction` 
elements set to `dryrun` which were initially set to `deny`

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRestrictRoleBindings
metadata: # kpt-merge: /restrict-clusteradmin-rolebindings
  name: restrict-clusteradmin-rolebindings
  annotations:
    # This constraint is not certified by CIS.
    description: "Restricts use of the cluster-admin role."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.1
spec:
  enforcementAction: dryrun
  parameters:
    restrictedRole:
      apiGroup: "rbac.authorization.k8s.io"
      kind: "ClusterRole"
      name: "cluster-admin"
    allowedSubjects:
    - apiGroup: "rbac.authorization.k8s.io"
      kind: "Group"
      name: "system:masters"
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sProhibitRoleWildcardAccess
metadata: # kpt-merge: /prohibit-role-wildcard-access
  name: prohibit-role-wildcard-access
  annotations:
    # This constraint is not certified by CIS.
    description: "Restricts use of wildcards in Roles and ClusterRoles."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.3
spec:
  enforcementAction: dryrun
```
