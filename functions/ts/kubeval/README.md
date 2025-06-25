# kubeval

## Overview

The `kubeval` function wraps the `kubeval` binary to validate resources
against their [json schemas].

This function is often used in the following scenarios:

- Validating resources as part of the local development workflow.
- Validating resources in CI.

## Usage

This function validates each resource using its json schema. If the json schema
is not available for a resource, the function will complain unless
the `ignore_missing_schemas` field is `true` or the kind of this resource
appears in the `skip_kinds` field.

This function can be used both declaratively and imperatively.

### FunctionConfig

The function configuration must be a ConfigMap.

The following keys can be used in the `data` field of the ConfigMap, and all of
them are optional:

- `schema_location`: The base URI used to fetch the json schemas. The default is
  empty. This feature only works with imperative runs, since declarative runs
  allow neither network access nor volume mount.
- `additional_schema_locations`: List of secondary base URIs used to fetch the
  json schemas. These URIs will be used if the URI specified
  by `schema_location` did not have the required schema. The default is empty.
  This feature only works with imperative runs.
- `ignore_missing_schemas`: Skip validation for resources without a schema. The
  default is `false`.
- `skip_kinds`: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas. The default is empty.
- `strict`: Disallow additional properties that are not in the schemas. The
  default is `false`.

The following is an example function configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  schema_location: "file:///abs/path/to/your/schema/directory"
  additional_schema_locations: "https://kubernetesjsonschema.dev,file:///abs/path/to/your/other/schema/directory"
  ignore_missing_schemas: "false"
  skip_kinds: "DaemonSet,MyCRD"
  strict: "true"
```

If neither `schema_location` nor `additional_schema_locations` is provided, we
will convert the baked-in OpenAPI document to json schemas and use them. The
baked-in OpenAPI document is from a GKE cluster with version v1.20.10. The
OpenAPI document contains kubernetes built-in types and GCP CRDs (including
Config Connector resources).

#### Convert OpenAPI to JSON Schema

If you want to convert OpenAPI to json schema, you can use
[openapi2jsonschema](https://github.com/instrumenta/openapi2jsonschema).

[json schemas]:https://json-schema.org
