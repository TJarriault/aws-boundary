


# Adds an org-level role granting administrative permissions within the core_group1 project
resource "boundary_role" "group2-prod_project_admin" {
  name           = "Group2-prod_admin"
  description    = "Administrator role for Group2 PROD"
  scope_id       = boundary_scope.org2.id
  grant_scope_id = boundary_scope.group_vault.id
  grant_strings = [
    "id=*;type=*;actions=*"
  ]
  principal_ids = concat(
    [for user in boundary_user.group-prod-vault : user.id],
  )
}


