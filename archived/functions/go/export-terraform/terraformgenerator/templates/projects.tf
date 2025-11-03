{{range $project := .Project}}{{ if $project.ShouldCreate }}
module "{{ $project.GetResourceName }}" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "{{ $project.GetDisplayName }}"{{ if ne $project.GetDisplayName $project.GetResourceName }}
  project_id = "{{ $project.GetResourceName }}"{{end}}
  org_id     = {{ $project.GetOrganization.GetTerraformId false }}{{if eq $project.Parent.Kind "Folder"}}
  folder_id  = {{ $project.Parent.GetTerraformId false }}{{end}}
{{ if $project.IsSVPCHost }}
  enable_shared_vpc_host_project = true{{end}}
  billing_account = {{ $project.References.BillingAccount.GetTerraformId false }}{{if $project.GetBool "metadata" "annotations" "cnrm.cloud.google.com/auto-create-network"}}
  auto_create_network = true{{end}}
}
{{end}}{{end}}

