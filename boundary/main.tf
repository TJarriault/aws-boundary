provider "boundary" {
  addr             = var.url
  recovery_kms_hcl = <<EOT
kms "awskms" {
	purpose    = "recovery"
	key_id     = "global_root"
  kms_key_id = "${var.kms_recovery_key_id}"
}
EOT
}

#module "auth-vault" {
#  source = "./modules/auth-vault"

#  scope_global = boundary_scope.global.id
#  pwd = var.pwd
#  url = var.url
#  issuer = var.issuer
#  group2_team = var.group2_team
#  token = var.token
#  vault_url = var.vault_url
#  vault_token = var.vault_token

#}
