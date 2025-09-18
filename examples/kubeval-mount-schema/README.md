# kubeval: Mount Schema Example

### Overview

If you want to use your own schema instead of the built-in schema, you can
follow this example.

This example demonstrates how to mount a local schema directory and then use it
with [`kubeval`] function to validate KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/kubeval-mount-schema
```

We have a `ReplicationController` in `replicationcontroller.yaml` that has a
schema violation:

- `spec.replicas` must not be a string.

We have a `jsonschema` directory that contains some json schema files that will
be used with `kubeval` function.

#### Schema

The json schema used in this example are converted from an openapi schema file
using [`openapi2jsonschema`](https://github.com/instrumenta/openapi2jsonschema).
We run the following command to do it:

```shell
$ openapi2jsonschema --stand-alone --expanded --kubernetes -o jsonschema path/to/openapi.json
```

### Function invocation

We can invoke function with the following command:

```shell
$ kpt fn eval kubeval-mount-schema -i ghcr.io/kptdev/krm-functions-catalog/kubeval:latest --results-dir /tmp \
  --mount type=bind,src="$(pwd)/kubeval-mount-schema/jsonschema",dst=/schema-dir/master-standalone \
  -- schema_location=file:///schema-dir
```

We mount the local schema directory into the container with path
`/schema-dir/master-standalone`. And then we tell the function the location of
the schema by specifying `schema_location=file:///schema-dir`. The function will
by default look for the `master-standalone` directory in the specified
`schema_location`.

### Expected Results

Let's look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: ghcr.io/kptdev/krm-functions-catalog/kubeval:latest
    exitCode: 1
    results:
      - message: 'Invalid type. Expected: [integer,null], given: string'
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec.replicas
        file:
          path: replicationcontroller.yaml
```

To fix the violation, replace the value of `spec.replicas` with an integer

Rerun the command, and it should succeed.

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/
