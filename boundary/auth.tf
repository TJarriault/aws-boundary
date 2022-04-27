resource "boundary_auth_method" "password" {
  name        = "corp_password_auth_method"
  description = "Password auth method for Corp org"
  type        = "password"
  scope_id    = boundary_scope.org.id
}

resource "boundary_auth_method" "global-password" {
  name        = "global_auth_method"
  description = "Global password auth method"
  type        = "password"
  scope_id    = boundary_scope.global.id
}

