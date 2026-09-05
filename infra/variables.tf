variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project_name" {
  type    = string
  default = "postech-tc3"
}

variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "database_url" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "jwt_ttl" {
  type    = string
  default = "15m"
}

variable "issuer_artifact_path" {
  type    = string
  default = "../issuer.zip"
}

variable "authorizer_artifact_path" {
  type    = string
  default = "../authorizer.zip"
}

variable "log_retention_days" {
  type    = number
  default = 14
}

variable "lab_role_name" {
  type    = string
  default = "LabRole"
}
