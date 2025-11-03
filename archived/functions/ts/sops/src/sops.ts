import {
  Configs,
  Severity,
  KubernetesObject,
  isKubernetesObject,
  generalResult,
  getAnnotation,
  addAnnotation,
  removeAnnotation,
  SOURCE_INDEX_ANNOTATION,
  LEGACY_SOURCE_PATH_ANNOTATION,
  LEGACY_SOURCE_INDEX_ANNOTATION,
  SOURCE_PATH_ANNOTATION,
  ID_ANNOTATION,
  LEGACY_ID_ANNOTATION,
} from 'kpt-functions';
import { DumpOptions, safeDump, safeLoad } from 'js-yaml';
import rw from 'rw';
import { spawnSync, execSync } from 'child_process';
import { JSONPath } from 'jsonpath-plus';
import base64url from 'base64url';
import { isConfigMap, isSecret } from './gen/io.k8s.api.core.v1';

const TEMP_PATH = '/tmp/tmp.yaml';

const CMD = 'cmd';
const CMD_FILTER = 'cmd-json-path-filter';
const CMD_TOLERATE_FAILURES = 'cmd-tolerate-failures';
const CMD_IMPORT_PGP = 'cmd-import-pgp';
const CMD_IMPORT_AGE = 'cmd-import-age';
const CMD_EXTRA_PARAMS_FILTER = 'cmd-extra-params-json-path-filter';
const IGNORE_MAC = 'ignore-mac';
const VERBOSE = 'verbose';
const OVERRIDE_PREEXEC_CMD = 'override-preexec-cmd';
const OVERRIDE_IMPORT_CMD = 'override-import-cmd';
const OVERRIDE_DETACHED_ANNOTATIONS = 'override-detached-annotations';

// the sops cmd we're doing now (decrypt/encrypt)
// decrypt by default for the backward compatibility
let cmd = 'decrypt';

// the documents must match to the filter
// to be processed. default - all
// example:
// cmd-json-path-filter: '$[?(@.metadata.name=="somename"&&@.kind=="somekind")]'
// will process only documents that have kind: somekind and metadata.name: somename
// and all others will be untouched
let cmdJsonPathFilter = '';

// if enabled this flag
// allows to skip documents that can't be decrypted/encrypted
// using the current secrets provided
// the document will stay intact and plugin won't fail
let cmdTolerateFailures = false;

// pre-exec command may be overriden from config
let preExecCmd =
  '[ "$SOPS_IMPORT_PGP" == "" ] || (echo "$SOPS_IMPORT_PGP" | gpg --import 2>/dev/null); \
  [ "$SOPS_IMPORT_AGE" == "" ] || (echo "$SOPS_IMPORT_AGE" >> $XDG_CONFIG_HOME/sops/age/keys.txt);';

// import command may be overriden from config
let importCmd = preExecCmd;

// list of annotations that will be detached before decryption
// this is needed, because tools like kpt add some annotations
// during processing, but sops fails trying to decrypt them.
// After decryption they will be attached back unchanged.
// it can be overriden from config
// TODO: there is a discusion about the complete list here:
// https://github.com/kubernetes-sigs/kustomize/issues/2996
// Please update the list when it is finished.
let detachedAnnotations: string[] = [
  SOURCE_INDEX_ANNOTATION,
  LEGACY_SOURCE_INDEX_ANNOTATION,
  SOURCE_PATH_ANNOTATION,
  LEGACY_SOURCE_PATH_ANNOTATION,
  ID_ANNOTATION,
  LEGACY_ID_ANNOTATION,
  'kustomize.config.k8s.io/id',
];

interface SopsKubernetesObject extends KubernetesObject {}

function isSopsKubernetesObject(
  o: KubernetesObject
): o is SopsKubernetesObject {
  return o && o.hasOwnProperty('sops');
}

function isKubernetsObjectMatchesJsonPathFilter(
  o: KubernetesObject,
  jsonPathFilter: string
): boolean {
  if (jsonPathFilter !== '') {
    const items = JSONPath({ json: [o], path: jsonPathFilter });
    if (items.length === 0) {
      return false;
    }
  }
  return true;
}

const YAML_STYLE: DumpOptions = {
  // indentation width to use (in spaces).
  indent: 2,
  // when true, will not add an indentation level to array elements.
  noArrayIndent: true,
  // 'undefined' is an invalid value for safeDump.
  // TODO(frankf): Explore ways to make this safer.
  // Either reason about not having 'undefined' in all cases OR
  // only skip 'undefined'.
  skipInvalid: true,
  // unset lineWidth from default of 80 to avoid reformatting
  lineWidth: -1,
  // avoid refs because many YAML parsers in the k8s ecosystem don't support them
  noRefs: true,
};

async function writeFile(path: string, data: string): Promise<void> {
  return new Promise((resolve, reject) => {
    rw.writeFile(path, data, 'utf8', (err: object) => {
      if (err) return reject(err);
      resolve();
    });
  });
}

function importCmdRun(key: string, value: string) {
  if (importCmd !== '') {
    const env = Object.assign({}, process.env);
    env[key] = value;
    execSync(importCmd, { env: { ...env } });
  }
}

export async function sops(configs: Configs) {
  // Validate config data and read arguments.
  const args = readSopsArguments(configs);

  // run preexec if it's not empty
  if (preExecCmd !== '') {
    execSync(preExecCmd, { env: process.env });
  }

  // Put all documents into the local place and
  // cleanup the configs storage. This is to keep
  // the order of documents
  const allDocs = configs.getAll();

  if (cmd === 'decrypt') {
    configs.deleteAll();
    for (const object of allDocs) {
      if (
        isSopsKubernetesObject(object) &&
        isKubernetsObjectMatchesJsonPathFilter(object, cmdJsonPathFilter)
      ) {
        // decrypt and add
        await decryptAndInsertSopsKubernetesObject(args, configs, object);
      } else {
        // add it back
        configs.insert(object);
      }
    }
  } else if (cmd === 'encrypt') {
    configs.deleteAll();
    for (const object of allDocs) {
      if (isKubernetsObjectMatchesJsonPathFilter(object, cmdJsonPathFilter)) {
        // encrypt and add
        await encryptAndInsertKubernetesObject(args, configs, object);
      } else {
        // add it back
        configs.insert(object);
      }
    }
  }
}

function detachAnnotations(object: SopsKubernetesObject): Map<string, string> {
  const detached = new Map<string, string>();

  for (const annotation of detachedAnnotations) {
    const value = getAnnotation(object, annotation);
    if (value !== undefined) {
      detached.set(annotation, value);
      removeAnnotation(object, annotation);
    }
  }
  return detached;
}

function attachDetachedAnnotations(
  object: SopsKubernetesObject,
  detached: Map<string, string>
) {
  detached.forEach((value: string, key: string) => {
    addAnnotation(object, key, value);
  });
}

async function decryptAndInsertSopsKubernetesObject(
  args: string[],
  configs: Configs,
  object: SopsKubernetesObject
) {
  args.push(...['-d', TEMP_PATH]);
  await cmdAndInsertKubernetesObject('decrypt', args, configs, object);
}

async function encryptAndInsertKubernetesObject(
  args: string[],
  configs: Configs,
  object: KubernetesObject
) {
  args.push(...['-e', TEMP_PATH]);
  await cmdAndInsertKubernetesObject('encrypt', args, configs, object);
}

async function cmdAndInsertKubernetesObject(
  cmd: string,
  args: string[],
  configs: Configs,
  object: KubernetesObject
) {
  let error;
  // write encrypted file to the temp file
  const detached = detachAnnotations(object);
  const stringifiedObject = safeDump(object, YAML_STYLE);
  await writeFile(TEMP_PATH, stringifiedObject);
  attachDetachedAnnotations(object, detached);

  try {
    // run sops
    const child = spawnSync('sops', args);
    error = child.stderr;
    // read the decrypted yaml from stdout and parse
    const decryptedObject = safeLoad(child.stdout);
    if (object && isKubernetesObject(decryptedObject)) {
      attachDetachedAnnotations(decryptedObject, detached);
      configs.insert(decryptedObject);
      return;
    }
  } catch (err) {
    configs.addResults(
      generalResult(
        `Exception for apiVersion: ${object.apiVersion}, kind: ${object.kind}, name: ${object.metadata.name}: ${err}`,
        'error'
      )
    );
  }
  if (error && error.length > 0) {
    let severity: Severity = 'error';
    if (cmdTolerateFailures) {
      severity = 'warn';
    }
    configs.addResults(
      generalResult(
        `Sops ${cmd} command results in error for\n${stringifiedObject}\n ${error.toString()}`,
        severity
      )
    );
  }
  // if we're here - there was an error
  // putting the original object back
  configs.insert(object);
}

function readSopsArguments(configs: Configs) {
  const args: string[] = [];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }

  const processArgument = (value: string, key: string): boolean => {
    if (key === CMD) {
      cmd = value;
    } else if (key === CMD_FILTER) {
      cmdJsonPathFilter = value;
    } else if (key === CMD_TOLERATE_FAILURES && value !== 'false') {
      cmdTolerateFailures = true;
    } else if (key === CMD_IMPORT_PGP) {
      importCmdRun('SOPS_IMPORT_PGP', value);
    } else if (key === CMD_IMPORT_AGE) {
      importCmdRun('SOPS_IMPORT_AGE', value);
    } else if (key === CMD_EXTRA_PARAMS_FILTER) {
      for (const object of configs.get(isConfigMap)) {
        if (isKubernetsObjectMatchesJsonPathFilter(object, value)) {
          if (object.data !== undefined) {
            for (const key in object.data) {
              if (!processArgument(object.data[key], key)) {
                return false;
              }
            }
          }
          // possibly support object.binaryData if needed
        }
      }
      for (const object of configs.get(isSecret)) {
        if (isKubernetsObjectMatchesJsonPathFilter(object, value)) {
          if (object.data !== undefined) {
            for (const key in object.data) {
              if (
                !processArgument(
                  base64url.decode(base64url.fromBase64(object.data[key])),
                  key
                )
              ) {
                return false;
              }
            }
          }
          if (object.stringData !== undefined) {
            for (const key in object.stringData) {
              if (!processArgument(object.stringData[key], key)) {
                return false;
              }
            }
          }
        }
      }
    } else if (key === OVERRIDE_PREEXEC_CMD) {
      preExecCmd = value;
    } else if (key === OVERRIDE_IMPORT_CMD) {
      importCmd = value;
    } else if (key === OVERRIDE_DETACHED_ANNOTATIONS) {
      detachedAnnotations = value.replace(/\s/g, '').split(',');
    } else if (key === VERBOSE || key === IGNORE_MAC) {
      args.push('--' + key);
    } else {
      args.push('--' + key);
      args.push(value);
    }
    return true;
  };
  configMap.forEach((value: string, key: string) => {
    processArgument(value, key);
  });
  return args;
}

sops.usage = `
Sops function (see https://github.com/mozilla/sops).

The function supports 2 operations: encrypt and decrypt.
Operation is set with the following parameter:
cmd: [Optional: default decrypt]           defines the operation that sops will
                                           perform

The typical configuration example for encryption:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-encrypt-config
  annotations:
    config.k8s.io/function: |
      container:
        image: ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest
        envs:
        - SOPS_IMPORT_PGP
        - SOPS_PGP_FP
    config.kubernetes.io/local-config: "true"
data:
  cmd: 'encrypt'
  unencrypted-regex: '^(kind|apiVersion|group|metadata)$'

To run this example it will be necessary to set 2 ENV variable values:
SOPS_IMPORT_PGP must contain the PGP keys that will be imported by the function
before encryption.
SOPS_PGP_FP must contain 1 or more key fingerprits separated with comma that
will be used for encryption.
E.g.:
curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
SOPS_IMPORT_PGP="$(cat key.asc)"
SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
export SOPS_IMPORT_PGP
export SOPS_PGP_FP

The typical configuration example for decryption:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-decrypt-config
  annotations:
    config.k8s.io/function: |
      container:
        image: ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest
        envs:
        - SOPS_IMPORT_PGP
    config.kubernetes.io/local-config: "true"
data:
  cmd: 'decrypt'

To run this example it will be necessary to set 1 ENV variable value:
SOPS_IMPORT_PGP must contain the PGP keys that will be imported by the function
before decryption. It has to contain the same keys that were used for
encryption.
E.g.
curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
SOPS_IMPORT_PGP="$(cat key.asc)"
export SOPS_IMPORT_PGP

There is a list of the parameters that don't depend on the command:
verbose: true [Optional: default empty]    Enable sops verbose logging output.
override-detached-annotations: [Optional:
default see detachedAnnotations var]       The list of annotations that didn't
                                           present when the document
                                           was encrypted, but added by
                                           different tools later. The function
                                           will detach them before decryption
                                           and added unchanged
                                           after successfull decryption. This
                                           allows sops to check the
                                           consistency of the decrypted document

cmd-json-path-filter: [Optional: default   The operation will be executed
empty]                                     only for the documents that
                                           match to the filter. E.g.
                                           cmd-json-path-filter:
                                           '$[?(@.metadata.name=="somename"
                                           &&@.kind=="somekind")]'
                                           will process documents with name
                                           'somename' and kind 'somekind'.

cmd-tolerate-failures: [Optional: default  The operation will continue
false]                                     even if it wasn't successfull
                                           and the document will stay
                                           intact.

If cmd is 'decrypt' the function runs sops -d for all documents that match the
json filter and that have field 'sops:'
and puts the decrypted result back. Decrypt has the following additional flags:
ignore-mac: true [Optional: default empty] Ignore Message Authentication Code
                                           during decryption.
keyservice value [Optional: default empty] Specify the key services to use in
                                           addition to the local one.
                                           Can be specified more than once.
                                           Syntax: protocol://address.
                                           Example: tcp://myserver.com:5000

If cmd is 'encrypt' the function runs sops -e for all documents that match the
json filter and puts the encrypted results back.
Function adds '--' to the field name, adds value and passes all these params to
the sops binary. For more details on the parameters plese refer to 'sops --help'

Example:

To decrypt the documents use:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: ghcr.io/kptdev/krm-functions-catalog/contrib/sops:latest
        envs:
	- SOPS_IMPORT_PGP
    config.kubernetes.io/local-config: "true"
data:
  verbose: true
`;
