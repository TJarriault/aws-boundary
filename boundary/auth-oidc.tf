
resource "boundary_auth_method_oidc" "this" {
  name                 = "keycloak-global"
  description          = "boundary keycloak-global"
  scope_id             = "global"
  state                = "active-public"
  is_primary_for_scope = true
  callback_url         = "${var.url}/v1/auth-methods/oidc:authenticate:callback"
  issuer               = var.issuer
  client_id            = var.client_id
  client_secret        = var.client_secret
  allowed_audiences = [
    var.client_id
  ]
  signing_algorithms = ["RS256"]
  api_url_prefix     = var.url
}

resource "boundary_managed_group" "app_users" {
  name           = "app-user-group"
  description    = "App Users"
  auth_method_id = boundary_auth_method_oidc.this.id
  #filter         = "\"${azuread_group.app_users.id}\" in \"/token/groups\""
  filter = "\"orgadmin\" in \"/token/resource_access/boundary.roles\""
}

resource "boundary_role" "global_oidc_admin" {
  name           = "global_oidc_admin"
  description    = "global oidc admin"
  principal_ids  = [boundary_managed_group.app_users.id]
  grant_strings  = ["id=*;type=*;actions=*"]
  scope_id       = boundary_scope.global.id
  grant_scope_id = boundary_scope.global.id
}

resource "boundary_role" "global_oidc_admin_group1" {
  name          = "global_oidc_admin-group1"
  description   = "global oidc admin-group1"
  principal_ids = [boundary_managed_group.app_users.id]
  grant_strings = ["id=*;type=*;actions=*"]
  scope_id      = boundary_scope.core_group1.id
  #grant_scope_id= boundary_scope.org.id
}

#resource "boundary_role" "global_oidc_admin_group2-prod" {
#  name          = "global_oidc_admin-groupé-prod"
#  description   = "global oidc admin-group2-prod"
#  principal_ids = [boundary_managed_group.app_users.id]
#  grant_strings = ["id=*;type=*;actions=*"]
#  scope_id      = boundary_scope.core_group2-prod.id
#  #grant_scope_id= boundary_scope.org.id
#}

