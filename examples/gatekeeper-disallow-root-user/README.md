---
parent_function: "gatekeeper"
---
# gatekeeper: Disallow Root User

### Overview

This example demonstrates how to run [gatekeeper] function declaratively to
enforce the policy `Containers must not run as root` on resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/gatekeeper-disallow-root-user
```

There are 3 resources: a `ConstraintTemplate`, a `DisallowRoot` and
a `Deployment`.

The following is the `ConstraintTemplate` we use:

```yaml
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: disallowroot
spec:
  crd:
    spec:
      names:
        kind: DisallowRoot
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |-
        package disallowroot
        violation[{"msg": msg}] {
          not input.review.object.spec.template.spec.securityContext.runAsNonRoot
          msg := "Containers must not run as root"
        }
```

We can see that there is a violation when
field `spec.template.spec.securityContext.runAsNonRoot` is `false`. This policy
disallows containers to be run as root.

The following is the `Constraint` of kind `NoRoot` that instantiates
the `ConstraintTemplate` above:

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: DisallowRoot
metadata:
  name: disallowroot
spec:
  match:
    kinds:
      - apiGroups:
          - 'apps'
        kinds:
          - Deployment
```

We can see that this constraint only checks if `Deployment` violates the above
policy.

### Function invocation

Run the function:

```shell
$ kpt fn render gatekeeper-invalid-configmap --results-dir /tmp
```

### Expected result

Let's take a look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: ghcr.io/kptdev/krm-functions-catalog/gatekeeper:latest
    stderr: |-
      [error] apps/v1/Deployment/nginx-deploy : Containers must not run as root
      violatedConstraint: disallowroot
    exitCode: 1
    results:
      - message: |-
          Containers must not run as root
          violatedConstraint: disallowroot
        severity: error
        resourceRef:
          apiVersion: apps/v1
          kind: Deployment
          name: nginx-deploy
        file:
          path: deployment.yaml
```

You can find:

- a detailed error message complaining: `Containers must not run as root`
- what resource violates the constraints
- what constraint does it violate
- where does the resource live

To pass validation, let's set
field `spec.template.spec.securityContext.runAsNonRoot` to `true` in
the `Deployment` in `resources.yaml`. Rerun the command. It will succeed.

[gatekeeper]: https://catalog.kpt.dev/gatekeeper/v0.1/
