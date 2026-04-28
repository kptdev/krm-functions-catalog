# annotate-apply-time-mutations

## Overview

<!--mdtogo:Short-->

When deploying Kubernetes resources, you sometimes need to reference a field
value from one resource in another — but that value isn't known until the source
resource is created and reconciled (e.g. a project number, IP address, or
generated name).

[Apply-time mutations](https://kpt.dev/reference/annotations/apply-time-mutation/)
solve this by deferring field substitution to `kpt live apply` time. When
`kpt live apply` encounters a resource with the
`config.kubernetes.io/apply-time-mutation` annotation, it:

1. Ensures source resources are applied and reconciled first (implicit
   dependency ordering).
2. Reads the specified field values from the source resources.
3. Substitutes those values into the target resource before applying it.

The `annotate-apply-time-mutations` function generates these annotations from
simpler input formats, so you don't have to write the annotation YAML by hand.

Input formats:
- [Inline field comment: apply-time-mutation](#inline-field-comment-input)
- [Custom resource object: ApplyTimeMutation](#custom-resource-object-input)

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

The `annotate-apply-time-mutations` function can be executed by itself or as
part of a [kpt workflow](https://kpt.dev/book/02-concepts/#workflows).

To execute by itself:

```shell
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

To execute as part of a kpt workflow, first modify the Kptfile to add the
function to the pipeline:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

Then execute the pipeline:

```
kpt fn render
```

Either way, the function will read the input files and generate
`config.kubernetes.io/apply-time-mutation` annotations on the target object(s).

The function does not perform the mutation on the target object itself. The
mutation is performed later by `kpt live apply`, which reads the annotation as
input. So the function needs to be run before applying.

### Inline Field Comment Input

Inline field comments can be used as an alternate way to specify mutations. This
function will convert `apply-time-mutation` comments into apply-time-mutation
annotations.

With inline comments, the mutation is specified closer to the target field that
will be updated. This can aid debugging and onboarding by reducing indirection.
It also reduces the configuration required, because the target object and field
don't need to be explicitly specified.

Here is a simple example — inject a database host (only known after the
Database resource is reconciled) into a Deployment's environment variable:

```yaml
spec:
  containers:
  - name: app
    env:
    - name: DB_HOST
      value: "" # apply-time-mutation: ${example.com/namespaces/example/Database/my-db:$.status.host}
```

The general format for inline field comments is:

```yaml
field: "" # apply-time-mutation: [PREFIX]${GROUP/[VERSION/][namespaces/NAMESPACE/]KIND/NAME:FIELD_PATH}[SUFFIX]
```

Fields and delimiters surrounded in square brackets (`[]`) are optional. Your
comment should include the curly braces (`${}`) but NOT the square brackets.

- `PREFIX` (Optional) - A string to prepend to the substituted value.
- `GROUP` - The API group of the source object. Must be non-empty. For "core"
  group resources (e.g. Pod, Service, ConfigMap), use the
  [Custom Resource Object](#custom-resource-object-input) input method instead.
- `VERSION` (Optional) - The API version of the source object. When supplied, it
  will match only objects using this exact API version. It's recommended to
  just use `GROUP` without version to make the reference less brittle and able
  to survive CRD version updates.
- `NAMESPACE` (Optional) - The namespace of the source object, required for
  namespace-scoped resources
- `KIND` - The kind of the source object
- `NAME` - The name of the source object
- `FIELD_PATH` - A JSONPath expression that identifies the source object field
- `SUFFIX` (Optional) - A string to append to the substituted value

When the function runs, if `PREFIX` or `SUFFIX` is specified, the field value
will be replaced with a string including the `PREFIX` and `SUFFIX`, surrounding
a generated token for substitution.

```yaml
field: "PREFIX${ref1}SUFFIX" # apply-time-mutation: ...
```

When the function runs, if neither `PREFIX` nor `SUFFIX` are specified, the
field value will not be replaced and no token will be specified, causing the
whole field value to be replaced, using the type of the source object field.

The apply-time-mutation comment will be preserved so that the function is
idempotent, producing the same output when run multiple times.

### Custom Resource Object Input

Custom resource objects can be used to specify mutations. This function will
convert `ApplyTimeMutation` objects into apply-time-mutation annotations.

With custom resource objects, the mutation is specified with KRM, which is
more indirect, but allows for generation, manipulation, and templating of the
mutation specification. One big win with this method is that kpt setters can be
used to configure the source and target object references (ex: name & namespace).

`ApplyTimeMutation` resource objects can be specified with the following format:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: ApplyTimeMutation
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  targetRef:
    kind: ConfigMap
    name: target-object
    namespace: test-namespace
  substitutions:
  - sourceRef:
      kind: ConfigMap
      name: source-object
      namespace: test-namespace
    sourcePath: $.spec.data
    targetPath: $.spec.data
```

The `ApplyTimeMutation` resource follows the standard
[Kubernetes Resource Model (KRM)](https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/resource-management.md)
with top level `apiVersion`, `kind`, `metadata` fields, as well as the
conventional `spec` field for specification configuration. Like other KRM
resources, the `ApplyTimeMutation` resource also supports the standard metadata
fields, like label and annotation. The function will simply ignore them.

If you're familiar with the apply-time-mutation annotation syntax, the
`spec.substitutions` field of the `ApplyTimeMutation` resource should look
familiar. For details about the substitution schema, see the
[apply-time-mutation reference docs](https://kpt.dev/reference/annotations/apply-time-mutation/).

In addition to the substitutions, when using the `ApplyTimeMutation` resource,
the target object must be referenced. The `spec.targetRef` field uses the
[ObjectReference schema](https://kpt.dev/reference/annotations/apply-time-mutation/#objectreference).
The target object reference specifies which object will receive the
apply-time-mutation annotation, the object with target fields to be modified.

Remember to use the
[local-config annotation](https://kpt.dev/reference/annotations/local-config/)
so the resource is not applied by `kpt live apply`.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Inline Field Comment Example

Inject configuration values from custom resources into a Deployment using inline
comments. Demonstrates full field replacement (no token) and token substitution
with prefix/suffix.

```shell
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

### Custom Resource Object Example

Inject a Pod's IP address and port into another Pod's environment variable using
an `ApplyTimeMutation` resource. Demonstrates multiple substitutions on the same
target field using tokens, and referencing core group resources.

```shell
kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/annotate-apply-time-mutations:latest
```

<!--mdtogo-->
