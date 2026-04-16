# Release Process

This doc covers the release process for the functions in the
krm-functions-catalog repo.

1. Checking the [CI status](https://github.com/kptdev/krm-functions-catalog/actions/workflows/ci.yaml) of the main branch. 
   If the CI is failing on the main, we need to fix it before doing a release.
1. Go to the [releases pages] in your browser.
1. Click `Draft a new release` to create a new release for a function. The tag
   version format should be `functions/go/{function-name}/{semver}`. e.g.
   `functions/go/set-namespace/v0.1.0`. The release name should be
   `{function-name} {semver}` (see [VERSIONING.md](VERSIONING.md) for the semver strategy).
   The release notes for this function should be in the body.
1. Click `Publish release` button.
1. Verify the new functions are released in ghcr.io/kptdev/krm-functions-catalog/{function-name}/{semver} or, if using the GitHub based CD flow, check
   the relevant [GitHub packages section](https://github.com/orgs/kptdev/packages?repo_name=krm-functions-catalog)
1. Send an announcement on the [kpt slack channel]

## Updating or creating function docs

After creating a release, open a new PR to update/create the docs:

1. Ensure you are on the `main` branch and it is up to date
2. Create a new branch for the doc update:
   ```shell
   git checkout -b docs/set-namespace-v0.4
   ```
3. Run the doc generation for the released function:
   ```shell
   make generate-docs FN=set-namespace
   ```
4. Preview the docs locally (see [documentation/README.md](documentation/README.md)):
   ```shell
   make serve-docs
   ```
5. Commit the generated docs and submit a PR

See `make help` for additional targets.

[repo]: https://github.com/kptdev/krm-functions-catalog
[releases pages]: https://github.com/kptdev/krm-functions-catalog/releases
[kpt slack channel]: https://kubernetes.slack.com/channels/kpt/
