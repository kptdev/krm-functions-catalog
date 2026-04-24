# annotate-apply-time-mutations: Inline Comment Example

### Overview

This example shows how to use the `annotate-apply-time-mutations` function with
inline field comments.

The scenario: a Deployment needs configuration values from other resources that
are only available after those resources are reconciled — a replica count from a
cluster config and a database host and port. Inline `apply-time-mutation`
comments on the target fields tell `kpt live apply` to defer the substitutions
to apply time.

This example exercises several inline comment patterns:
- Full field replacement with no token (replicas, DB_PORT)
- Token substitution with prefix/suffix (DB_URL)
- An existing annotation that is not modified (unmodified-key)
- A field without a comment that is not modified (unmodifiedField)

Running `annotate-apply-time-mutations` function on the example package will:

1. Scan resources for `apply-time-mutation` inline comments.
2. Generate the `config.kubernetes.io/apply-time-mutation` annotation on the
   target resources, which `kpt live apply` reads to perform the substitution.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/annotate-apply-time-mutations-inline-comment
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval annotate-apply-time-mutations-inline-comment --image ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

### Expected result

1. File resources.yaml will include `config.kubernetes.io/apply-time-mutation`
   annotations generated from the inline comments.
2. The DB_URL field with prefix/suffix will be updated with a generated
   replacement token.
3. Fields without comments and existing annotations are unchanged.

When `kpt live apply` is subsequently run, it will use these annotations to
wait for source resources to be reconciled and then substitute the referenced
field values into the target resources.

