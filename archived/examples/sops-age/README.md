# sops: AGE Workflow example

### Overview

This example demonstrates invocation of `sops` KRM config function to decrypt and encrypt all other resources using the
already existing AGE keys. That includes decryption and encryption of some meta-resources, e.g.
[apply-setters.yaml](apply-setters.yaml), that is a setter configuration used for rendering and that may contain
sensitive information, e.g. passwords, keys & etc and that may be necessary to keep encrypted in git repo.

The `sops` KRM config function encrypts and decrypts resources using the sops tool. Learn more on the [sops website].

### Fetch the example package

Get the example package by running the following commands:
```shell
$ kpt pkg get https://github.com/kptdev/krm-functions-catalog.git/contrib/examples/sops-age
```

The package resources, e.g. [deployment.yaml](deployment.yaml) as well as some meta-resources, e.g. [apply-setters.yaml](apply-setters.yaml)
are stored in encrypted form in git repo.

### Function invocation

Invoke the decryption by running the following command:

```shell
$ kpt fn eval \
        --fn-config sops-age/decrypt.yaml \
        --env SOPS_IMPORT_AGE="$(cat sops-age/age_keys.txt)" \
        --image ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest \
        --include-meta-resources \
        sops-age/
```

Note: `sops-age/age_keys.txt` is provided as an example and in real life it should be taken from outside of the package.

Modify the decrypted settings and invoke the rendering by running the following command:

```shell
$ sed -i 's/1.14.1/1.14.0/' sops-age/apply-setters.yaml
$ kpt fn render sops-age/
```

Note: since all resouces are decrypted at that point it's possible to do all standard operations like render, apply &etc.
See [Kptfile](Kptfile) to check what will be done on render invocation.

Invoke the encryption by running the following command:

```shell
$ kpt fn eval \
        --fn-config sops-age/encrypt.yaml \
        --include-meta-resources \
        --image ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest \
        sops-age/
```

Note: [encrypt-keys.yaml](encrypt-keys.yaml) contains info about users who will be able to decrypt resources.
It is intentionally made as a part of package and can be controlled by adding/removing keys from that file.

### Expected result

Verify the updated configuration after decryption step using command:

```shell
$ kpt pkg diff sops-age/
```

Both `deployment.yaml` and `apply-setters.yaml` have to be decrypted.
That means that all fields will now have decrypted values and the field `sops` that contained sops metadata
will disappear from both resources.

Verify the updated configuration after encryption step using the same command.
Both `deployment.yaml` and `apply-setters.yaml` have to be encrypted again.
That means that both documents will have all datafileds encrypted and the field `sops` with sops metadata
will be added to both resources.

### Function Reference

Please find the `sops` KRM config function description [here](/contrib/functions/ts/sops/README.md)

[sops website]: https://github.com/mozilla/sops#encrypting-using-age

