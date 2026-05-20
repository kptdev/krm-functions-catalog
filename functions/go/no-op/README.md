# no-op

## Overview

<!--mdtogo:Short-->

A no-op KRM function that passes resources through unchanged.

<!--mdtogo-->

This function reads a ResourceList from stdin and writes it back to stdout without any modifications.
It is useful for testing and validating kpt pipeline behavior without altering resources.

<!--mdtogo:Long-->

## Usage

The `no-op` function requires no configuration. It simply passes all resources through unmodified.

This can be used in both **imperative** (`kpt fn eval`) and **declarative** (`functionConfig`) modes.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Run imperatively

```sh
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/no-op
```

### Declarative usage in Kptfile

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/no-op
```

<!--mdtogo-->
