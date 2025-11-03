module "project-in-external" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-in-external"
  org_id     = var.org_id
  folder_id  = "335620346181"

  billing_account = var.billing_account
}

module "project-in-folder" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-name"
  project_id = "project-in-folder"
  org_id     = var.org_id
  folder_id  = google_folder.test.name

  billing_account = var.billing_account
}

module "project-in-org" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-in-org"
  org_id     = var.org_id

  billing_account = var.billing_account
}
