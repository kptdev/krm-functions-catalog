#!/bin/bash
set -euo pipefail

REPO_URL="https://github.com/yannh/kubernetes-json-schema.git"
TMP_DIR="tmp-jsonschema"
K8S_VERSION="master"

echo "📦 Updating schemas for Kubernetes version: ${K8S_VERSION}"

# 1️⃣ Create a temporary repo folder
rm -rf "$TMP_DIR"
mkdir -p "$TMP_DIR"
cd "$TMP_DIR"

# 2️⃣ Sparse checkout only the standalone folders for the desired version
git init
git remote add origin "$REPO_URL"
git config core.sparseCheckout true
echo "${K8S_VERSION}-standalone/" >> .git/info/sparse-checkout
echo "${K8S_VERSION}-standalone-strict/" >> .git/info/sparse-checkout
git pull --depth 1 origin master

cd ..

# 3️⃣ Create tarball with the pulled directories (keep their structure)
rm -f jsonschema-k8s.tar.gz
tar -czf jsonschema-k8s.tar.gz \
    -C "$TMP_DIR" "${K8S_VERSION}-standalone" "${K8S_VERSION}-standalone-strict"

# 4️⃣ Cleanup
rm -rf "$TMP_DIR"  # optional

echo "✅ jsonschema-k8s.tar.gz created successfully for K8s ${K8S_VERSION}!"
