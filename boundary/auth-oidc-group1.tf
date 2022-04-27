
resource "boundary_auth_method_oidc" "devoxx" {
  name                 = "keycloak Devoxx"
  scope_id             = boundary_scope.org.id
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

resource "boundary_managed_group" "app_users_devoxx" {
  name           = "group-app-user-devoxx"
  description    = "App Users"
  auth_method_id = boundary_auth_method_oidc.devoxx.id
  filter = "\"devoxx\" in \"/userinfo/email\""
}

resource "boundary_role" "global_oidc_devoxx" {
  name          = "global_oidc_admin devoxx"
  description   = "global oidc admin devoxx"
  principal_ids = [boundary_managed_group.app_users_devoxx.id]
  grant_strings = ["id=*;type=*;actions=*"]
  scope_id      = boundary_scope.org.id
}

resource "boundary_role" "global_oidc_group1" {
  name           = "global_oidc_devoxx project"
  description    = "global oidc devoxx project"
  principal_ids  = [boundary_managed_group.app_users_devoxx.id]
  grant_strings  = ["id=*;type=*;actions=*"]
  scope_id       = boundary_scope.org.id
  grant_scope_id = boundary_scope.core_group1.id
}

