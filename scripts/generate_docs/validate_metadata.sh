#!/usr/bin/env bash
# Copyright 2026 The kpt Authors
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

# validate_metadata.sh - Validate all metadata.yaml files against the schema
#
# Usage:
#   ./scripts/generate_docs/validate_metadata.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCHEMA="${REPO_ROOT}/functions/go/metadata-schema.json"
FUNCTIONS_DIR="${REPO_ROOT}/functions/go"

if [ ! -f "$SCHEMA" ]; then
  echo "ERROR: schema not found at ${SCHEMA}"
  exit 1
fi

pip3 install jsonschema pyyaml --quiet 2>/dev/null || pip install jsonschema pyyaml --quiet 2>/dev/null || true

FAILED=0

for metadata_file in "${FUNCTIONS_DIR}"/*/metadata.yaml; do
  fn_name=$(basename "$(dirname "$metadata_file")")
  [ "$fn_name" = "_template" ] && continue

  result=$(python3 -c "
import json, yaml, sys
from jsonschema import validate, ValidationError

with open('${SCHEMA}') as f:
    schema = json.load(f)
with open('${metadata_file}') as f:
    data = yaml.safe_load(f)
try:
    validate(instance=data, schema=schema)
except ValidationError as e:
    print(e.message)
    sys.exit(1)
" 2>&1) || {
    echo "FAIL ${fn_name}: ${result}"
    FAILED=1
    continue
  }
  echo "  OK ${fn_name}"
done

if [ "$FAILED" -eq 1 ]; then
  echo ""
  echo "Metadata validation failed"
  exit 1
fi

echo ""
echo "All metadata files valid"
