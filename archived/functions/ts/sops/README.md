# SOPS

### Overview

Use [sops] to encrypt/decrypt KRM resources.

### Usage

Because of the need to provide private-key or other sensitive information for
resources decryption, this function is supposed to be called imperatevly in that use-case.
The sensitive information should be provided as ENV variables.

The encryption step typically should be performed before storing data back to git repo,
and that means that encryption part isn't part of render operation and is supposed to be called
imperatevly for that use-case as well.

Please note that the function behavior isn't idempotent for encryption: the encryption of the same data
will produce the different result each time. That's because of [sops encryption protocol implementation](https://github.com/mozilla/sops#encryption-protocol):
on every encryption invocation it generates a random internal key, encrypts all data with that, and after that it encrypts that key with different methods, e.g. with each pgp public key provided, each age recipient, each key from Vault &etc and puts all that data into sops metadata section. Because of that the recepient user who wants to decrypt the data typically needs only 1 private key - it's enough to decrypt the original random internal key and decrypt all the data after that.

The decryption operation is idempotent.


### FunctionConfig

The function configuration must be a ConfigMap.

The following keys can be used in the `data` field of the ConfigMap, and all of
them are optional:

- `cmd`: The operation that sops perform: `encrypt` or `decrypt`. The default
  is `decrypt`.
- `cmd-json-path-filter`: Operation will be performed only to the resources that match
  to the filter, e.g. `$[?(@.metadata.name=="somename" && @.kind=="somekind")]` will
  process only resources with name `somename` and kine `somekind`. The default is empty.
- `cmd-tolerate-failures`: Ignore sops error and keep KRM resource unchanged rather than
  exit the function with error. The default is `false`.
- `cmd-extra-params-json-path-filter`: Get additional set of params from the function input that matches json-filter. e.g. `$[?(@.metadata.name=="encrypt-keys")]`. Only
  ConfigMap and Secret are supported kinds. In case of Secret kind the value will be de-base64-ed before using it.
- `cmd-import-pgp` and `cmd-import-age`: pre-import keys provided in config file rather than ENV variables (see below). This should be used for public keys primarily.
- `verbose`: Sops verbose logging output is enabled. The default is `false`.
- `ignore-mac`: Sops will ignore Message Authentication Code during decryption. The default
  is `false`.
- `override-preexec-cmd`: The command that will be executed prior to sops exectution. The default is
  `[ "$SOPS_IMPORT_PGP" == "" ] || (echo "$SOPS_IMPORT_PGP" | gpg --import 2>/dev/null); [ "$SOPS_IMPORT_AGE" == "" ] || (echo "$SOPS_IMPORT_AGE" >> $XDG_CONFIG_HOME/sops/age/keys.txt);`.
- `override-import-cmd`: The command will be executed on each `cmd-import-[pgp|age]` field met in documents. The value is set to the 
  related ENV variable: `SOPS_IMPORT_PGP` or `SOPS_IMPORT_AGE`. This command allows to import encryption keys via ENV variables. The default is the same as `override-preexec-cmd`.
- `override-detached-annotations`: List of comma-separated annotations that will be removed from the KRM resource
  if exist prior to sops execution and added back after execution. That helps to avoid decryption issues
  in cases the composer (e.g. kpt) adds its internal annotations. The default is
  `config.kubernetes.io/index,config.kubernetes.io/path,config.k8s.io/id,kustomize.config.k8s.io/id`.
- all other provided keys will be converted to the sops command arguments using pattern `--<key name> <value>`, e.g.
  `unencrypted-regex: '^(kind|apiVersion|group|metadata)$'` will add sops parameter `--unencrypted-regex: '^(kind|apiVersion|group|metadata)$'`.

In order to encrypt or decrypt yaml, `sops` may accept a variety of ENV variables, e.g. to work
with Hashicorp Vault it will be necessary to set: `VAULT_ADDR` and
`VAULT_TOKEN`. This option can be used to set different encryption parameters that shouldn't be stored
in version control system repository, e.g. private keys, external services credentials.
This function introduces 2 additional ENV variables: `SOPS_IMPORT_PGP` and `SOPS_IMPORT_AGE` that must contain the PGP or AGE keys and that
make possible to work with PGP and AGE encryption.

For `pgp` if you have a file with keys it's possible to run:

```shell
$ kpt fn eval \
        --fn-config <path to decrypt config>.yaml \
        --env SOPS_IMPORT_PGP="$(cat <file with exported key>)" \
        --image ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest \
        <folder>
```

or if your keys are already presented in `gpg`-storage, it's possible to run:

```shell
$ kpt fn eval \
        --fn-config <path to decrypt config>.yaml \
        --env SOPS_IMPORT_PGP="$(gpg --armor --export-secret-keys)" \
        --image ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest \
        <folder>
```

To make `sops` decrypt `age` it's necessary to keep all keys in the single file `~/.config/sops/age/keys.txt`. If that file exists, it's possible to invoke sops function and provide it with that keys by the command:

```shell
$ kpt fn eval \
        --fn-config <path to decrypt config>.yaml \
        --env SOPS_IMPORT_AGE="$(cat ~/.config/sops/age/keys.txt)" \
        --image ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest \
        <folder>
```

Please refer to [example 1] and [example 2] to get more details.

[sops]: https://github.com/mozilla/sops
[example 1]: /contrib/examples/sops-age/
[example 2]: /contrib/examples/sops-pgp-and-age/
