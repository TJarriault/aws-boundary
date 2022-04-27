
resource "boundary_scope" "org2" {
  scope_id               = var.scope_global
  name                   = "Devoxx-boundary-vault"
  description            = "Devoxx Boundary with Vault Organization scope"
  auto_create_admin_role = true
}

// create a project for group1 infrastructure
resource "boundary_scope" "group_vault" {
  name                     = "prod-with-vault"
  description              = "prod project environment"
  scope_id                 = boundary_scope.org2.id
  auto_create_admin_role   = true
  auto_create_default_role = true
}

