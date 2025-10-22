---
parent_function: "apply-replacements"
---
# apply-replacements: Simple Example

### Overview

In this example, we will see how to invoke the apply-replacements function
with a simple example

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/apply-replacements-simple
```

We use a `ApplyReplacements` object to configure the `apply-replacements` function. 

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: ApplyReplacements
metadata:
  name: replacements-fn-config
replacements:
- source: 
    kind: Pod
    name: my-pod
    fieldPath: spec
  targets:
  - select:
      name: hello
      kind: Job
    fieldPaths: 
    - spec.template.spec
    options:
      create: true
```

This replacement will take the `spec` of the Pod named "my-pod", and propagate it to the `spec.template.spec`
field of the Job named "hello", creating it if it isn't already there.

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render apply-replacements-simple
```

### Expected result

`job.yaml` should now have the following conents: 

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - image: busybox
        name: myapp-container
```
