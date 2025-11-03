resource "google_folder" "authoritative" {
  display_name = "authoritative"
  parent       = "organizations/${var.org_id}"
}

resource "google_folder" "test" {
  display_name = "Test Display"
  parent       = "organizations/${var.org_id}"
}
