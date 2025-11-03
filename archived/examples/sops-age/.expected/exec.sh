#! /bin/bash

# decrypt
kpt fn eval --fn-config decrypt.yaml --env SOPS_IMPORT_AGE="$(cat age_keys.txt)" \
	--include-meta-resources \
	--image-pull-policy never --image ghcr.io/kptdev/krm-functions-catalog/archived/sops:latest

# make sure that docs don't contain ENC entries
grep 'ENC\[AES256_GCM' ./*.yaml && exit 1

# NOTE: at this point can modify setter config, do kpt render, apply, do 3way merge of submodules & etc

# encrypt
kpt fn eval --fn-config encrypt.yaml \
	--include-meta-resources \
	--image-pull-policy never --image ghcr.io/kptdev/krm-functions-catalog/archived/sops:latest

# make sure that docs contain ENC entries
grep 'ENC\[AES256_GCM' ./*.yaml || exit 1

# decrypt
kpt fn eval --fn-config decrypt.yaml --env SOPS_IMPORT_AGE="$(cat age_keys.txt)" \
        --include-meta-resources \
        --image-pull-policy never --image ghcr.io/kptdev/krm-functions-catalog/archived/sops:latest
