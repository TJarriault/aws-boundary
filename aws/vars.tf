resource "random_pet" "test" {
  length = 1
}

locals {
  tags = {
    Name = "${var.tag}-${random_pet.test.id}"
  }

  pub_cidrs  = cidrsubnets("10.0.0.0/24", 2, 2, 2, 2)
  priv_cidrs = cidrsubnets("10.0.100.0/24", 2, 2, 2, 2)
}

variable "tag" {
  default = "boundary-t"
}

variable "boundary_bin" {
  #default = "~/projects/boundary/bin"
  default = "/appli/sogeti/hashicorp/boundary-aws/aws/source"
}

variable "pub_ssh_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "priv_ssh_key_path" {
  default = ""
}

variable "num_workers" {
  default = 1
}

variable "num_controllers" {
  default = 2
}

variable "num_targets" {
  default = 1
}

variable "num_subnets_public" {
  default = 3
}

variable "num_subnets_private" {
  default = 3
}

variable "tls_cert_path" {
  default = "/etc/pki/tls/boundary/boundary.cert"
}

variable "tls_key_path" {
  default = "/etc/pki/tls/boundary/boundary.key"
}

variable "tls_disabled" {
  default = true
}

variable "kms_type" {
  default = "aws"
}
