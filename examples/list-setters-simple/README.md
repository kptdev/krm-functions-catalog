---
parent_function: "list-setters"
---
# list-setters: Simple Example

### Overview

In this example, we will see how to list setters in a package.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/examples/list-setters-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/list-setters:latest
```

### Expected result

```shell
[RUNNING] "ghcr.io/kptdev/krm-functions-catalog/list-setters:latest"
[PASS] "ghcr.io/kptdev/krm-functions-catalog/list-setters:latest"
  Results:
    [INFO] Name: env, Value: [stage, dev], Type: array, Count: 1
    [INFO] Name: nginx-replicas, Value: 3, Type: int, Count: 1
    [INFO] Name: tag, Value: 1.16.2, Type: str, Count: 1
```

#### Note:

Refer to the [apply-setters] function documentation for information about updating the field values parameterized by setters.

[apply-setters]: {{< relref "function-catalog/apply-setters/v0.1/" >}}