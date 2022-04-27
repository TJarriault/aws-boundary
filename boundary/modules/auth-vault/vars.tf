variable "url" {
  type = string
}

variable "scope_global" {
  type = string
}

variable "group2_team" {
  type = set(string)
  default = [
    "olivier",
    "sandra",
  ]
}


variable "target_ips" {
  type = set(string)
  default = [
    "10.0.100.29",
  ]
}

variable "target_ips_group1" {
  type = set(string)
  default = [
    "10.0.100.41",
  ]
}


variable "issuer" {
  type = string
}


variable "pwd" {
  type = string
}

variable "token" {
  type = string
}


variable "vault_url" {
  type = string
}

variable "vault_token" {
  type = string
}
