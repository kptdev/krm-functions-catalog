module "organization-iam" {
  source  = "terraform-google-modules/iam/google//modules/organizations_iam"
  version = "~> 7.4"

  organizations = ["11111111111"]

  bindings = {
    
    "roles/editor" = [
      "group:gcp-organization-admins@example.com",
      "group:gcp-developers@example.com",
    ]
    
    "roles/orgpolicy.policyAdmin" = [
      "group:gcp-organization-admins@example.com",
    ]
    
  }
}


module "folder-1-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = ["folders/335620346181"]

  bindings = {
    
    "roles/viewer" = [
      "group:gcp-developers@example.com",
    ]
    
  }
}


module "authoritative-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = [google_folder.authoritative.name]

  bindings = {
    
    "roles/viewer" = [
      "group:gcp-another@example.com",
      "group:gcp-test@example.com",
    ]
    
  }
}


module "test-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = [google_folder.test.name]

  bindings = {
    
    "roles/editor" = [
      "group:gcp-developers@example.com",
    ]
    
    "roles/viewer" = [
      "group:gcp-developers@example.com",
      "group:gcp-viewers@example.com",
      "group:gcp-devops@example.com",
    ]
    
  }
}


resource "google_organization_iam_audit_config" "org_config" {
  org_id  = var.org_id
  service = "allServices"

  audit_log_config {
      log_type = "ADMIN_READ"
  }
}
