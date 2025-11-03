# VPC and Subnets
module "vpc-shared-base" {
    source  = "terraform-google-modules/network/google"
    version = "~> 5.0"

    project_id   = module.prj-network.project_id
    network_name = "vpc-shared-base"
    routing_mode = "GLOBAL"
    description  = "vpc-shared-base VPC"

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
    
    routes = [
      {
        name = "rt-vpc-shared-base-1000-all-default-private-api"
        description = "Route through IGW to allow private google api access."
        destination_range = "199.36.153.8/30"
        priority = "1000"
        next_hop_internet = "true"
      },{
        name = "rt-vpc-shared-base-1000-egress-internet-default"
        description = "Tag based route through IGW to access internet"
        destination_range = "0.0.0.0/0"
        priority = "1000"
        next_hop_internet = "true"
        tags = "egress-internet"
      },
    ]
}
# Firewall Rules
resource "google_compute_firewall" "vpc-shared-base-allow-iap-rdp" {
  name      = "vpc-shared-base-allow-iap-rdp"
  network   = module.vpc-shared-base.network_name
  project   = module.prj-network.project_id
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
resource "google_compute_firewall" "vpc-shared-base-allow-iap-ssh" {
  name      = "vpc-shared-base-allow-iap-ssh"
  network   = module.vpc-shared-base.network_name
  project   = module.prj-network.project_id
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
resource "google_compute_firewall" "vpc-shared-base-allow-icmp" {
  name      = "vpc-shared-base-allow-icmp"
  network   = module.vpc-shared-base.network_name
  project   = module.prj-network.project_id
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
resource "google_compute_router" "cr-vpc-shared-base-central1-router" {
  name    = "cr-vpc-shared-base-central1-router"
  project = module.prj-network.project_id
  region  = "us-central1"
  network = module.vpc-shared-base.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-base-central1-egress" {
  name                               = "rn-vpc-shared-base-central1-egress"
  project                            = module.prj-network.project_id
  router                             = google_compute_router.cr-vpc-shared-base-central1-router.name
  region                             = "us-central1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-base-central1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-base-central1-1" {
  project = module.prj-network.project_id
  name    = "ca-vpc-shared-base-central1-1"
  region  = "us-central1"
}
resource "google_compute_router" "cr-vpc-shared-base-west1-router" {
  name    = "cr-vpc-shared-base-west1-router"
  project = module.prj-network.project_id
  region  = "us-west1"
  network = module.vpc-shared-base.network_self_link
}

resource "google_compute_router_nat" "rn-vpc-shared-base-west1-egress" {
  name                               = "rn-vpc-shared-base-west1-egress"
  project                            = module.prj-network.project_id
  router                             = google_compute_router.cr-vpc-shared-base-west1-router.name
  region                             = "us-west1" 
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = google_compute_address.ca-vpc-shared-base-west1-1.*.self_link 
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config { 
    filter = "TRANSLATIONS_ONLY" 
    enable = true
  }
}

resource "google_compute_address" "ca-vpc-shared-base-west1-1" {
  project = module.prj-network.project_id
  name    = "ca-vpc-shared-base-west1-1"
  region  = "us-west1"
}

# Service Networking for Cloud SQL & other services
resource "google_service_networking_connection" "svc-net-vpc-shared-base-dev" {
  network                 = module.vpc-shared-base.network_self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.ga-vpc-shared-base-dev-vpc-peering-internal.name]
}

resource "google_compute_global_address" "ga-vpc-shared-base-dev-vpc-peering-internal" {
  name          = "ga-vpc-shared-base-dev-vpc-peering-internal"
  project       = module.prj-network.project_id
  purpose       = "VPC_PEERING" 
  address_type  = "INTERNAL"
  address       = "10.16.64.0"
  prefix_length = "21"
  network       = module.vpc-shared-base.network_self_link
}
