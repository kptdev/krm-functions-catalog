# drop-comments

## Overview

<!--mdtogo:Short-->

A test KRM function that strips all YAML comments from resources.

<!--mdtogo-->

This function is used in kpt e2e tests to verify that kpt's kyaml library
correctly preserves comments even when a pipeline function removes them.

<!--mdtogo:Long-->

## Usage

The `drop-comments` function requires no configuration. It removes all YAML comments from resources via a JSON round-trip.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Run imperatively

```sh
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/drop-comments
```

<!--mdtogo-->
