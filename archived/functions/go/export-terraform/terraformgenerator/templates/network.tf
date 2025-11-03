{{range $vpc := .ComputeNetwork}}{{ if $vpc.ShouldCreate }}
# VPC and Subnets
module "{{ $vpc.GetResourceName }}" {
    source  = "terraform-google-modules/network/google"
    version = "~> 5.0"

    project_id   = module.{{ .Parent.GetResourceName }}.project_id
    network_name = "{{ $vpc.GetResourceName }}"{{ with .GetStringFromObject "spec" "routingMode" }}
    routing_mode = "{{ . }}"{{end}}{{ with .GetStringFromObject "spec" "description" }}
    description  = "{{ . }}"{{end}}

    subnets = [
       {{range $subnet := $vpc.GetChildrenByKind "ComputeSubnetwork" }}
        {
            subnet_name           = "{{ $subnet.GetResourceName }}"
            subnet_ip             = "{{ $subnet.GetStringFromObject "spec" "ipCidrRange" }}"
            subnet_region         = "{{ $subnet.GetStringFromObject "spec" "region" }}"{{ with $subnet.GetBool "spec" "privateIpGoogleAccess" }}
            subnet_private_access = {{ . }}{{end}}{{ if $subnet.GetStringFromObject "spec" "logConfig" "aggregationInterval" }}
            subnet_flow_logs      = true{{ with $subnet.GetFloat "spec" "logConfig" "flowSampling" }}
            subnet_flow_logs_sampling = "{{ . }}"{{end}}{{ with $subnet.GetStringFromObject "spec" "logConfig" "metadata" }}
            subnet_flow_logs_metadata = "{{ . }}"{{end}}
            subnet_flow_logs_interval = "{{ $subnet.GetStringFromObject "spec" "logConfig" "aggregationInterval" }}"{{end}}
        },{{end}}
    ]
    {{if $vpc.GetChildrenByKind "ComputeRoute"}}
    routes = [
      {{range $route := $vpc.GetChildrenByKind "ComputeRoute" }}{
        name = "{{ $route.GetResourceName }}"{{ with $route.GetStringFromObject "spec" "description" }}
        description = "{{ . }}"{{end}}{{ with $route.GetStringFromObject "spec" "destRange" }}
        destination_range = "{{ . }}"{{end}}{{ with $route.GetInt "spec" "priority" }}
        priority = "{{ . }}"{{end}}{{ with $route.GetStringFromObject "spec" "nextHopGateway" }}{{if eq . "default-internet-gateway"}}
        next_hop_internet = "true"{{end}}{{end}}{{ with $route.GetStringsFromObject "spec" "tags" }}
        tags = "{{ . | strSliceToCommaSepStr }}"{{end}}
      },{{end}}
    ]{{end}}
}
# Firewall Rules{{range $fw := $vpc.GetChildrenByKind "ComputeFirewall" }}
resource "google_compute_firewall" "{{ $fw.GetResourceName }}" {
  name      = "{{ $fw.GetResourceName }}"
  network   = module.{{ $vpc.GetResourceName }}.network_name
  project   = module.{{ $vpc.Parent.GetResourceName }}.project_id{{ with $fw.GetStringFromObject "spec" "direction" }}
  direction = "{{ . }}"{{end}}{{ with $fw.GetInt "spec" "priority" }}
  priority  = {{.}}{{end}}
{{ if $fw.GetBool "spec" "enableLogging" }}
  log_config {
      metadata = "INCLUDE_ALL_METADATA"
    }
{{end}}{{ range $fwAllow := $fw.GetFirewallAllowPortsProtocol }}
  allow {
    protocol = "{{ .Protocol }}"{{ if .Ports }}
    ports    = [{{ range .Ports }}"{{ . }}",{{end}}]{{end}}
  }
{{end}}{{ if $fw.GetStringsFromObject "spec" "sourceRanges" }}
  source_ranges = [{{ range $fw.GetStringsFromObject "spec" "sourceRanges" }}
  "{{ . }}",{{end}}
  ]{{end}}
}{{end}}
# NAT Router and config{{range $router := $vpc.GetChildrenByKind "ComputeRouter" }}
resource "google_compute_router" "{{ $router.GetResourceName }}" {
  name    = "{{ $router.GetResourceName }}"
  project = module.{{ $vpc.Parent.GetResourceName }}.project_id
  region  = "{{ $router.GetStringFromObject "spec" "region" }}"
  network = module.{{ $vpc.GetResourceName }}.network_self_link
}
{{range $routerNat := $router.GetChildrenByKind "ComputeRouterNAT" }}
resource "google_compute_router_nat" "{{ $routerNat.GetResourceName }}" {
  name                               = "{{ $routerNat.GetResourceName }}"
  project                            = module.{{ $vpc.Parent.GetResourceName }}.project_id
  router                             = google_compute_router.{{ $router.GetResourceName }}.name
  region                             = "{{ $routerNat.GetStringFromObject "spec" "region" }}" {{ with $routerNat.GetStringFromObject "spec" "natIpAllocateOption" }}
  nat_ip_allocate_option             = "{{ . }}"{{end}}
  nat_ips                            = google_compute_address.{{ $routerNat.References.ComputeAddress.GetResourceName }}.*.self_link {{ with $routerNat.GetStringFromObject "spec" "sourceSubnetworkIpRangesToNat" }}
  source_subnetwork_ip_ranges_to_nat = "{{ . }}"{{end}}
{{ if $routerNat.GetBool "spec" "logConfig" "enable" }}
  log_config { {{ with $routerNat.GetStringFromObject "spec" "logConfig" "filter" }}
    filter = "{{ . }}" {{ end }}
    enable = true
  }{{ end }}
}
{{with $routerNat.References.ComputeAddress }}
resource "google_compute_address" "{{ .GetResourceName }}" {
  project = module.{{ $vpc.Parent.GetResourceName }}.project_id
  name    = "{{ .GetResourceName }}"
  region  = "{{ .GetStringFromObject "spec" "location" }}"
}{{end}}{{end}}{{end}}
{{range $svcNet := $vpc.GetChildrenByKind "ServiceNetworkingConnection" }}
# Service Networking for Cloud SQL & other services
resource "google_service_networking_connection" "{{ $svcNet.GetResourceName }}" {
  network                 = module.{{ $vpc.GetResourceName }}.network_self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.{{ $svcNet.References.ComputeAddress.GetResourceName }}.name]
}
{{with $svcNet.References.ComputeAddress }}
resource "google_compute_global_address" "{{ .GetResourceName }}" {
  name          = "{{ .GetResourceName }}"
  project       = module.{{ $vpc.Parent.GetResourceName }}.project_id{{ with .GetStringFromObject "spec" "purpose" }}
  purpose       = "{{ . }}" {{end}}{{ with .GetStringFromObject "spec" "addressType" }}
  address_type  = "{{ . }}"{{end}}{{ with .GetStringFromObject "spec" "address" }}
  address       = "{{ . }}"{{ end }}{{ with .GetInt "spec" "prefixLength" }}
  prefix_length = "{{ . }}"{{ end }}
  network       = module.{{ $vpc.GetResourceName }}.network_self_link
}
{{end}}{{end}}{{end}}{{end}}