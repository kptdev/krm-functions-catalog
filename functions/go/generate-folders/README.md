# generate-folders

## Overview

The `generate-folders` function generates
[Folder](https://cloud.google.com/config-connector/docs/reference/resource-docs/resourcemanager/folder)
resources from `ResourceHierarchy` custom resources.

## Usage

The function accepts `ResourceHierarchy` resources in the following API versions:

| API Version | Status |
| --- | --- |
| `blueprints.cloud.google.com/v1alpha3` | **Current** |
| `dev.cft.v1alpha2` | Deprecated |
| `dev.cft.v1alpha1` | Deprecated |

### Simple Example (v3)

```yaml
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  name: my-hierarchy
spec:
  parentRef:
    kind: Organization
    external: "123456789"
  config:
    - Dev:
      - Team1
      - Team2
    - Prod
```

This produces `Folder` resources named `dev`, `dev.team1`, `dev.team2`, and
`prod`, each with proper parent references and display names.

### Subtree Example (v3)

```yaml
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  name: my-hierarchy
spec:
  parentRef:
    kind: Organization
    external: "123456789"
  subtrees:
    teams:
      - Team1
      - Team2
  config:
    - Dev:
        $subtree: teams
    - Prod:
        $subtree: teams
```

Subtrees allow reusing folder structure definitions across multiple branches.

### Parent Reference Types

- `kind: Organization` — uses `spec.organizationRef.external` on generated
  Folders (v3) or `cnrm.cloud.google.com/organization-id` annotation (v1/v2)
- `kind: Folder` — uses `spec.folderRef.external` on generated Folders (v3) or
  `cnrm.cloud.google.com/folder-ref` annotation (v1/v2)

### Annotation Inheritance

For v2/v3 `ResourceHierarchy` resources, annotations on the hierarchy object
are inherited by generated `Folder` resources, except for internal kpt
annotations (e.g., `config.kubernetes.io/local-config`,
`internal.config.kubernetes.io/*`). The v1 implementation does **not**
support annotation inheritance.

### Name Normalization

Generated folder names follow Kubernetes DNS subdomain naming rules:

- Converted to lowercase
- Quotes removed
- Underscores and spaces replaced with dashes
- Invalid characters removed
- Path segments joined with dots (e.g., `dev.team1`)

## Function Invocation

```shell
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/generate-folders:unstable
```

## Building

```shell
go build -o generate-folders .
```

## Testing

```shell
go test -v ./transformer/
```
