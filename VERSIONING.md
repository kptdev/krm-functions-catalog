# Versioning

## SemVer and shorter SemVer

We use [semantic versioning] for all of our images. The images with fully
specified version (e.g. `vX.Y.Z`) are immutable and will never be changed. We
also support a shorter version of semantic versioning. E.g.
`v<major>` and `v<major>.<minor>`.

The shorter version of semantic versioning is going to be floating tags:

- `v<major>.<minor>` always points to `v<major>.<minor>.X` where X is the latest
  (largest) patch version number. E.g. assume `v1.2` points `v1.2.0` initially.
  After `v1.2.1` is released, `v1.2` now points to `v1.2.1`.

- `v<major>` always points to `v<major>.X` where X is the latest (largest) minor
  version number. E.g. assume `v1` points `v1.2.0` initially. After `v1.3.0` is
  released, `v1` now points to `v1.3.0`. After `v1.3.1` is released, `v1` now
  points to `v1.3.1`.

Note: We do NOT officially support the `latest` tag, since we cannot provide any
compatibility guarantee for it, and the pipeline won’t produce deterministic
results. The `latest` tag should only be used for testing and development.

## User-facing Surfaces

All functions in the catalog comply with the [function spec]. To learn more
about the functions concept,
see [here](http://kpt.dev/book/02-concepts/03-functions).

There are 2 user-facing surfaces for a function:

- The `functionConfig`
- The function behavior

### functionConfig Surface

A `functionConfig` can be either a core resource (e.g. `ConfigMap`) or a custom
resource. If the `functionConfig` is a CRD, it can be versioned independently as
a normal CRD.

### Function Behavior Surface

#### What are NOT part of the function behavior surface

- The formatting of serialization for output items and results. e.g. yaml
  indentation and order of fields in a map.
- The order of resources in the output items.
- The order of result items in the results.
- The content of the unstructured messages in the results.

#### What are part of the function behavior surface

- The supported `functionConfig`:
    - If the function supports `ConfigMap` as `functionConfig`, the supported
      fields in the `ConfigMap`.
    - If the function supports a custom resource as `functionConfig`, the
      supported versions of the custom resource.
- How the function behave given the input items and `functionConfig`:
    - The remaining aspects of the output items that are not mentioned in the
      previous section.
    - The remaining aspects of the results that are not mentioned in the
      previous section.

For example, if the `kubeval` function stops supporting
the `ignore_missing_schemas` option in the `ConfigMap`, it will be a breaking
change.

Another example, if the `set-namespace` function stops supporting custom
resource of apiVersion `fn.kpt.dev/v1alpha1` and kind `SetNamespace`, it will be
a breaking change.

### Breaking Changes

We define a breaking change as: For any given input (including `input items`,
`functionConfig` and OpenAPI), the function produces a different output
(including output items, results) that are part of the user-facing surface.

### Backwards Compatibility

For post v1.0.0 versions, we will:

- Bump major version: There are breaking changes.

- Bump minor version: There are backward-compatible features.

- Bump patch version: There are only bug fixes and security fixes (e.g.
  dependency package non-breaking version bump and base image non-breaking
  version bump).

For pre v1.0.0 versions, the major version is always `0` and we will:

- Bumping minor version: There are breaking changes.

- Bumping patch version: In all other cases, including backward-compatible
  features, bug fixes and security fixes.

## Best Practices

There are 2 ways to specify your desired version:

- You can fully specify the whole SemVer: The benefits are that you get
  immutable infrastructure, and you control when to upgrade functions.
- You can use floating tags (e.g. `vX.Y` and `vX`): The benefits are that there
  are less maintenance toil, since it automatically pick up the security and bug
  fixes.

Avoid using the `latest` tag for your own function images, since
it’s [not a best practice] for production and also it
is [not recommended by kubernetes].

[not a best practice]: https://vsupalov.com/docker-latest-tag/

[not recommended by kubernetes]: https://kubernetes.io/docs/concepts/configuration/overview/#container-images

[semantic versioning]: https://semver.org/

[function spec]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
