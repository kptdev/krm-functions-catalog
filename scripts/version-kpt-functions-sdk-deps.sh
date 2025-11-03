#!/usr/bin/env bash
# Copyright 2019 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Usage ex:
# version-kpt-functions-sdk-deps.sh 0.13.0.-rc.1

set -euo pipefail


TAG_VERSION=${1};

pushd .
cd "archived/functions/ts"
npm install "kpt-functions@${TAG_VERSION}"
git add package.json package-lock.json
popd
git commit -m "Update deps to use kpt-functions@${TAG_VERSION}"