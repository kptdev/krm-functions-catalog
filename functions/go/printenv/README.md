# printenv

## Overview

<!--mdtogo:Short-->

A test KRM function that prints environment variables to stderr and exits with code 1.

<!--mdtogo-->

This function is used in kpt e2e tests to verify that environment variables are correctly
passed to container functions via the `-e` flag.

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
