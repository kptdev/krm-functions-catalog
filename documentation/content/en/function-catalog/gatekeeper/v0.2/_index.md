---
title: "gatekeeper"
linkTitle: "gatekeeper"
tags: "validator"
weight: 4
description: |
   Validate the KRM resources using Gatekeeper policies
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

## Overview

<!--mdtogo:Short-->

The `gatekeeper` function enforces policies on the package resources. You can
define policies for governance and legal requirements or to enforce best
practices and organizational conventions.

For example, you can enforce policies like:

- `ConfigMap` must not contain fields with `private_key` name
- All `pods` must have resource limits
- All `namespaces` must have a label that lists a point-of-contact

The `gatekeeper` function follows the [executable configuration] pattern.

<!--mdtogo-->

You can learn more about how to use the [`Gatekeeper`] project [here][howto].
The policies are expressed using the `OPA Constraint Framework`, you can read
more about it [here][concept].

<!--mdtogo:Long-->

## Usage

This function can be used both declaratively and imperatively.

There are 2 kinds of resources needed to define a policy, and they need to be
provided using `input items` along with other KRM resources to be validated.

- [Constraint Template]: Define the schema and logic of a policy. The policy
  logic in a Constraint Template must be written in the [Rego] language.
- [Constraint]: Signal the Gatekeeper the corresponding constraints need to be
  enforced. Every Constraint must be backed by a Constraint Template.

The constraint templates and the constraints resources should be in the same
package containing the KRM resources.

The following is a `ConstraintTemplate`:

```yaml
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: noroot
spec:
  crd:
    spec:
      names:
        kind: NoRoot
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |-
        package noroot
        violation[{"msg": msg}] {
          not input.review.object.spec.template.spec.securityContext.runAsNonRoot
          msg := "Containers must not run as root"
        }
```

This is a simple example of `ConstraintTemplate`, it contains several important
pieces:

- `targets`: What "target" the constraint applies to. You can learn more
  about "target" [here][target].
- `rego`: The logic that enforces the constraint.

You can learn more about `ConstraintTemplate` [here][GHConstraintTemplate]. You will find

- other fields commonly used in a `ConstraintTemplate` such as `validation`
  and `libs`.
- more detailed Rego semantics for defining your policies.

The following is a `Constraint` that instantiates the `ConstraintTemplate`
above.

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: NoRoot
metadata:
  name: noroot
spec:
  match:
    kinds:
      - apiGroups:
          - 'apps'
        kinds:
          - Deployment
```

<!--mdtogo-->

[`Gatekeeper`]: https://open-policy-agent.github.io/gatekeeper/website/docs/

[Constraint Template]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraint-templates

[Constraint]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraints

[Rego]: https://www.openpolicyagent.org/docs/latest/#rego

[howto]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto

[concept]: https://github.com/open-policy-agent/frameworks/tree/master/constraint#opa-constraint-framework

[target]: https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-target

[GHConstraintTemplate]: https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-constraint-template
