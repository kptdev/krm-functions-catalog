---
title: "apply-replacements"
linkTitle: "apply-replacements"
tags: "mutator"
weight: 4
description: |
   Use the kustomize replacements feature as a KRM function
menu:
  main:
    parent: "Function Catalog"
---

{{< listversions >}}

{{< listexamples >}}

### Overview

<!--mdtogo:Short-->

Use the [kustomize replacements] feature as a KRM function. 

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

We use a `ApplyReplacements` object to configure the `apply-replacements` function. 

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: ApplyReplacements
metadata:
  name: replacements-fn-config
replacements:
  # your replacements here
```

The syntax for `replacements` is described in the [kustomize replacements] docs. For example,
you can have a functionConfig such as:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: ApplyReplacements
metadata:
  name: replacements-fn-config
replacements:
- source:
    kind: Deployment
    fieldPath: metadata.name
  targets:
  - select: 
      name: my-resource
```
<!--mdtogo-->

[kustomize replacements]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/replacements/ ':target=_blank'
