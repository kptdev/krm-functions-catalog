---
parent_function: "source-gcloud-config"
---
# source-gcloud-config: Simple Example

### Overview

This example demonstrates how to imperatively run `source-gcloud-config` function
to add a ConfigMap resource containing the gcloud configurations.

### Before you begin

1. Install [gcloud SDK](https://cloud.google.com/sdk/docs/install)
1. [Setting defaults for gcloud commands](https://cloud.google.com/artifact-registry/docs/gcloud-defaults)
   Suggest setting up your `project`, `region` and `zone`

### Fetch the executable

This package contains the `source-gcloud-config` executable (MacOS).
```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog/tree/master/archived/examples/source-gcloud-config-simple
```

If you are in Linux, please run the following to build the binary.
```shell
git clone git@github.com:kptdev/krm-functions-catalog.git
cd krm-functions-catalog/archived/functions/go/source-gcloud-config
go build -o ssource-gcloud-config
```

### Run the function

```shell
$ kpt fn eval --exec ./source-gcloud-config .
```

### Expected result

A new file `gcloud-config.yaml` should be added which contains a ConfigMap resource
 and have your gcloud default configurations stored in the `.data` field.