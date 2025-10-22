---
title: "upsert-resource"
linkTitle: "upsert-resource"
tags: "mutator"
weight: 4
description: |
   Upsert an existing KRM resource
menu:
  main:
    parent: "Function Catalog"
---

{{< listversions >}}

{{< listexamples >}}

### Overview

<!--mdtogo:Short-->

Insert a resource, or if the resource already exists, update the existing resource.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

Upsert is an operation that adds a resource(uniquely identified by Group, Kind, Name and Namespace)
if it does not already exist, or replaces it if it already exists in the input list of resources.
`upsert-resource` function offers a safe way to upsert a resource to the list of input resources.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Replace an existing resource

Let's start with the list of resources in a package:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: mySpace
spec:
  selector:
    app: foo
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: mySpace
spec:
  replicas: 3
```

#### Add a new resource

For the same input resource list above, pass the following resource to upsert.
Note that the name of the resource is `myService2`.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService2
  namespace: mySpace
spec:
  selector:
    app: bar
```

Invoking `upsert-resource` function adds the input resource to the package.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: mySpace
spec:
  selector:
    app: foo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: mySpace
spec:
  replicas: 3
---
apiVersion: v1
kind: Service
metadata:
  name: myService2
  namespace: mySpace
spec:
  selector:
    app: bar
```

<!--mdtogo-->
