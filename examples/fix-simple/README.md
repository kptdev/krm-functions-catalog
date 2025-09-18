# fix: Simple Example

In this example, we will fix a simple package which is compatible with v0.X version of kpt,
and make it compatible with kpt 1.0

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/fix-simple
```

Let's start with the input resources

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map # {"$kpt-set":"name"}
data:
  some-key: some-value
```

Here is an example Kptfile in the package:

```yaml
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: nginx
packageMetadata:
  shortDescription: describe this package
upstream:
  type: git
  git:
    commit: 4d2aa98b45ddee4b5fa45fbca16f2ff887de9efb
    repo: https://github.com/kptdev/kpt
    directory: package-examples/nginx
    ref: v0.2
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      x-k8s-cli:
        setter:
          name: name
          value: the-map
```

Invoking `fix` function on the package transforms the resources as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-new-map # kpt-set: ${name}
data:
  some-key: some-value
```

<!-- @skip -->
```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
upstream:
  type: git
  git:
    repo: https://github.com/kptdev/kpt
    directory: package-examples/nginx
    ref: v0.2
  updateStrategy: resource-merge
upstreamLock:
  type: git
  git:
    repo: https://github.com/kptdev/kpt
    directory: package-examples/nginx
    ref: v0.2
    commit: 4d2aa98b45ddee4b5fa45fbca16f2ff887de9efb
info:
  description: describe this package
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
```

The transformed package is compatible with kpt 1.0 binary.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval fix-simple --image ghcr.io/kptdev/krm-functions-catalog/fix:latest --include-meta-resources
```

### Expected result

Check the resources in the package are transformed as described above.
