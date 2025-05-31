package consts

const DomainFieldSpecs = `
domains:
# Blueprint iam-location
- path: spec/member
  regexPattern: (?P<prefix>group:\S+@)(?P<domain>\S+)
  group: iam.cnrm.cloud.google.com
  version: v1beta1
  kind: IAMPolicyMember
`
