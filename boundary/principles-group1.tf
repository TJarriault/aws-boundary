resource "boundary_user" "group1" {
  for_each    = var.group1_team
  name        = each.key
  description = "Group1 user: ${each.key}"
  account_ids = [boundary_account.group1_user_acct[each.value].id]
  scope_id    = boundary_scope.org.id
}

resource "boundary_account" "group1_user_acct" {
  for_each       = var.group1_team
  name           = each.key
  description    = "User account for ${each.key}"
  type           = "password"
  login_name     = lower(each.key)
  password       = var.pwd
  auth_method_id = boundary_auth_method.password.id
}

// project level group for group1 management
resource "boundary_group" "group1_core_infra" {
  name        = "group1"
  description = "Group1 team group"
  member_ids  = [for user in boundary_user.group1 : user.id]
  scope_id    = boundary_scope.core_group1.id
}
