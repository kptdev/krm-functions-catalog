---
parent_function: "generate-kpt-pkg-docs"
---
# generate-kpt-pkg-docs: Simple Example

### Overview

This example shows how the [generate-kpt-pkg-docs] function works.

Running `generate-kpt-pkg-docs` function on the example package will:

1. Update `GENERATED.md` file with the blueprint readme.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/generate-kpt-pkg-docs-simple
```

### Function invocation

Invoke the function with the following command:

```shell
$ kpt fn eval -i generate-kpt-pkg-docs:latest --include-meta-resources \
--mount type=bind,src="$(pwd)",dst="/tmp",rw=true -- readme-path=/tmp/GENERATED.md
```

### Expected result

1. File `GENERATED.md` will be updated with the generated readme.

[generate-kpt-pkg-docs]: https://catalog.kpt.dev/generate-kpt-pkg-docs/v0.1
