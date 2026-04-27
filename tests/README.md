# E2E Test Harness

This directory contains the e2e test runner for the KRM Functions Catalog.

## How it works

The test harness (`e2etest/e2e_test.go`) uses the
[kpt test runner](https://github.com/kptdev/kpt/tree/main/pkg/test/runner)
to recursively scan the repository for test cases. Any directory containing a
`.expected/` subdirectory is treated as a test case.

Test cases are found in two locations:

- `examples/` — user-facing examples that also serve as e2e tests
- `functions/go/<fn>/tests/` — edge case and error handling tests for each
  function

## Running tests

Run all e2e tests:

```shell
cd tests/e2etest
go test -v ./...
```

Run tests for a specific function (examples and edge case tests separately):

```shell
go test -v ./... -run "TestE2E/../../examples/set-labels"
go test -v ./... -run "TestE2E/../../functions/go/set-labels/tests"
```

Run a single test case:

```shell
go test -v ./... -run "TestE2E/../../examples/set-labels-simple"
go test -v ./... -run "TestE2E/../../functions/go/set-labels/tests/empty-list"
```

## Updating expected output

If a function's behaviour changes, update the expected diff and results:

```shell
KPT_E2E_UPDATE_EXPECTED=true go test -v ./... -run "TestE2E/../../examples/set-labels-simple"
```

## Test case structure

Each test case directory contains:

- Resource files (YAML) — the input package
- `Kptfile` — pipeline configuration (for render tests)
- `.expected/config.yaml` — test configuration (test type, exit code, image, etc.)
- `.expected/diff.patch` — expected git diff after running the function
- `.expected/results.yaml` — expected function results (optional)
- `.expected/setup.sh` — pre-test script (optional)
- `.expected/teardown.sh` — post-test script (optional)
- `.expected/exec.sh` — custom test command (optional, replaces kpt fn eval/render)

See the [kpt test runner README](https://github.com/kptdev/kpt/blob/main/pkg/test/runner/README.md)
for full configuration options.
