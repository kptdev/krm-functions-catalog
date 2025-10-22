---
title: "list-setters"
linkTitle: "list-setters"
tags: "viewer"
weight: 4
description: |
   List setters on resource fields
menu:
  main:
    parent: "Function Catalog"
---
{{< listversions >}}

{{< listexamples >}}

## Overview

<!--mdtogo:Short-->

Lists information about [setters] like setter name, values and count.

Refer to the [create-setters] function documentation for information about creating new setters or [apply-setters] function documentation for information about updating the field values parameterized by setters.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

`list-setters` function is expected to be executed imperatively like

```shell
$ kpt fn eval -i list-setters:v0.1.0
```

`list-setters` function performs the following steps:

1. Searches for setter comments in input list of resources.
1. Lists discovered setters and related information.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Listing setters in a package

Let's start with the input resource in a package

```yaml
# resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 4 # kpt-set: ${nginx-replicas}
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: "nginx:1.16.1" # kpt-set: nginx:${tag}
        ports:
        - protocol: TCP
          containerPort: 80
```

Invoke the function:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/list-setters:v0.1
```

Output looks like the following:

```shell
  Results:
    [INFO] Name: nginx-replicas, Value: 4, Type: int, Count: 1
    [INFO] Name: tag, Value: 1.16.1, Type: str, Count: 1
```

<!--mdtogo-->

[setter]: {{< relref "function-catalog/apply-setters/v0.2/#definitions" >}}
[create-setters]: {{< relref "function-catalog/create-setters/v0.1/" >}}
