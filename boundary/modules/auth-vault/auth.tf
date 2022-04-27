resource "boundary_auth_method" "password" {
  name        = "Org2_password_auth_method"
  description = "Password auth method for org2"
  type        = "password"
  scope_id    = boundary_scope.org2.id
}


resource "boundary_credential_store_vault" "vault" {
  name        = "cred-vault-store-2"
  description = "Vault credential store!"
  address     = var.vault_url
  token       = var.vault_token
  scope_id    = boundary_scope.group_vault.id
}

resource "boundary_credential_library_vault" "this" {
  name                = "vault-cred-library"
  description         = "Vault credential library!"
  credential_store_id = boundary_credential_store_vault.vault.id
  path                = "secret/boundary-org2" # change to Vault backend path
  http_method         = "GET"
}

resource "boundary_credential_library_vault" "bar" {
  name                = "bar"
  description         = "My second Vault credential library!"
  credential_store_id = boundary_credential_store_vault.vault.id
  path                = "secret/back2" # change to Vault backend path
  http_method         = "POST"
  http_request_body   = <<EOT
{
  "key": "Value",
}
EOT
}
