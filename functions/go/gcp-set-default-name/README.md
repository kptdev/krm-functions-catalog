# set-default-name

### Feature
`set-default-name` sets two types of default name values:

- It sets the `metadata.name` fields for white listed KRM resources.
- It sets custom field specs to their referred `metadata.name` fields.  

In case one, you use `ConfigMap` to set the name. The ConfigMap should
contain a `.data.name` field and be provided as input resource. If the KRM 
resources are not listed, their `'metadata.name` won't be updated. 
The white list can be found in `./fieldspec/name.go`.

In case two, you can find the nameReferences whiteliest in `./fieldspec/nameref.go`
A `nameReference` contains a GVK and a list of referrals. The GVK gives 
the resources' `metadata.data` to read, the `referral` lists the resource 
field spec to update. This is different from kustomize nameReference, which 
only updates referrals metadata.name field and requires the referrals and 
referror have the same `metadata.name`.  
