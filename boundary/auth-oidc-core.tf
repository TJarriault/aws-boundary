resource "boundary_managed_group" "app_users_core" {
  name           = "group-app-user-core-infra"
  description    = "App Users core infra"
  auth_method_id = boundary_auth_method_oidc.devoxx.id
  filter = "\"jim.fr\" in \"/userinfo/email\""
}

resource "boundary_role" "global_oidc_core" {
  name          = "global_oidc_admin core infra"
  description   = "global oidc admin core infra"
  principal_ids = [boundary_managed_group.app_users_core.id]
  grant_strings = ["id=*;type=*;actions=*"]
  scope_id      = boundary_scope.org.id
}

resource "boundary_role" "global_oidc_group-core" {
  name           = "global_oidc_admin group1 orga level"
  description    = "global oidc admin group1 orga level"
  principal_ids  = [boundary_managed_group.app_users_core.id]
  grant_strings  = ["id=*;type=*;actions=*"]
  scope_id       = boundary_scope.core_infra.id
  #grant_scope_id = boundary_scope.core_infra.id
}

