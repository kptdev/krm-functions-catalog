---
parent_function: "format"
---
# format: Imperative Example

>? This example only works with 1.0.0-beta.1 or lower versions of kpt. Starting
>  from 1.0.0-beta.2, the order of the fields is preserved by kpt CLI.

This example demonstrates how to imperatively invoke the [format] function to
format KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/archived/examples/format-imperative
```

The `format` function formats the field ordering in YAML configuration files.
This example depicts the functionality of `format` function by formatting a
`Deployment` resource.

## Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval format-imperative --image ghcr.io/kptdev/krm-functions-catalog/archived/format:latest
```

## Expected result

The fields should be ordered as per OpenAPI schema definition of `Deployment`
resource. For e.g. `metadata.name` field is placed before `metadata.labels`
field. The keys in `metadata.labels` field are sorted alphabetically.
