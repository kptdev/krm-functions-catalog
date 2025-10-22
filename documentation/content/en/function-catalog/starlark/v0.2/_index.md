---
title: "starlark"
linkTitle: "starlark"
tags: "mutator, validator"
weight: 4
description: |
   Run a Starlark script to mutate or validate resources
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

## Overview

<!--mdtogo:Short-->

The `starlark` function contains a Starlark interpreter to run a Starlark script
to mutate or validate resources.

The starlark script can be used to:

- Add an annotation on the basis of a condition
- Inject a sidecar container in all KRM resources that contain a `PodTemplate`.
- Validate all KRM resources that contain a `PodTemplate` to ensure no
  containers run as root.

It follows the [executable configuration] pattern. It makes writing simple
function much easier.

<!--mdtogo-->

## Starlark Language

[Starlark] is a dialect of Python. It is commonly used as a configuration
language. It is an untyped dynamic language with high-level data types,
first-class functions with lexical scope, and garbage collection. You can find
the spec for the Starlark language [here][spec]. You can also find its API
reference [here][apiref].

<!--mdtogo:Long-->

## Usage

You need to put your starlark script source in the `functionConfig` of
kind `StarlarkRun` and then the function will run the starlark script that you
provide.

This function can be used both declaratively and imperatively.

### FunctionConfig

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `StarlarkRun`

To use a `ConfigMap` as the `functionConfig`, the starlark script source must be
specified in the `data.source` field. Additional parameters can be specified in
the `data` field.

Here's an example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: set-replicas
data:
  replicas: "5"
  source: |
    def setReplicas(resources, replicas):
      for r in resources:
        if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
          r["spec"]["replicas"] = replicas
    replicas = ctx.resource_list["functionConfig"]["data"]["replicas"]
    setReplicas(ctx.resource_list["items"], replicas)
```

In the example above, the script accesses the `replicas` parameters
using `ctx.resource_list["functionConfig"]["data"]["replicas"]`.

To use a `StarlarkRun` as the `functionConfig`, the starlark script source must
be specified in the `source` field. Additional parameters can be specified in
the `params` field. The `params` field supports any complex data structure as
long as it can be represented in yaml.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: conditionally-add-annotations
params:
  toMatch:
    config.kubernetes.io/local-config: "true"
  toAdd:
    configmanagement.gke.io/managed: disabled
source: |
  def conditionallySetAnnotation(resources, toMatch, toAdd):
    for resource in resources:
      match = True
      for key in toMatch:
        if key not in resource["metadata"]["annotations"] or resource["metadata"]["annotations"][key] != toMatch[key]:
          match = False
          break
      if match:
        for key in toAdd:
          resource["metadata"]["annotations"][key] = toAdd[key]
  toMatch = ctx.resource_list["functionConfig"]["params"]["toMatch"]
  toAdd = ctx.resource_list["functionConfig"]["params"]["toAdd"]
  conditionallySetAnnotation(ctx.resource_list["items"], toMatch, toAdd)
```

In the example above, the script accesses the `toMatch` parameters
using `ctx.resource_list["functionConfig"]["params"]["toMatch"]`.

There are 2 ways to run the function declaratively.

- Have your `Kptfile` with the inline `ConfigMap` as the `functionConfig`.
- Have your `Kptfile` pointing to a `functionConfig` file that contains either a
  `ConfigMap` or a `StarlarkRun`.

After that, you can render it with:

```shell
$ kpt fn render
```

There are 2 ways to run the function imperatively.

- Run it using a `ConfigMap` that generated from the command line arguments. The
  starlark script lives in `main.star` file.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/starlark:v0.2.2 -- source="$(cat main.star)" param1=value1 param2=value2
```

- Run it using `--fn-config` with either a `ConfigMap` or a `StarlarkRun` that
  lives in `fn-config.yaml`.

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/starlark:v0.2.2 --fn-config fn-config.yaml
```

### Developing Starlark Script

In Starlark, a [for loop] is permitted only within a function definition. It
means if you want to iterate over `ctx.resource_list["items"]`, it has to be in
a function. You can refer the example `functionConfig` above.

Here's what you can do in the Starlark script:

- Read resources from `ctx.resource_list`. The `ctx.resource_list` complies with
  the [KRM Functions Specification]. You can read the input resources from
  `ctx.resource_list["items"]` and the `functionConfig` from
  `ctx.resource_list["functionConfig"]`.
- Write resources to `ctx.resource_list["items"]`.
- Return an error using [`fail`][fail].
- Write error message to stderr using [`print`][print]

Here's what you currently cannot do in the Starlark script:

- We don't support the `results` field yet, i.e. if you write some results in
  `ctx.resource_list["results"]`, they will be ignored.
- While Starlark programs don't support working with yaml comments on resources,
  kpt will attempt to retain comments by copying them from the function inputs
  to the function outputs.

### Debugging

It is possible to debug the `starlark` functions using [`print`][print].

For example, you can add something like the following in your Starlark script:

```python
print(ctx.resource_list["items"][0]["metadata"]["name"])
```

Then you can run the function:

```shell
kpt fn render --results-dir /tmp
```

You will find your debugging output in `functionResultList items.stderr`.

<!--mdtogo-->

[Starlark]: https://docs.bazel.build/versions/master/skylark/language.html

[executable configuration]: https://kpt.dev/book/05-developing-functions/#authoring-executable-configuration

[spec]: https://github.com/bazelbuild/starlark/blob/master/spec.md

[apiref]: https://docs.bazel.build/versions/master/skylark/lib/skylark-overview.html

[KRM Functions Specification]: https://kpt.dev/book/05-developing-functions/#functions-specification

[for loop]: https://github.com/bazelbuild/starlark/blob/master/spec.md#for-loops

[fail]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#fail

[print]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#print
