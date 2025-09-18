/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {
  Configs,
  configFileResult,
  kubernetesObjectResult,
  Severity,
  KubernetesObject,
} from 'kpt-functions';
import { spawnSync } from 'child_process';

const FLAG_ARGS = 'flags';
const USE_KUBE_FLAG = '--use-kube';
const OUTPUT_SHORT_FLAG = '-o';
const OUTPUT_LONG_FLAG = '--output';

interface IstioResult {
  code: string;
  level: 'Error' | 'Warn' | 'Info';
  origin: string;
  reference: string;
  message: string;
  documentation_url: string;
}

// Analyze istio configs using istioctl analyze.
export async function istioctlAnalyze(configs: Configs) {
  // Validate config data and read arguments.
  const args = readArguments(configs);

  for (const object of configs.getAll()) {
    try {
      const child = spawnSync('istioctl', args, {
        input: JSON.stringify(object),
        encoding: 'utf-8',
      });
      const error = child.stderr;
      if (error && error.length > 0) {
        configs.addResults(
          configFileResult(
            `Istioctl analyze command results in error: ${error.toString}`,
            '',
            'error'
          )
        );
      }
      if (child.stdout && child.stdout !== 'null') {
        const outputs: IstioResult[] = JSON.parse(child.stdout);
        addIstioResults(outputs, object, configs);
      }
    } catch (err) {
      configs.addResults(configFileResult(`${err}`, '', 'error'));
    }
  }
}

function addIstioResults(
  outputs: IstioResult[],
  object: KubernetesObject,
  configs: Configs
) {
  if (outputs && outputs.length) {
    outputs.forEach((output) => {
      const result = kubernetesObjectResult(
        output.message,
        object,
        undefined,
        output.level.toLowerCase() as Severity
      );
      result.tags = {
        ['documentation_url']: output.documentation_url,
        ['origin']: output.origin,
        ['code']: output.code,
      };
      configs.addResults(result);
    });
  }
}

function readArguments(configs: Configs) {
  // Initialize to output json
  const args: string[] = ['analyze', '-', '-o', 'json'];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }
  configMap.forEach((value: string, key: string) => {
    if (key === FLAG_ARGS) {
      args.push(value);
    } else if (key === OUTPUT_SHORT_FLAG || key === OUTPUT_LONG_FLAG) {
    } else if (key === USE_KUBE_FLAG) {
      // use-kube flag needs equals sign instead of space separator
      if (value !== undefined) {
        args.push(`${key}=${value}`);
      }
    } else {
      args.push(key);
      args.push(value);
    }
  });
  return args;
}

istioctlAnalyze.usage = `
Istioctl analyze is a diagnostic tool that can detect potential issues with
your Istio configuration and output errors to the results field. This function
runs against local configuration files to catch problems before you apply
changes to a cluster.

Configure this function using a ConfigMap with keys for "${FLAG_ARGS}" and
arbitrary istioctl analyze flags. The "${FLAG_ARGS}" argument takes an array of
flags which do not take arguments while flags which take their own arguments,
like --suppress, should be passed as separate arguments. The “${OUTPUT_SHORT_FLAG}” and “${OUTPUT_LONG_FLAG}”
flags are ignored as all output is included in config results. Consult the
reference for additional flags at:
https://istio.io/latest/docs/reference/commands/istioctl/#analyze-istio

Accepted arguments:
${FLAG_ARGS}: [Optional] List of flag arguments to istioctl analyze.

Example: Analyze '/path/to/istio/configs' recursively using '--use-kube=false'
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: ghcr.io/kptdev/krm-functions-catalog/contrib/analyze-istio:latest
    config.kubernetes.io/local-config: "true"
data:
  "${FLAG_ARGS}": ["--recursive"]
  "${USE_KUBE_FLAG}": "false"
`;
