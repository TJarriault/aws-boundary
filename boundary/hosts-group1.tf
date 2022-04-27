resource "boundary_host_catalog" "group1_servers" {
  name        = "group1_servers"
  description = "Web servers for group1 team"
  type        = "static"
  scope_id    = boundary_scope.core_group1.id
}

resource "boundary_host" "group1_servers" {
  for_each        = var.target_ips_group1
  type            = "static"
  name            = "group1_server_${each.value}"
  description     = "Backend server #${each.value}"
  address         = each.key
  host_catalog_id = boundary_host_catalog.group1_servers.id
}

resource "boundary_host_set" "group1_servers" {
  type            = "static"
  name            = "group1_servers"
  description     = "Host set for group1 servers"
  host_catalog_id = boundary_host_catalog.group1_servers.id
  host_ids        = [for host in boundary_host.group1_servers : host.id]
}
