# set-default-labels

## Overview

<!--mdtogo:Short-->

The `set-default-labels` function applies KPT label convention to your KPT package.


### `app.kubernetes.io/name` 

Running the function can add this label using the value from the root `Kptfile` name.

This convention believes that all KRM resources under the same kpt package should be served for a specific application.
An application (or KPT package) can be composed by other applications (nested sub KPT packages). 
So the function can accept multiple Kptfile and use the root Kptfile to set the app name.

<!--mdtogo-->

You can learn more about the recommended labels [here][recommended labels].

<!--mdtogo:Long-->

## Usage

This function should be run in a KPT package. It does not require function config.

### Run the function once
```shell
$ kpt fn eval --image set-default-labels:unstable
```

### Run the function in a Kpt pipeline

Execute the `set-default-labels` function and save the config to Kptfile pipeline if the function passes.
```shell
$ kpt fn eval -t mutator -s -i set-default-labels:unstable
```

Check the Kptfile file, it now contains the `set-default-labels` function in its pipeline.
```shell
$ cat Kptfile

apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: set-default-label-example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-default-labels:unstable
```

```shell
$ kpt fn render
```
<!--mdtogo-->


[recommended labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
