---
title: "Function Metadata Schema"
linkTitle: "Metadata Schema"
weight: 1
description: |
  Reference for the metadata.yaml file required by each KRM function.
---

## Overview

Every function in the catalog must include a `metadata.yaml` file in its source
directory. This file provides machine-readable information about the function
that is used for:

- Generating the catalog documentation site
- Powering `kpt fn doc --doc` output
- Validating function contributions in CI

The schema is enforced automatically on every pull request via
`make validate-metadata`.

## Example

```yaml
image: ghcr.io/kptdev/krm-functions-catalog/set-labels:v0.1
description: Set labels on all resources.
tags:
  - mutator
sourceURL: https://github.com/kptdev/krm-functions-catalog/tree/main/functions/go/set-labels
examplePackageURLs:
  - https://github.com/kptdev/krm-functions-catalog/tree/main/examples/set-labels-simple
license: Apache-2.0
hidden: false
```

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | Container image reference (without tag). For functions in this catalog, the image must start with `ghcr.io/kptdev/krm-functions-catalog/`. The version tag is appended by the release process. |
| `description` | string | Yes | One or two sentence description of what the function does. Used in the catalog listing and `kpt fn doc` output. |
| `tags` | array of strings | Yes | At least one category tag. See [allowed values](#allowed-tags) below. |
| `sourceURL` | string | No | URL pointing to the function's source code directory on GitHub. |
| `examplePackageURLs` | array of strings | No | URLs to example kpt packages in the `examples/` directory. Referenced directories must exist in the repo. |
| `license` | string | No | SPDX license identifier. Defaults to `Apache-2.0`. |
| `hidden` | boolean | No | If `true`, the function is excluded from documentation generation and the catalog listing. Defaults to `false`. |

## Allowed Tags

The `tags` field accepts one or more of the following values. Only these values
are permitted — the schema validation will reject any other value.

```
mutator, validator, generator, viewer, test, testing, timing, name prefix, name suffix
```

## Validation

Metadata files are validated against the
[JSON Schema](https://github.com/kptdev/krm-functions-catalog/blob/main/scripts/generate_docs/metadata-schema.json)
on every pull request. To run validation locally:

```shell
make validate-metadata
```

The validator checks:
- All required fields are present
- Field types and values match the schema
- No unknown fields are present (`additionalProperties: false`)
- Referenced example directories in `examplePackageURLs` exist on disk

## Usage in Functions

The `metadata.yaml` file is embedded into the function binary using Go's
`//go:embed` directive and passed to `fn.WithDocs`:

```go
//go:embed metadata.yaml
var metadata []byte

func main() {
    runner := fn.WithContext(context.Background(), &MyFunction{})
    if err := fn.AsMain(runner, fn.WithDocs(readme, metadata)); err != nil {
        os.Exit(1)
    }
}
```

This enables `kpt fn doc --doc` to output machine-readable function metadata
without requiring access to the source repository.
