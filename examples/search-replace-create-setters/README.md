# search-replace: Create Setters Example

The `search-replace` function can also be used to search for fields using available matchers
and add [setter comments] to the resource fields. Please refer to [create-setters] 
documentation to create setters for simple use-cases. This is an advanced example
of creating fine-grained setter comments using path expressions, regex capture groups etc.

This is an advanced example depicting [setter] creation process using `search-replace` function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/examples/search-replace-create-setters
```

Let's start with the input resource

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
data:
  FQDN: nginx.com
  MY_BAR_URL: https://nginx.com/bar
  MY_BAZ_URL: https://nginx.com/baz
  MY_FOO_URL: https://nginx.com/foo
```

### Function invocation

Suppose you want to parameterize the value `nginx` in all the fields in `data` 
section but not the value of `metadata.name` in the `ConfigMap` resource.
You can target specific values using path expression and regex capture groups, 
and add setter comments as per your intent.

```shell
$ kpt fn eval search-replace-create-setters --image ghcr.io/kptdev/krm-functions-catalog/search-replace:latest -- \
by-path='data.**' by-value-regex='(.*)nginx.com(.*)' put-comment='kpt-set: ${1}${host}${2}'
```

### Expected result

Verify that the setter comments are added as below:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-urls
data:
  FQDN: nginx.com # kpt-set: ${host}
  MY_BAR_URL: https://nginx.com/bar # kpt-set: https://${host}/bar
  MY_BAZ_URL: https://nginx.com/baz # kpt-set: https://${host}/baz
  MY_FOO_URL: https://nginx.com/foo # kpt-set: https://${host}/foo
```

Please refer to [apply-setters] documentation for information about applying desired setter values.

[setter]: https://catalog.kpt.dev/apply-setters/v0.1/
[setters]: https://catalog.kpt.dev/apply-setters/v0.1/
[apply-setters]: https://catalog.kpt.dev/apply-setters/v0.1/
[create-setters]: https://catalog.kpt.dev/create-setters/v0.1/
[setter comments]: https://catalog.kpt.dev/apply-setters/v0.1/?id=definitions
