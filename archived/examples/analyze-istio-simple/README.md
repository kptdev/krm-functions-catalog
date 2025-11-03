# analyze-istio: Simple Example

The `analyze-istio` KRM config function detects potential issues with your
Istio configuration and outputs structured results detailing any errors found
during analysis. This example invokes the `analyze-istio` function using
declarative configuration.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/archived/examples/analyze-istio-simple
```

### Function Invocation

Invoke the function with the following command:

```shell
kpt fn render analyze-istio
```

### Expected Results

This should give the following output:

```shell
[ERROR] Schema validation error: gateway must have at least one server in object 'networking.istio.io/v1alpha3/Gateway//httpbin-gateway' in file example-config.yaml
error: exit status 1
```

The error comes from the httpbin-gateway resource in
`example-config.yaml`. Uncomment `spec.servers` in that file to fix the
error and rerun the command. This will return success (no output).
