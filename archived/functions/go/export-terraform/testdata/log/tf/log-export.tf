module "logsink-123456789012-bqsink" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.bqlogexportdataset-destination.destination_uri
  log_sink_name        = "123456789012-bqsink"
  parent_resource_id   = var.org_id
  parent_resource_type = "organization"
  include_children     = true
}

module "bqlogexportdataset-destination" {
  source  = "terraform-google-modules/log-export/google//modules/bigquery"
  version = "~> 7.3.0"

  project_id               = module.prj-logging.project_id
  dataset_name             = "bqlogexportdataset"
  log_sink_writer_identity = module.logsink-123456789012-bqsink.writer_identity
  expiration_days          = "365"
  location                 = "US"
}

module "logsink-123456789012-orglogbucketsink" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.my-log-k8s-bucket-destination.destination_uri
  log_sink_name        = "123456789012-orglogbucketsink"
  parent_resource_id   = var.org_id
  parent_resource_type = "organization"
  include_children     = true
}

module "my-log-k8s-bucket-destination" {
  source  = "terraform-google-modules/log-export/google//modules/logbucket"
  version = "~> 7.4.1"

  project_id               = module.prj-logging.project_id
  name                     = "my-log-k8s-bucket"
  location                 = "global"
  retention_days           = 30
  log_sink_writer_identity = module.logsink-123456789012-orglogbucketsink.writer_identity
}

module "logsink-123456789012-pubsubsink" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.pubsub-logexport-dataset-destination.destination_uri
  log_sink_name        = "123456789012-pubsubsink"
  parent_resource_id   = var.org_id
  parent_resource_type = "organization"
  include_children     = true
}

module "pubsub-logexport-dataset-destination" {
  source  = "terraform-google-modules/log-export/google//modules/pubsub"
  version = "~> 7.3.0"

  project_id               = module.prj-logging.project_id
  topic_name               = "pubsub-logexport-dataset"
  log_sink_writer_identity = module.logsink-123456789012-pubsubsink.writer_identity
}

module "logsink-123456789012-storagesink" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.my-storage-bucket-destination.destination_uri
  log_sink_name        = "123456789012-storagesink"
  parent_resource_id   = var.org_id
  parent_resource_type = "organization"
  include_children     = true
}

module "my-storage-bucket-destination" {
  source  = "terraform-google-modules/log-export/google//modules/storage"
  version = "~> 7.3.0"

  project_id                  = module.prj-logging.project_id
  storage_bucket_name         = "my-storage-bucket"
  log_sink_writer_identity    = module.logsink-123456789012-storagesink.writer_identity
  uniform_bucket_level_access = false
  location                    = "US"
  storage_class               = "MULTI_REGIONAL"
  retention_policy = {
    retention_period_days = 365,
    is_locked             = false
  }
}
