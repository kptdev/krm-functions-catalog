{{range $ref := .Organization}}{{ if $ref.HasIAMBindings }}
module "{{ $ref.GetResourceName }}-iam" {
  source  = "terraform-google-modules/iam/google//modules/organizations_iam"
  version = "~> 7.4"

  organizations = ["{{ $ref.Name }}"]

  bindings = {
    {{ range $role, $binding := $ref.GetIAMBindings }}
    "{{ $role }}" = [{{ range $member := $binding.Members }}
      "{{ $member.Member }}",{{ end }}
    ]
    {{ end }}
  }
}

{{end}}{{end}}{{range $ref := .Folder}}{{ if $ref.HasIAMBindings }}
module "{{ $ref.GetResourceName }}-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = [{{ $ref.GetTerraformId }}]

  bindings = {
    {{ range $role, $binding := $ref.GetIAMBindings }}
    "{{ $role }}" = [{{ range $member := $binding.Members }}
      "{{ $member.Member }}",{{ end }}
    ]
    {{ end }}
  }
}

{{end}}{{end}}{{range $ref := .IAMAuditConfig}}
resource "google_organization_iam_audit_config" "org_config" {
  org_id  = {{ $ref.Parent.GetTerraformId false}}
  service = "{{ .GetStringFromObject "spec" "service" }}"
{{ range $cfg := $ref.GetIAMAuditLogConfigs }}
  audit_log_config {
      log_type = "{{ $cfg.LogType }}"{{ with $cfg.ExemptedMembers}}
      exempted_members = [{{ range $member := $cfg.ExemptedMembers }}
      "{{ $member }}"{{end}}
      ]{{end}}
  }{{end}}
}
{{end}}
