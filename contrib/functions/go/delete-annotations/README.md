# delete-annotations

## Overview

<!--mdtogo:Short-->

Deletes the supplied annotation keys from the resource(s).

<!--mdtogo-->

One of the use cases for this function is to help users remove annotations that are 
not necessary for deployment across a package. E.g. a user may add annotations to
resource(s) for local processing but those annotations may not be necessary for
deployment or for the functioning of the workload. This function can be used to
clean up such unnecessary annotations before resources are deployed.

<!--mdtogo:Long-->

## Usage

You can delete multiple annotations provided as a comma separated string as part of the function config.

To execute imperatively:
```shell
$ kpt fn eval -i ghcr.io/kptdev/krm-functions-catalog/contrib/delete-annotations:latest -- annotationKeys=annotation-to-delete,another-annotation-to-delete
```

To execute `delete-annotations` declaratively include the function in kpt package pipeline as follows:
```yaml
...
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/contrib/delete-annotations:latest
      configMap:
        annotationKeys: annotation-to-delete,another-annotation-to-delete
...
```

### FunctionConfig

This function takes the annotation key names as part of the function config parameter
`annotationKeys` where the key names can be provided as comma separated values as follows:

`annotationKeys=annotation-key-1,annotation-key-2`

In the previous example, the function will delete annotations `annotation-key-1` and `annotation-key-2`
in all resource(s) where those annotations are present.

The `annotationKeys` field is a required parameter.

<!--mdtogo-->
