resource "boundary_scope" "global" {
  global_scope           = true
  name                   = "global"
  scope_id               = "global"
  auto_create_admin_role = true
}

resource "boundary_scope" "org" {
  scope_id               = boundary_scope.global.id
  name                   = "Devoxx-boundary"
  description            = "Devoxx Boundary Organization scope"
  auto_create_admin_role = true
}

// create a project for core infrastructure
resource "boundary_scope" "core_infra" {
  name                     = "core_infra"
  description              = "Backend infrastrcture project"
  scope_id                 = boundary_scope.org.id
  auto_create_admin_role   = true
  auto_create_default_role = true
}


// create a project for development project
resource "boundary_scope" "core_group1" {
  name                     = "development"
  description              = "Development project environment"
  scope_id                 = boundary_scope.org.id
  auto_create_admin_role   = true
  auto_create_default_role = true
}

