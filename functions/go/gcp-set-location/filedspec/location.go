package filedspec

const LocationFieldSpecs = `
regions:
- path: spec/location
  group: storage.cnrm.cloud.google.com
  version: v1beta1
  kind: StorageBucket

- path: spec/region
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  kind: RedisInstance

- path: spec/config
  group: spanner.cnrm.cloud.google.com
  version: v1beta1
  kind: SpannerInstance
  regexPattern: (?P<prefix>regional-)(?P<location>\S+)

zones:
- path: spec/location
  apiVersion: test.cnrm.cloud.google.com/v1beta1
  kind: Test
  create: false
`
