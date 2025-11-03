resource "google_folder" "test" {
  display_name = "Test Display"
  parent       = "organizations/${var.org_id}"
}
