#! /bin/bash

set -euo pipefail

rm -rf new-dir

# We are not using eval here, since eval will also touch id annotation.
# We want to ensure the starlark function doesn't touch id annotations.
# fn sink doesn't allow us to write to existing directory, so we write a
# different directory and then copy the files.
kpt fn source --fn-config fn-config.yaml | docker run -i ghcr.io/kptdev/krm-functions-catalog/starlark:latest | docker run -i ghcr.io/kptdev/krm-functions-catalog/format:latest | kpt fn sink new-dir
mv new-dir/* .
rm -rf new-dir
