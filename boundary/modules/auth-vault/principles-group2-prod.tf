resource "boundary_user" "group-prod-vault" {
  for_each    = var.group2_team
  name        = each.key
  description = "Group user: ${each.key}"
  account_ids = [boundary_account.group-prod-vault_user_acct[each.value].id]
  scope_id    = boundary_scope.org2.id
}

resource "boundary_account" "group-prod-vault_user_acct" {
  for_each       = var.group2_team
  name           = each.key
  description    = "PROD User account for ${each.key}"
  type           = "password"
  login_name     = lower(each.key)
  password       = "foo2022foo"
  #password       = var.pwd
  auth_method_id = boundary_auth_method.password.id
}

// project level group for group2 management
resource "boundary_group" "group_vault" {
  name        = "group-prod"
  description = "PROD team group"
  member_ids  = [for user in boundary_user.group-prod-vault : user.id]
  scope_id    = boundary_scope.group_vault.id
}
