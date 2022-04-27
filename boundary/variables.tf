variable "url" {
  default = "http://boundary-t-controller-glider-f4c794b6e55bbe02.elb.us-east-1.amazonaws.com:9200"
}

variable "backend_team" {
  type = set(string)
  default = [
    "jim",
    "mike",
  ]
}

variable "frontend_team" {
  type = set(string)
  default = [
    "randy",
    "susmitha",
  ]
}

variable "group1_team" {
  type = set(string)
  default = [
    "tony",
    "paul",
  ]
}


variable "group2_team" {
  type = set(string)
  default = [
    "olivier",
    "sandra",
  ]
}


variable "leadership_team" {
  type = set(string)
  default = [
    "jeff",
    "globaluser",
  ]
}


variable "global_team" {
  type = set(string)
  default = [
    "admin2",
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


variable "target_ips_group2-prod" {
  type = set(string)
  default = [
    "10.0.100.48",
  ]
}


variable "kms_recovery_key_id" {
  default = ""
  sensitive   = true
}

variable "issuer" {
  default = ""
  sensitive   = true
}


variable "client_id" {
  default = "boundary"
}


variable "client_secret" {
  default = ""
  sensitive   = true
}

variable "pwd" {
  default = ""
  sensitive   = true
}

variable "token" {
  default = ""
  sensitive   = true
}

variable "vault_url" {
  default = ""
  sensitive   = true
}
variable "vault_token" {
  default = ""
  sensitive   = true
}
