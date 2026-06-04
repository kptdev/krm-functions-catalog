# foo

## Overview

<!--mdtogo:Short-->

A test KRM function that prints 'foo' to stderr and exits with code 1.

<!--mdtogo-->

This function is used in kpt e2e tests to verify image pull policy behavior.
It passes resources through unchanged but prints "foo" to stderr and returns a non-zero exit code.

<!--mdtogo:Long-->

## Usage

The `foo` function requires no configuration. It prints "foo" to stderr, passes resources through, and exits 1.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Run imperatively

```sh
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/foo
```

<!--mdtogo-->
