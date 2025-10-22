---
title: "search-replace"
linkTitle: "search-replace"
tags: "mutator"
weight: 4
description: |
   Search and optionally replace fields across all resources
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

### Overview

<!--mdtogo:Short-->

Search and optionally replace fields across all resources.

<!--mdtogo-->

### FunctionConfig

There is a spectrum of configuration customization techniques as described in
[this document].

<!--mdtogo:Long-->

One of the most basic and simplest customization techniques is Search and Replace.
The user fetches a package of resources, searches all the files for fields matching
a criteria, and replaces their values.

Search matchers are provided with `by-` prefix. When multiple matchers
are provided they are AND’ed together. `put-` matchers are mutually exclusive.

#### Matchers

```
by-value
Match by value of a field.

by-value-regex
Match by Regex for the value of a field. The syntax of the regular expressions
accepted is the same general syntax used by Go, Perl, Python, and other languages.
More precisely, it is the syntax accepted by RE2 and described at
https://golang.org/s/re2syntax. With the exception that it matches the entire
value of the field by default without requiring start (^) and end ($) characters.

by-path
Match by path expression of a field. Path expressions are used to deeply navigate
and match particular yaml nodes. Please note that the path expressions are not
regular expressions.

put-value
Set or update the value of the matching fields. Input can be a pattern for which
the numbered capture groups are resolved using --by-value-regex input.

put-comment
Set or update the line comment for matching fields. Input can be a pattern for
which the numbered capture groups are resolved using --by-value-regex input.
```

We use ConfigMap to configure the `search-replace` function. The inputs are
provided as key-value pairs using `data` field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: search-replace-fn-config
data:
  by-path: metadata.name
  by-value: the-deployment
  put-value: my-deployment
```

The function can be invoked using:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 --fn-config /path/to/fn-config.yaml
```

Alternatively, data can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- 'by-path=metadata.name' 'put-value=the-deployment'
```

Supported Path expressions:

```yaml
a.b.c

a:
  b:
    c: thing # MATCHES
```

```yaml
a.*.c

a:
  b1:
    c: thing # MATCHES
    d: whatever
  b2:
    c: thing # MATCHES
    f: something irrelevant
```

```yaml
a.**.c

a:
  b1:
    c: thing1 # MATCHES
    d: cat
  b2:
    c: thing2 # MATCHES
    d: dog
  b3:
    d:
    - f:
        c: thing3 # MATCHES
        d: beep
    - f:
        g:
          c: thing4 # MATCHES
          d: boop
    - d: mooo
```

```yaml
a.b[1].c

a:
  b:
  - c: thing0
  - c: thing1 # MATCHES
  - c: thing2
```

```yaml
a.b[*].c

a:
  b:
  - c: thing0 # MATCHES
    d: what..ever
  - c: thing1 # MATCHES
    d: blarh
  - c: thing2 # MATCHES
    f: thingamabob
```

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

```shell
# Matches fields with value "3":
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-value=3
```

```shell
# Matches fields with value prefixed by "nginx-":
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-value-regex='ngnix-.*'
```

```shell
# Matches field with path "spec.namespaces" set to "bookstore":
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-path='metadata.namespace' by-value='bookstore'
```

```shell
# Matches fields with name "containerPort" arbitrarily deep in "spec" that have value of 80:
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-path='spec.**.containerPort' by-value=80
```

```shell
# Set namespaces for all resources to "bookstore", even namespace is not set on a resource:
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-path='metadata.namespace' put-value='bookstore'
```

```shell
# Search and Set multiple values using regex numbered capture groups
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-value-regex='something-(.*)' put-value='my-project-id-${1}'
metadata:
  name: something-foo
  namespace: something-bar
...
metadata:
  name: my-project-id-foo
  namespace: my-project-id-bar
```

#### Create setters examples

```shell
# Put the setter pattern as a line comment for matching fields.
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-value='my-project-id-foo' put-comment='kpt-set: ${project-id}-foo'
metadata:
  name: my-project-id-foo # kpt-set: ${project-id}-foo

# Setter pattern comments can be added to multiple values matching a regex numbered capture groups
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/search-replace:v0.1 -- by-value-regex='my-project-id-(.*)' put-comment='kpt-set: ${project-id}-${1}'
metadata:
  name: my-project-id-foo # kpt-set: ${project-id}-foo
  namespace: my-project-id-bar # kpt-set: ${project-id}-bar
```

<!--mdtogo-->

[this document]: https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/declarative-application-management.md
