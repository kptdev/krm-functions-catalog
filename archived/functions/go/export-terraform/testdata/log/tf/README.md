# Google Cloud Foundation Blueprint

This directory contains Terraform configuration for a foundational environment on Google Cloud.

It includes a subset of resources configured via the [setup checklist](https://cloud.google.com/docs/enterprise/setup-checklist)
and is based on the [security foundations blueprint](https://cloud.google.com/architecture/security-foundations).

## Prerequisites

To run the commands described in this document, you need the following:

1. Install the [Google Cloud SDK](https://cloud.google.com/sdk/install) version 319.0.0 or later
1. Install [Terraform](https://www.terraform.io/downloads.html) version 0.13.7 or later.
1. Set up a Google Cloud
   [organization](https://cloud.google.com/resource-manager/docs/creating-managing-organization).
1. Set up a Google Cloud
   [billing account](https://cloud.google.com/billing/docs/how-to/manage-billing-account).
1. For the user who will run the Terraform install, grant the
   following roles:
   -  The `roles/billing.admin` role on the billing account.
   -  The `roles/resourcemanager.organizationAdmin` role on the Google
      Cloud organization.
   -  The `roles/resourcemanager.folderCreator` role on the Google
      Cloud organization.
   -  The `roles/resourcemanager.projectCreator` role on the Google
      Cloud organization.

## Deploying

1. Run `terraform init`.
1. Run `terraform plan` and review the output.
1. Run `terraform apply`.

## Next steps

Once you have the basic foundation deployed, you should explore:
1. Building an [advanced foundation](https://github.com/terraform-google-modules/terraform-example-foundation) using the security blueprint
2. Automatically deploying Terraform with [Cloud Build](https://cloud.google.com/architecture/managing-infrastructure-as-code)
