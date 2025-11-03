{{range $logsink := .LoggingLogSink}}{{ if $logsink.ShouldCreate }}
module "logsink-{{ $logsink.GetResourceName }}" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.{{with or $logsink.References.BigQueryDataset $logsink.References.PubSubTopic $logsink.References.StorageBucket $logsink.References.LoggingLogBucket }}{{.GetResourceName}}{{end}}-destination.destination_uri
  log_sink_name        = "{{ $logsink.GetDisplayName }}"
  parent_resource_id   = {{ $logsink.GetOrganization.GetTerraformId false }}
  parent_resource_type = "organization"
  include_children     = true
}
{{end}}{{with $logsink.References.BigQueryDataset }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/bigquery"
  version = "~> 7.3.0"

  project_id               = module.{{ .Parent.GetResourceName }}.project_id
  dataset_name             = "{{ .GetResourceName }}"
  log_sink_writer_identity = module.logsink-{{ $logsink.GetResourceName }}.writer_identity{{ with .GetInt "spec" "defaultTableExpirationMs" }}
  expiration_days          = "{{ . | msToDays }}"{{end}}{{ with .GetStringFromObject "spec" "location" }}
  location                 = "{{.}}"{{end}}
}
{{end}}{{with $logsink.References.PubSubTopic }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/pubsub"
  version = "~> 7.3.0"

  project_id               = module.{{ .Parent.GetResourceName }}.project_id
  topic_name               = "{{ .GetResourceName }}"
  log_sink_writer_identity = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
}
{{end}}{{with $logsink.References.StorageBucket }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/storage"
  version = "~> 7.3.0"

  project_id                  = module.{{ .Parent.GetResourceName }}.project_id
  storage_bucket_name         = "{{ .GetResourceName }}"
  log_sink_writer_identity    = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
  uniform_bucket_level_access = {{ .GetBool "spec" "uniformBucketLevelAccess" }}{{ with .GetStringFromObject "spec" "location" }}
  location                    = "{{.}}"{{end}}{{ with .GetStringFromObject "spec" "storageClass" }}
  storage_class               = "{{.}}"{{end}}{{ if .GetInt "spec" "retentionPolicy" "retentionPeriod"}}
  retention_policy = {
    retention_period_days = {{ .GetInt "spec" "retentionPolicy" "retentionPeriod" | sToDays }},
    is_locked             = {{ .GetBool "spec" "retentionPolicy" "isLocked" }}
  }{{end}}
}
{{end}}{{with $logsink.References.LoggingLogBucket }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/logbucket"
  version = "~> 7.4.1"

  project_id               = module.{{ .Parent.GetResourceName }}.project_id
  name                     = "{{ .GetResourceName }}"{{ with .GetStringFromObject "spec" "location" }}
  location                 = "{{.}}"{{end}}{{ if .GetInt "spec" "retentionDays" }}
  retention_days           = {{ .GetInt "spec" "retentionDays" }}{{end}}
  log_sink_writer_identity = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
}
{{end}}{{end}}