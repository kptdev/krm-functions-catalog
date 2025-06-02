import {
  Configs,
  KubernetesObject,
  kubernetesObjectResult,
  Result
} from 'kpt-functions';
import { ChildProcess, spawn } from 'child_process';
import { Writable } from 'stream';

const DEFAULT_SCHEMA_LOCATION = '/jsonschema';

const SCHEMA_LOCATION = 'schema_location';
const ADDITIONAL_SCHEMA_LOCATIONS = 'additional_schema_locations';
const IGNORE_MISSING_SCHEMAS = 'ignore_missing_schemas';
const SKIP_KINDS = 'skip_kinds';
const STRICT = 'strict';

type Feedback = FeedbackItem[];

interface FeedbackItem {
  filename: string;
  kind: string;
  status: 'valid' | 'invalid';
  errors: string[];
}

/**
 *
 * @param configs
 */
export async function kubeval(configs: Configs): Promise<void> {
  const schemaLocation = configs.getFunctionConfigValue(SCHEMA_LOCATION);
  const additionalSchemaLocationsStr = configs.getFunctionConfigValue(
    ADDITIONAL_SCHEMA_LOCATIONS
  );
  const additionalSchemaLocations = additionalSchemaLocationsStr
    ? additionalSchemaLocationsStr.split(',')
    : [];
  const ignoreMissingSchemas = JSON.parse(
    configs.getFunctionConfigValue(IGNORE_MISSING_SCHEMAS) || 'false'
  );
  const skipKindsStr = configs.getFunctionConfigValue(SKIP_KINDS);
  const skipKinds = skipKindsStr ? skipKindsStr.split(',') : [];
  const strict = JSON.parse(configs.getFunctionConfigValue(STRICT) || 'false');

  const results: Result[] = [];

  const args = buildKubevalArgs(
    schemaLocation,
    additionalSchemaLocations,
    ignoreMissingSchemas,
    skipKinds,
    strict
  );

  for (const object of configs.getAll()) {
    await runKubeval(object, results, args);
  }

  configs.addResults(...results);
}

/**
 *
 * @param object
 * @param results
 * @param args
 */
async function runKubeval(
  object: KubernetesObject,
  results: Result[],
  args: string[]
): Promise<void> {
  const kubevalProcess = spawn('kubeval', args, {
    stdio: ['pipe', 'pipe', process.stderr],
  });
  const serializedObject = JSON.stringify(object);
  await writeToStream(kubevalProcess.stdin, serializedObject);
  kubevalProcess.stdin.end();
  const rawOutput = await readStdoutToString(kubevalProcess);

  if (rawOutput.includes('Failed initializing schema file')) {
    results.push(
      kubernetesObjectResult(
        `Validating arbitrary CRDs is not supported yet. You can skip them by setting ${IGNORE_MISSING_SCHEMAS} or ${SKIP_KINDS} in the function config:\n` +
          rawOutput,
        object,
        undefined,
        'error'
      )
    );
    return;
  }

  try {
    const feedback = JSON.parse(rawOutput) as Feedback;

    for (const { status, errors } of feedback) {
      if (status !== 'valid') {
        for (const error of errors) {
          const [path, ...rest] = error.split(':');
          let result;
          if (rest.length > 0) {
            result = kubernetesObjectResult(
              rest.join(':').trim(),
              object,
              {
                path,
              },
              'error'
            );
          } else {
            result = kubernetesObjectResult(error, object, undefined, 'error');
          }
          results.push(result);
        }
      }
    }
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : String(error);

    results.push(
      kubernetesObjectResult(
        'Failed to parse raw kubeval output:\n' +
          message +
          '\n\n' +
          rawOutput,
        object
      )
    );
  }
}

/**
 *
 * @param schemaLocation
 * @param additionalSchemaLocations
 * @param ignoreMissingSchemas
 * @param skipKinds
 * @param strict
 */
function buildKubevalArgs(
  schemaLocation: string | undefined,
  additionalSchemaLocations: string[],
  ignoreMissingSchemas: boolean,
  skipKinds: string[],
  strict: boolean
) {
  const args = ['--quiet', '--output', 'json'];

  if (schemaLocation) {
    args.push('--schema-location');
    args.push(schemaLocation);
  }

  if (additionalSchemaLocations.length > 0) {
    args.push('--additional-schema-locations');
    args.push(additionalSchemaLocations.join(','));
  }

  if (!schemaLocation && additionalSchemaLocations.length === 0) {
    args.push('--schema-location');
    args.push('file://' + DEFAULT_SCHEMA_LOCATION);
  }

  if (ignoreMissingSchemas) {
    args.push('--ignore-missing-schemas');
  }

  if (skipKinds.length > 0) {
    args.push('--skip-kinds');
    args.push(skipKinds.join(','));
  }

  if (strict) {
    args.push('--strict');
  }
  return args;
}

/**
 *
 * @param stream
 * @param data
 */
function writeToStream(stream: Writable, data: string): Promise<void> {
  return new Promise((resolve, reject) =>
    stream.write(data, 'utf-8', (err) => (err ? reject(err) : resolve()))
  );
}

/**
 *
 * @param childProcess
 */
function readStdoutToString(childProcess: ChildProcess): Promise<string> {
  return new Promise<string>((resolve) => {
    let result = '';
    childProcess.stdout!.on('data', (data) => {
      result += data.toString();
    });
    childProcess.on('close', () => {
      resolve(result);
    });
  });
}

kubeval.usage = `
Use kubeval to validate KRM resources against their json schemas.

The function configuration must be a ConfigMap.

The following keys can be used in the data field of the ConfigMap, and all of
them are optional:

${SCHEMA_LOCATION}: The base URI used to fetch the json schemas. The default is
  empty. This feature only works with imperative runs, since declarative runs
  allow neither network access nor volume mount.
${ADDITIONAL_SCHEMA_LOCATIONS}: List of secondary base URIs used to fetch the
  json schemas.  These URIs will be used if the URI specified by
  ${SCHEMA_LOCATION} did not have the required schema.  The default is empty.
  This feature only works with imperative runs.
${IGNORE_MISSING_SCHEMAS}: Skip validation for resources without a schema. The
  default is false.
${SKIP_KINDS}: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas. The default is empty.
${STRICT}: Disallow additional properties that are not in the schemas. The
  default is false.

The following is an example function configuration:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  schema_location: "file:///abs/path/to/your/schema/directory"
  additional_schema_locations: "https://kubernetesjsonschema.dev,file:///abs/path/to/your/other/schema/directory"
  ignore_missing_schemas: "false"
  skip_kinds: "DaemonSet,MyCRD"
  strict: "true"

If neither ${SCHEMA_LOCATION} nor ${ADDITIONAL_SCHEMA_LOCATIONS} is provided, we
will convert the baked-in OpenAPI document to json schemas and use them.
The baked-in OpenAPI document is from a GKE cluster with version v1.20.10. The
OpenAPI document contains kubernetes built-in types and GCP CRDs (including
Config Connector resources).
`;
