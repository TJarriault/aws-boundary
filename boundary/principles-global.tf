resource "boundary_role" "global_admin" {
  name        = "global_admin"
  description = "Global admin"
  for_each       = var.global_team
  scope_id       = "global"
  grant_scope_id = boundary_scope.global.id
  grant_strings = [
    "id=*;type=*;actions=*"
  ]
  principal_ids = [for user in boundary_user.global : user.id]
}



resource "boundary_user" "global" {
  for_each    = var.global_team
  name        = each.key
  description = "Global admin user: ${each.key}"
  account_ids = [boundary_account.global_user_acct[each.value].id]
  scope_id    = boundary_scope.global.id
}

resource "boundary_account" "global_user_acct" {
  for_each       = var.global_team
  name           = each.key
  description    = "User account for ${each.key}"
  type           = "password"
  login_name     = lower(each.key)
  password       = var.pwd
  auth_method_id = boundary_auth_method.global-password.id
}

// organiation level group for the leadership team
resource "boundary_group" "global" {
  name        = "global_team"
  description = "Organization group for global team"
  member_ids  = [for user in boundary_user.global : user.id]
  scope_id    = boundary_scope.global.id
}

