package consts

// OrgFieldSpec contains the fieldSpec paths of Google Cloud OrganizationID
const OrgFieldSpec = `
organizationsIDs:
# Blueprint iam-foundation
- path: spec/resourceRef 
  group: iam.cnrm.cloud.google.com
  version: v1beta1
  kind: IAMPolicyMember
`
