# Contributing to krm-functions-catalog

We'd love to accept your contributions to this project. There are just a few
small guidelines you need to follow.

## Developer Certificate of Origin (DCO)

Contributors to this project should state that they agree with the terms published at https://developercertificate.org/
for their contribution. To do this when creating a commit with the Git CLI, a sign-off can be added with
[the -s option](https://git-scm.com/docs/git-commit#git-commit--s). The sign-off is stored as part of the commit message
itself.

## Copyright notices

All files should have the copyright notice.
```
// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```

If the file has never been modified: use the creation year only

* Example: `Copyright 2026 The kpt Authors`

If the file has been modified: use a year range from creation to last modification

* Example: `Copyright 2024-2026 The kpt Authors`

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult [GitHub Help] for more
information on using pull requests.

Process for code reviews. Before requesting human review, a PR must:

* All tests passing
* All linting passing
* Meeting project code quality requirements, including passing all configured
  static analysis / SonarCloud quality gates and not reducing automated test coverage for the
  affected components
* The comments from the first run of automatically generated comments (AI generated comments,
  SonarCloud comments, bot generated comments, etc.) of the PR are addressed (addressing further
  re-runs of AI are optional)
* If it is not possible to resolve an automatic comment, please add a sub-comment indicating why
  the automated comment cannot be resolved or ask for help in resolving the comment
* The PR description states whether AI was used to help create the PR; if so, it lists the AI tools
  used and the areas where they were used

All PRs should be approved by at least two members listed in the [CODEOWNERS](./CODEOWNERS) file, have all checks passing, and have all discussions resolved before merge.

### Stale PRs

If there is no activity from the PR author for more than 2 weeks (for example, no commits or replies to review comments), the PR will be closed.
The PR can be re-opened at any time if there is willingness to continue the work after a break.


## Declare any use of AI

> In addition to the above, the use of AI in the creation of PRs is allowed, but you must declare any use of AI and you must be able to explain the PR code independently of any AI tools.

Update the PR description to state whether you used AI to help you create this PR; if so, list the AI tools you have used and in what areas.

For example:
```text
I have used AI in the creation of this PR.

I have used the following AI tools:
- GitHub Copilot to analyze the code
- Claude Code to generate the function someNewFunctionIAdded()
- Amazon Q to generate unit tests
```

### Attribute AI in the Git commit messages

Following the [guidance of the Linux kernel](https://docs.kernel.org/process/coding-assistants.html#attribution)
we recommend the attribution of AI tools in the commit messages using the following format:

```text
Assisted-by: AGENT_NAME:MODEL_VERSION [TOOL1] [TOOL2]
```

Where:

- `AGENT_NAME` is the name of the AI tool or framework
- `MODEL_VERSION` is the specific model version used
- `[TOOL1] [TOOL2]` are optional specialized analysis tools used (e.g., coccinelle, sparse, smatch, clang-tidy)

Basic development tools (git, gcc, make, editors) should not be listed.

Example:

```text
Assisted-by: Claude:claude-3-opus coccinelle sparse
```


## Style Guides

Contributions are required to follow these style guides:

- [Error Message Style Guide]
- [Documentation Style Guide]

## How to Contribute

### Repo Layout

```shell
├── examples: Home for all function examples
│     ├── function_bar_example
│     └── function_foo_example
├── functions
│     └── go: Home for all golang-based function source code
│         ├── Makefile
│         ├── go_function_bar
│         └── go_function_foo
├── scripts
├── tests: Home for e2e tests
└── build
      └── docker
          └── go: Home for default golang Dockerfile
               └── Dockerfile
```

For each function, its files spread in the follow places:

- `functions/` directory: Each function must have its own directory in one
  of `functions/` sub-directory. In each function's directory, it must have the
  following:
    - Source code (and unit tests).
    - A README.md file serving as the usage doc and will be shown in
      the [catalog website]. Functions should follow [this template][golang-template].
    - A metadata.yaml file that follows the [function metadata schema](scripts/generate_docs/metadata-schema.json).
      See the [schema reference](documentation/content/en/metadata-schema/_index.md) for field descriptions.
    - (Optional) A Dockerfile to build the docker container. If a Dockerfile is
      not defined, the [default Dockerfile for the language][docker-common] will
      be used.
- `examples/` directory: It contains examples for functions, and these examples
  are also being tested as e2e tests. Each function should have at least one
  example here. There must be a README.md file in each example directory, and it
  should follow the [template][example-template].
- The `tests/` directory contains the e2e test harness (`tests/e2etest/`) which
  scans both `examples/` and `functions/go/*/tests/` for test cases.
- Each function may also have a `tests/` subdirectory inside its source
  directory (`functions/go/<fn>/tests/`) for edge case and error handling tests
  that are not suitable as user-facing examples.
- `master` branch should should contain examples with the `latest` tag for
  your function images.  When you release the function version that tag should 
  have the samples and tests that match the function version.

For golang-based functions, you need to generate some doc related variables from
the `README.md` by running

```shell
$ cd functions/go
$ make generate
```

### Tests

#### Unit Tests

To run all unit tests

```shell
$ make unit-test
```

#### Building a function image

Note: We use `docker buildx` to build images. Please ensure you have it installed.

To build all function images
```shell
$ make build
```

To build a single function image (e.g. `apply-setters`)
```shell
$ cd functions/go
$ make apply-setters-BUILD
```

#### E2E Tests

The e2e tests are the recommended way to test functions in the catalog. They are
very easy to write and set up with our e2e test harness. You can find all the
supported options and expected test directory
structure [here][e2e test harness doc].

You can choose to put the e2e test in either the `examples/` directory or in
the function's `tests/` subdirectory (`functions/go/<fn>/tests/`) depending on
whether it is worthwhile to be shown as a user-facing example.

**Note**: The e2e tests don't build the images. So you need to ensure you have built
the latest image(s) before running any e2e tests.

To test a specific example or the e2e test, run

```shell
$ cd tests/e2etest
$ go test -v ./... -run TestE2E/../../examples/$EXAMPLE_NAME

# To test edge case tests for a specific function
$ go test -v ./... -run TestE2E/../../functions/go/$FUNCTION_NAME/tests/$TEST_NAME

# To run all tests for a specific function (examples + edge cases)
$ go test -v ./... -run "TestE2E/../../examples/$FUNCTION_NAME"
$ go test -v ./... -run "TestE2E/../../functions/go/$FUNCTION_NAME/tests"
```

If you encounter some test failure saying something like "actual diff doesn't
match expected" or "actual results doesn't match expected", you can update the
expected `diff.patch` or `results.yaml` by running the following commands:

```shell
# Update one example
$ KPT_E2E_UPDATE_EXPECTED=true go test -v ./... -run TestE2E/../../examples/$EXAMPLE_NAME

# Update one edge case test
$ KPT_E2E_UPDATE_EXPECTED=true go test -v ./... -run TestE2E/../../functions/go/$FUNCTION_NAME/tests/$TEST_NAME

# Update all examples
$ KPT_E2E_UPDATE_EXPECTED=true go test -v ./...
```


Most contributors don't need this, but if you happen to need to test all
examples and e2e tests, run the following command

```shell
$ make e2e-test
```

### Documentation

For details on running the documentation site locally, refer to the
[documentation README](documentation/README.md).

### Change Existing Functions

You must follow the layout convention when you make changes to existing
functions.

If you implement a new feature, you must add a new example or modify existing
one to cover it.

If you fix a bug, you must add (unit or e2e) tests to cover that.

### Contribute New Functions

You must follow the layout convention when you contribute new functions.

You need to add new function name to the respective Makefile.

- `functions/go/Makefile` for golang.

## Contact Us

Do you need a review or release of functions? We’d love to hear from you!

* Message our [Slack channel]
* Join our [discussions]


## Useful links

- [catalog website](https://catalog.kpt.dev/)
- [e2e test harness doc](https://github.com/kptdev/kpt/blob/main/pkg/test/runner/README.md)
- [golang-template](https://raw.githubusercontent.com/kptdev/krm-functions-catalog/main/functions/go/_template/README.md)
- [docker](https://github.com/kptdev/krm-functions-catalog/tree/main/build/docker/go)
- [example-template](https://raw.githubusercontent.com/kptdev/krm-functions-catalog/main/examples/_template/README.md)
- [Slack channel](https://kubernetes.slack.com/channels/kpt/)
- [error message style guide](https://github.com/kptdev/kpt/blob/main/architecture/style-guides/errors.md)
- [GitHub Help](https://help.github.com/articles/about-pull-requests/)
- [discussions](https://github.com/kptdev/kpt/discussions)
