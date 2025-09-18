# apply-setters

### Overview

<!--mdtogo:Short-->

Update the field values parameterized by setters.

#### Definitions

**Setters**: Setters serve as parameters for customizing field values.
Setters are a safer way to parameterize field values compared to common templating techniques.
By using comments instead of interleaving templating directives, the resource is still
valid, adheres to the KRM schema, and can be consumed by other tools. 


**Setter Name**: Name of the parameter.

**Setter Value**: Value of the parameter.

**Setter Comment**: A field value can be fully or partially parameterized using setter comments.
A setter comment can be derived by replacing all the instances of setter values 
in the field value, with the corresponding setter names along with 'kpt-set:' prefix.

```shell
e.g. image: ghcr.io/nginx:1.16.1 # kpt-set: ghcr.io/${image}:${tag}
```

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

We use ConfigMap to configure the `apply-setters` function. The desired setter
values are provided as key-value pairs using `data` field where key is the name of the
setter and value is the new desired value for the setter.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-func-config
data:
  setter_name1: setter_value1
  setter_name2: setter_value2
```

`apply-setters` function performs the following steps when invoked:
1. Searches for the field values tagged by setter comments.
2. Updates the field value fully or partially with the corresponding input setter values.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Setting scalar values

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

Declare the new desired values for setters in the functionConfig file.

```yaml
# apply-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  nginx-replicas: "3"
  tag: 1.16.2
```

Invoke the function:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/apply-setters:latest --fn-config ./apply-setters-fn-config
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/apply-setters:latest -- tag=1.16.2 nginx-replicas=3
```

Modified resource looks like the following:

```yaml
# resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 3 # kpt-set: ${nginx-replicas}
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
        image: "nginx:1.16.2" # kpt-set: nginx:${tag}
        ports:
        - protocol: TCP
          containerPort: 80
```

#### Setting array values

Array values can also be parameterized using setters. Since the values of configMap
in pipeline definition must be of string type, the array values must be wrapped into
string. However, the rendered values in the resources will be array type.

Let's start with the input resource

```yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev
  - stage
```

Declare the desired array values, wrapped into string.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  env: |
    - prod
    - dev
```

Invoke the function using the input config:

```shell
$ kpt fn eval --image ghcr.io/kptdev/krm-functions-catalog/apply-setters:latest --fn-config ./apply-setters-fn-config
```

Modified resource looks like the following:

```yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - prod
  - dev
```

<!--mdtogo-->

#### Note:

Refer to the `create-setters` function documentation for information about creating setters.
