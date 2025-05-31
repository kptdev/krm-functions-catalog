package fieldspec

const (
	NameReferenceFieldSpecs = `
nameReference:
# Blueprint redis-bucket
- kind: RedisInstance
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/displayName
    kind: RedisInstance
    group: redis.cnrm.cloud.google.com
    version: v1beta1

# Blueprint spanner
- kind: SpannerInstance
  group: spanner.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/displayName
    kind: SpannerInstance
    group: spanner.cnrm.cloud.google.com
    version: v1beta1
  - path: spec/instanceRef/name
    kind: SpannerDatabase
    group: spanner.cnrm.cloud.google.com
    version: v1beta1
`
)
