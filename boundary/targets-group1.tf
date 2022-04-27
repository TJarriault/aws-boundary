resource "boundary_target" "group1_servers_ssh" {
  type                     = "tcp"
  name                     = "groupB1_servers_ssh"
  description              = "GroupB1 SSH target"
  scope_id                 = boundary_scope.core_group1.id
  session_connection_limit = -1
  default_port             = 22
  host_set_ids = [
    boundary_host_set.group1_servers.id
  ]
}

resource "boundary_target" "group1_servers_website" {
  type                     = "tcp"
  name                     = "groupB1_servers_website"
  description              = "GroupB1 website target"
  scope_id                 = boundary_scope.core_group1.id
  session_connection_limit = -1
  default_port             = 80
  host_set_ids = [
    boundary_host_set.group1_servers.id
  ]
}
