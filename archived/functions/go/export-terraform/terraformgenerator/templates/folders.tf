{{range $folder := .Folder}}{{ if $folder.ShouldCreate }}
resource "google_folder" "{{ $folder.GetResourceName }}" {
  display_name = "{{ $folder.GetDisplayName }}"
  parent       = {{ $folder.Parent.GetTerraformId }}
}
{{end}}{{end}}

