# VPC and Subnets
module "vpc-shared-dev" {
    source  = "terraform-google-modules/network/google"
    version = "~> 5.0"

    project_id   = module.prj-network1.project_id
    network_name = "vpc-shared-dev"
    routing_mode = "GLOBAL"
    description  = "vpc-shared-dev VPC"

    subnets = [
       
        {
            subnet_name           = "sb-dev-shared-base-us-central1"
            subnet_ip             = "10.0.64.0/21"
            subnet_region         = "us-central1"
            subnet_private_access = true
            subnet_flow_logs      = true
            subnet_flow_logs_sampling = "0.5"
            subnet_flow_logs_metadata = "INCLUDE_ALL_METADATA"
            subnet_flow_logs_interval = "INTERVAL_10_MIN"
        },
        {
            subnet_name           = "sb-dev-shared-base-us-west1"
            subnet_ip             = "10.1.64.0/21"
            subnet_region         = "us-west1"
            subnet_private_access = true
            subnet_flow_logs      = true
            subnet_flow_logs_sampling = "0.5"
            subnet_flow_logs_metadata = "INCLUDE_ALL_METADATA"
            subnet_flow_logs_interval = "INTERVAL_10_MIN"
        },
    ]
    
}
# Firewall Rules
resource "google_compute_firewall" "vpc-shared-dev-allow-iap-rdp" {
  name      = "vpc-shared-dev-allow-iap-rdp"
  network   = module.vpc-shared-dev.network_name
  project   = module.prj-network1.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "tcp"
    ports    = ["3389",]
  }

  source_ranges = [
  "35.235.240.0/20",
  ]
}
resource "google_compute_firewall" "vpc-shared-dev-allow-iap-ssh" {
  name      = "vpc-shared-dev-allow-iap-ssh"
  network   = module.vpc-shared-dev.network_name
  project   = module.prj-network1.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "tcp"
    ports    = ["22",]
  }

  source_ranges = [
  "35.235.240.0/20",
  ]
}
resource "google_compute_firewall" "vpc-shared-dev-allow-icmp" {
  name      = "vpc-shared-dev-allow-icmp"
  network   = module.vpc-shared-dev.network_name
  project   = module.prj-network1.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "icmp"
  }

  source_ranges = [
  "0.0.0.0/0",
  ]
}
# NAT Router and config
resource "google_compute_router" "cr-vpc-shared-dev-central1-router" {
  name    = "cr-vpc-shared-dev-central1-router"
  project = module.prj-network1.project_id
  region  = "us-central1"
  network = module.vpc-shared-dev.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-dev-central1-egress" {
  name                               = "rn-vpc-shared-dev-central1-egress"
  project                            = module.prj-network1.project_id
  router                             = google_compute_router.cr-vpc-shared-dev-central1-router.name
  region                             = "us-central1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-dev-central1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-dev-central1-1" {
  project = module.prj-network1.project_id
  name    = "ca-vpc-shared-dev-central1-1"
  region  = "us-central1"
}
resource "google_compute_router" "cr-vpc-shared-dev-west1-router" {
  name    = "cr-vpc-shared-dev-west1-router"
  project = module.prj-network1.project_id
  region  = "us-west1"
  network = module.vpc-shared-dev.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-dev-west1-egress" {
  name                               = "rn-vpc-shared-dev-west1-egress"
  project                            = module.prj-network1.project_id
  router                             = google_compute_router.cr-vpc-shared-dev-west1-router.name
  region                             = "us-west1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-dev-west1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-dev-west1-1" {
  project = module.prj-network1.project_id
  name    = "ca-vpc-shared-dev-west1-1"
  region  = "us-west1"
}

# Service Networking for Cloud SQL & other services
resource "google_service_networking_connection" "svc-net-vpc-shared-dev-dev" {
  network                 = module.vpc-shared-dev.network_self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.ga-vpc-shared-dev-dev-vpc-peering-internal.name]
}

resource "google_compute_global_address" "ga-vpc-shared-dev-dev-vpc-peering-internal" {
  name          = "ga-vpc-shared-dev-dev-vpc-peering-internal"
  project       = module.prj-network1.project_id
  purpose       = "VPC_PEERING" 
  address_type  = "INTERNAL"
  address       = "10.16.64.0"
  prefix_length = "21"
  network       = module.vpc-shared-dev.network_self_link
}

# VPC and Subnets
module "vpc-shared-prod" {
    source  = "terraform-google-modules/network/google"
    version = "~> 5.0"

    project_id   = module.prj-network2.project_id
    network_name = "vpc-shared-prod"
    routing_mode = "GLOBAL"
    description  = "vpc-shared-prod VPC"

    subnets = [
       
        {
            subnet_name           = "sb-prod-shared-base-us-central1"
            subnet_ip             = "10.0.64.0/21"
            subnet_region         = "us-central1"
            subnet_private_access = true
            subnet_flow_logs      = true
            subnet_flow_logs_sampling = "0.5"
            subnet_flow_logs_metadata = "INCLUDE_ALL_METADATA"
            subnet_flow_logs_interval = "INTERVAL_10_MIN"
        },
        {
            subnet_name           = "sb-prod-shared-base-us-west1"
            subnet_ip             = "10.1.64.0/21"
            subnet_region         = "us-west1"
            subnet_private_access = true
            subnet_flow_logs      = true
            subnet_flow_logs_sampling = "0.5"
            subnet_flow_logs_metadata = "INCLUDE_ALL_METADATA"
            subnet_flow_logs_interval = "INTERVAL_10_MIN"
        },
    ]
    
}
# Firewall Rules
resource "google_compute_firewall" "vpc-shared-prod-allow-iap-rdp" {
  name      = "vpc-shared-prod-allow-iap-rdp"
  network   = module.vpc-shared-prod.network_name
  project   = module.prj-network2.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "tcp"
    ports    = ["3389",]
  }

  source_ranges = [
  "35.235.240.0/20",
  ]
}
resource "google_compute_firewall" "vpc-shared-prod-allow-iap-ssh" {
  name      = "vpc-shared-prod-allow-iap-ssh"
  network   = module.vpc-shared-prod.network_name
  project   = module.prj-network2.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "tcp"
    ports    = ["22",]
  }

  source_ranges = [
  "35.235.240.0/20",
  ]
}
resource "google_compute_firewall" "vpc-shared-prod-allow-icmp" {
  name      = "vpc-shared-prod-allow-icmp"
  network   = module.vpc-shared-prod.network_name
  project   = module.prj-network2.project_id
  direction = "INGRESS"
  priority  = 10000

  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }

  allow {
    protocol = "icmp"
  }

  source_ranges = [
  "0.0.0.0/0",
  ]
}
# NAT Router and config
resource "google_compute_router" "cr-vpc-shared-prod-central1-router" {
  name    = "cr-vpc-shared-prod-central1-router"
  project = module.prj-network2.project_id
  region  = "us-central1"
  network = module.vpc-shared-prod.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-prod-central1-egress" {
  name                               = "rn-vpc-shared-prod-central1-egress"
  project                            = module.prj-network2.project_id
  router                             = google_compute_router.cr-vpc-shared-prod-central1-router.name
  region                             = "us-central1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-prod-central1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-prod-central1-1" {
  project = module.prj-network2.project_id
  name    = "ca-vpc-shared-prod-central1-1"
  region  = "us-central1"
}
resource "google_compute_router" "cr-vpc-shared-prod-west1-router" {
  name    = "cr-vpc-shared-prod-west1-router"
  project = module.prj-network2.project_id
  region  = "us-west1"
  network = module.vpc-shared-prod.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-prod-west1-egress" {
  name                               = "rn-vpc-shared-prod-west1-egress"
  project                            = module.prj-network2.project_id
  router                             = google_compute_router.cr-vpc-shared-prod-west1-router.name
  region                             = "us-west1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-prod-west1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-prod-west1-1" {
  project = module.prj-network2.project_id
  name    = "ca-vpc-shared-prod-west1-1"
  region  = "us-west1"
}
