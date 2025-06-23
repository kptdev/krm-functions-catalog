# Release Process

This doc covers the release process for the functions in the
krm-functions-catalog repo.

1. Checking the [CI status](https://github.com/kptdev/krm-functions-catalog/actions/workflows/ci.yaml) of the master branch. 
   If the CI is failing on the master, we need to fix it before doing a release.
1. Go to the [releases pages] in your browser.
1. Click `Draft a new release` to create a new release for a function. The tag
   version format should be `functions/{language}/{function-name}/{semver}`. e.g.
   `functions/go/set-namespace/v1.2.3` and `functions/ts/kubeval/v2.3.4`. The release name should be
   `{funtion-name} {semver}`. The release notes for this function should be in
   the body. 
1. Click `Publish release` button.
1. Verify the new functions is released in gcr.io/kpt-fn/{funtion-name}/{semver} or, if using the GitHub based CD flow, check
   the relevant [GitHub packages section](https://github.com/orgs/kptdev/packages?repo_name=krm-functions-catalog)
1. Send an announcement on the [kpt slack channel]

Note: For most functions, you can ignore the GitHub action "Run the CD script after tags that look like versions". It is only for functions that use the KO build and release setup. 

## Updating function docs

After creating a release, the docs for the function should be updated to reflect
the latest patch version. A script has been created to automate this process.
The `RELEASE_BRANCH` branch should already exist in the [repo] and a tag should
be created on the [releases pages]. `RELEASE_BRANCH` is in the form of
`${FUNCTION_NAME}/v${MAJOR_VERSION}.${MINOR_VERSION}`.
For example `set-namespace/v0.3`, `kubeval/v0.1`, etc.

1. Setup the release branch 
	Release branch should have existed in the [upstream repo](https://github.com/kptdev/krm-functions-catalog) in the form of `<FUNCTION_NAME>/v<MAJOR>.<MINOR>`. Let's take `set-namespace/v0.4` as an example. You should replace that to your RELEASE_BRANCH.  
	```shell
	> export RELEASE_BRANCH=set-namespace/v0.4
	```
2. Clean up the local branch
	The release script needs to run in the local <RELEASE BRANCH>. To avoid git ref conflicts, we suggest you delete your local branch OR make it up to date with the remote <RELEASE BRANCH>
```shell
> git branch -D ${RELEASE_BRANCH}
```
3. Fetch the upstream repository
	Your `upstream` repo should point to the official kpt-functions-catalog. Verify your git remote is set as below
```shell
> git remote -v | grep upstream
upstream	git@github.com:kptdev/krm-functions-catalog.git (fetch)
upstream	git@github.com:kptdev/krm-functions-catalog.git (push)
# Fetch the latest upstream repo
> git fetch upstream
```
4. Run the doc updating script.
```shell
git checkout remotes/upstream/master
RELEASE_BRANCH=${RELEASE_BRANCH} make update-function-docs
```
5. Send out a Pull Request. 
	Your local git reference is now pointing to the local RELEASE BRANCH.
	A new git commit is auto-generated which contains the function document referring to the latest function version in the form of 
	`<FUNCTION_NAME>/v<MAJOR>.<MINOR>.<PATCH>`
	You should be ready to submit the Pull Request against the upstream <RELEASE_BRANCH>. 
```shell
> git push -f origin ${RELEASE_BRANCH}
```

[repo]: https://github.com/kptdev/krm-functions-catalog
[releases pages]: https://github.com/kptdev/krm-functions-catalog/releases
[kpt slack channel]: https://kubernetes.slack.com/channels/kpt/

