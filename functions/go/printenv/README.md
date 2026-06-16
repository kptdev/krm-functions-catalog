# printenv

## Overview

<!--mdtogo:Short-->

A test KRM function that prints environment variables to stderr and exits with code 1.

<!--mdtogo-->

This function is used in kpt e2e tests to verify that environment variables are correctly
passed to container functions via the `-e` flag.

> **Note:** This image is intended for controlled test environments only. It prints all
> environment variables passed to the container to stderr. When run as a container function,
> it only receives variables explicitly injected via `-e` flags - it does not inherit the
> host environment.

<!--mdtogo:Long-->

## Usage

The `printenv` function requires no configuration. It prints all environment variables to stderr and exits 1.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Run imperatively

```sh
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/printenv -e MY_VAR=hello
```

<!--mdtogo-->
